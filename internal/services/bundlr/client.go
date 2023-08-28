package bundlr

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/warp-contracts/syncer/src/utils/arweave"
	"github.com/warp-contracts/syncer/src/utils/bundlr"
)

type Client struct {
	Signer      *bundlr.EthereumSigner
	url         string
	contentType string
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
	}, nil
}

// PrepareData prepares data for uploading to bundlr by compressing and encrypting input.
func (c *Client) PrepareData(data []byte, userDeviceID string, start, end time.Time) (bundlr.BundleItem, string, error) {
	fileName := fmt.Sprintf("%s-%d-%d.zip", userDeviceID, start.Unix(), end.Unix())
	compressedData, err := c.compress(data, fileName)
	if err != nil {
		return bundlr.BundleItem{}, "", err
	}

	// generating random 32 byte key for AES-256
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return bundlr.BundleItem{}, "", err
	}
	encryptionKey := hex.EncodeToString(bytes)

	encryptedData, err := c.encrypt(compressedData, bytes)
	if err != nil {
		return bundlr.BundleItem{}, encryptionKey, err
	}

	dataItem := bundlr.BundleItem{
		Data: arweave.Base64String(encryptedData),
		Tags: bundlr.Tags{
			bundlr.Tag{Name: "Trip-Type", Value: "segment"},
			bundlr.Tag{Name: "Device-ID", Value: userDeviceID},
			bundlr.Tag{Name: "Start-Time", Value: start.Format(time.RFC3339)},
			bundlr.Tag{Name: "End-Time", Value: end.Format(time.RFC3339)},
		},
	}

	err = dataItem.Sign(c.Signer)

	return dataItem, encryptionKey, err
}

func (c *Client) Upload(dataItem bundlr.BundleItem) (string, error) {
	// TO DO
	return "", nil
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

func (c *Client) encrypt(data, key []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, nonce, data, nil), nil
}
