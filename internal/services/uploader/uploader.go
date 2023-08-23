package uploader

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

type Uploader struct {
	Signer      *bundlr.EthereumSigner
	url         string
	contentType string
}
type uploadResp struct {
	ID        string `json:"id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

func New(settings *config.Settings) (*Uploader, error) {
	signer, err := bundlr.NewEthereumSigner("0x" + settings.EthereumSignerPrivateKey)
	if err != nil {
		return nil, err
	}

	return &Uploader{
		Signer:      signer,
		url:         settings.BundlrNetwork,
		contentType: "application/octet-stream",
	}, nil
}

// PrepareData prepares data for uploading to bundlr by compressing and encrypting input.
func (u *Uploader) PrepareData(data []byte, deviceID, startTime, endTime string) (bundlr.BundleItem, string, error) {
	compressedData, err := u.compress(data, startTime, endTime)
	if err != nil {
		return bundlr.BundleItem{}, "", err
	}

	// generating random 32 byte key for AES-256
	// this will change with PRO-1867 encryption keys created for minted devices
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return bundlr.BundleItem{}, "", err
	}
	encryptionKey := hex.EncodeToString(bytes)

	encryptedData, err := u.encrypt(compressedData, bytes)
	if err != nil {
		return bundlr.BundleItem{}, encryptionKey, err
	}

	dataItem := bundlr.BundleItem{
		Data: arweave.Base64String(encryptedData),
		// in the future-- allow tags to be passed in
		// ie, someone could name their trip
		Tags: bundlr.Tags{
			bundlr.Tag{Name: "Content-Type", Value: "text"},
			bundlr.Tag{Name: "Trip-Type", Value: "segment"},
			bundlr.Tag{Name: "Device-ID", Value: deviceID},
			bundlr.Tag{Name: "Start-Time", Value: startTime},
			bundlr.Tag{Name: "End-Time", Value: endTime},
		},
	}

	err = dataItem.Sign(u.Signer)
	return dataItem, encryptionKey, err
}

func (u *Uploader) Upload(dataItem bundlr.BundleItem) (string, error) {
	// TO DO
	return "", nil

	// reader, err := dataItem.Reader()
	// if err != nil {
	// 	return "", err
	// }

	// body, err := io.ReadAll(reader)
	// if err != nil {
	// 	return "", err
	// }

	// responseBody := bytes.NewBuffer(body)
	// resp, err := http.Post(u.url, u.contentType, responseBody)
	// if err != nil {
	// 	return "", err
	// }

	// defer resp.Body.Close()

	// // if resp.Header.Get("Content-Type") == "text/plain; charset=utf-8" {
	// // 	// the tx was already uploaded
	// // }

	// var r uploadResp
	// body, err = ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return "", err
	// }

	// err = json.Unmarshal(body, &r)
	// if err != nil {
	// 	return "", err
	// }

	// return string(body), nil
}

func (u *Uploader) compress(data []byte, start, end string) ([]byte, error) {
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

func (u *Uploader) encrypt(data, key []byte) ([]byte, error) {
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
