package bundlr

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"

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
func (c *Client) PrepareData(data []byte, deviceID, startTime, endTime string) (bundlr.BundleItem, string, error) {
	compressedData, err := c.compress(data, startTime, endTime)
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
			bundlr.Tag{Name: "Device-ID", Value: deviceID},
			bundlr.Tag{Name: "Start-Time", Value: startTime},
			bundlr.Tag{Name: "End-Time", Value: endTime},
		},
	}

	err = dataItem.Sign(c.Signer)
	return dataItem, encryptionKey, err
}

func (c *Client) Upload(dataItem bundlr.BundleItem) (string, error) {
	// TO DO
	return "", nil
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

func (c *Client) encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return []byte{}, err
	}
	nonce := make([]byte, aesGCM.NonceSize())

	return aesGCM.Seal(nonce, nonce, data, nil), nil
}
