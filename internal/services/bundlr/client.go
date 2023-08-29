package bundlr

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/DIMO-Network/trips-api/internal/config"
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
func (c *Client) PrepareData(data []byte, encryptionKey []byte, userDeviceID string, start, end time.Time) (bundlr.BundleItem, []byte, error) {
	fileName := fmt.Sprintf("%s-%d-%d.zip", userDeviceID, start.Unix(), end.Unix())
	compressedData, err := c.compress(data, fileName)
	if err != nil {
		return bundlr.BundleItem{}, []byte{}, err
	}

	encryptedData, nonce, err := c.encrypt(compressedData, encryptionKey)
	if err != nil {
		return bundlr.BundleItem{}, []byte{}, err
	}

	dataItem := bundlr.BundleItem{
		Data: arweave.Base64String(encryptedData),
		// in the future-- allow tags to be passed in
		// ie, someone could name their trip?
		Tags: bundlr.Tags{
			bundlr.Tag{Name: "Device-ID", Value: userDeviceID},
			bundlr.Tag{Name: "Start-Time", Value: start.Format(time.RFC3339)},
			bundlr.Tag{Name: "End-Time", Value: end.Format(time.RFC3339)},
			bundlr.Tag{Name: "Nonce", Value: hex.EncodeToString(nonce)},
		},
	}

	return dataItem, nonce, dataItem.Sign(c.Signer)
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

func (c *Client) compress(data []byte, fileName string) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	f, err := w.Create(fileName)
	if err != nil {
		return nil, err
	}

	if _, err := f.Write(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), w.Close()
}

func (c *Client) encrypt(data, key []byte) ([]byte, []byte, error) {
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	return aesgcm.Seal(nil, nonce, data, nil), nonce, nil
}
