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
	"net/http"
	"strconv"
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
	signer, err := bundlr.NewEthereumSigner("0x" + settings.BundlrPrivateKey)
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
func (c *Client) PrepareData(data []byte, encryptionKey []byte, tokenId int, start, end time.Time) (*bundlr.BundleItem, error) {
	fileName := fmt.Sprintf("%d-%d-%d.zip", tokenId, start.Unix(), end.Unix())
	compressedData, err := c.compress(data, fileName)
	if err != nil {
		return nil, err
	}

	encryptedData, nonce, err := c.encrypt(compressedData, encryptionKey)
	if err != nil {
		return nil, err
	}

	dataItem := &bundlr.BundleItem{
		Data: arweave.Base64String(encryptedData),
		Tags: bundlr.Tags{
			bundlr.Tag{Name: "Vehicle-Token-Id", Value: strconv.Itoa(tokenId)},
			bundlr.Tag{Name: "Start-Time", Value: start.Format(time.RFC3339)},
			bundlr.Tag{Name: "End-Time", Value: end.Format(time.RFC3339)},
			bundlr.Tag{Name: "Nonce", Value: hex.EncodeToString(nonce)},
		},
	}

	return dataItem, dataItem.Sign(c.Signer)
}

func (c *Client) Upload(dataItem *bundlr.BundleItem) error {
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

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
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

	err = w.Close()
	return buf.Bytes(), err
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
