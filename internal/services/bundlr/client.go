package bundlr

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/warp-contracts/syncer/src/utils/arweave"
	"github.com/warp-contracts/syncer/src/utils/bundlr"
)

type Client struct {
	Signer *bundlr.EthereumSigner
}

func New(settings *config.Settings) (*Client, error) {
	signer, err := bundlr.NewEthereumSigner("0x" + settings.EthereumSignerPrivateKey)
	if err != nil {
		return nil, err
	}

	return &Client{
		Signer: signer,
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
		Tags: bundlr.Tags{
			bundlr.Tag{Name: "Device-ID", Value: userDeviceID},
			bundlr.Tag{Name: "Start-Time", Value: start.Format(time.RFC3339)},
			bundlr.Tag{Name: "End-Time", Value: end.Format(time.RFC3339)},
			bundlr.Tag{Name: "Nonce", Value: hex.EncodeToString(nonce)},
		},
	}

	return dataItem, nonce, dataItem.Sign(c.Signer)
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

func (c *Client) decrypt(data, key, nonce []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return []byte{}, err
	}

	return aesgcm.Open(nil, nonce, data, nil)
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

		unzippedBytes, err := io.ReadAll(f)
		if err != nil {
			return unzippedResp, err
		}

		unzippedResp = append(unzippedResp, string(unzippedBytes))
	}

	return unzippedResp, nil
}
