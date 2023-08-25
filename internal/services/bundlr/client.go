package bundlr

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/DIMO-Network/trips-api/models"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/warp-contracts/syncer/src/utils/arweave"
	"github.com/warp-contracts/syncer/src/utils/bundlr"
)

type Client struct {
	Signer      *bundlr.EthereumSigner
	url         string
	contentType string
	currency    string
}

func New(settings *config.Settings) (*Client, error) {
	signer, err := bundlr.NewEthereumSigner("0x" + settings.EthereumSignerPrivateKey)
	if err != nil {
		return nil, err
	}

	return &Client{
		Signer:      signer,
		url:         settings.BundlrNetwork,
		contentType: "application/octet-stream",
		currency:    settings.BundlrCurrency,
	}, nil
}

// PrepareData prepares data for uploading to bundlr by compressing and encrypting input.
func (c *Client) PrepareData(data []byte, deviceID, startTime, endTime string) (bundlr.BundleItem, []byte, error) {
	compressedData, err := c.compress(data, startTime, endTime)
	if err != nil {
		return bundlr.BundleItem{}, []byte{}, err
	}

	// generating random 32 byte key for AES-256
	// this will change with PRO-1867 encryption keys created for minted devices
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		return bundlr.BundleItem{}, []byte{}, err
	}

	encryptedData, err := c.encrypt(compressedData, encryptionKey)
	if err != nil {
		return bundlr.BundleItem{}, encryptionKey, err
	}

	dataItem := bundlr.BundleItem{
		Data: arweave.Base64String(encryptedData),
		// in the future-- allow tags to be passed in
		// ie, someone could name their trip?
		Tags: bundlr.Tags{
			bundlr.Tag{Name: "Trip-Type", Value: "segment"},
			bundlr.Tag{Name: "Device-ID", Value: deviceID},
			bundlr.Tag{Name: "Start-Time", Value: startTime},
			bundlr.Tag{Name: "End-Time", Value: endTime},
		},
	}

	err = dataItem.Sign(c.Signer)
	return dataItem, encryptionKey, err
}

func (c *Client) Upload(dataItem bundlr.BundleItem) error {
	reader, err := dataItem.Reader()
	if err != nil {
		return err
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	postBody := bytes.NewBuffer(body)
	fmt.Println("Uploading to this url: ", c.url+"tx/"+c.currency)
	resp, err := http.Post(c.url+"tx/"+c.currency, c.contentType, postBody)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/plain") {
		return errors.New(string(body))
	}

	return nil
}

func (c *Client) Download(ctx context.Context, tripTokenID types.NullDecimal, db *sql.DB) ([]string, error) {

	segment, err := models.Trips(models.TripWhere.TripTokenID.EQ(tripTokenID)).One(ctx, db)
	if err != nil {
		return []string{}, err
	}

	resp, err := http.Get(c.url + segment.BunldrID)
	if err != nil {
		return []string{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}

	decryptedBody, err := c.decrypt(segment.EncryptionKey, body)
	if err != nil {
		return []string{}, err
	}

	return c.decompress(decryptedBody)
}

func (c *Client) compress(data []byte, start, end string) ([]byte, error) {
	b := new(bytes.Buffer)
	zw := zip.NewWriter(b)

	file, err := zw.Create(fmt.Sprintf("%s_%s.json", start, end))
	if err != nil {
		return nil, err
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, err
	}

	err = zw.Close()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (c *Client) decompress(decryptedBody []byte) ([]string, error) {
	unzippedResp := make([]string, 0)

	zipReader, err := zip.NewReader(bytes.NewReader(decryptedBody), int64(len(decryptedBody)))
	if err != nil {
		return unzippedResp, err
	}

	for _, zipFile := range zipReader.File {
		f, err := zipFile.Open()
		if err != nil {
			return unzippedResp, err
		}
		defer f.Close()

		unzippedBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return unzippedResp, err
		}

		unzippedResp = append(unzippedResp, string(unzippedBytes))
	}

	return unzippedResp, nil
}

func (c *Client) encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	b := base64.StdEncoding.EncodeToString(data)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func (c *Client) decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(text) < aes.BlockSize {
		return nil, errors.New("invalid cipher length")
	}

	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	return base64.StdEncoding.DecodeString(string(text))
}
