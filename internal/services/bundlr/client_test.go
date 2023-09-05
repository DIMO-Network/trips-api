package bundlr

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func TestPrepareData(t *testing.T) {
	assert := assert.New(t)

	data := `{"testData": ["this", "is", "a", "test"]}`
	dataB, err := json.Marshal(data)
	assert.NoError(err)
	start, _ := time.Parse(time.DateOnly, "2023-08-16")
	end, _ := time.Parse(time.DateOnly, "2023-08-17")

	uploader, err := New(&config.Settings{
		BundlrPrivateKey: "1234567890123456789123456789123456789123456789123456789123456789",
	})
	assert.NoError(err)

	fileName := fmt.Sprintf("%s-%d-%d.zip", ksuid.New().String(), start.Unix(), end.Unix())

	// generating random 32 byte key for AES-256
	key := make([]byte, 32)
	_, err = rand.Read(key)
	assert.NoError(err)

	// compress
	compressedData, err := uploader.compress(dataB, fileName)
	assert.NoError(err)

	// encrypt using key
	encryptedData, nonce, err := uploader.encrypt(compressedData, key)
	assert.NoError(err)

	// decrypt
	aes, err := aes.NewCipher(key)
	assert.NoError(err)

	aesgcm, err := cipher.NewGCM(aes)
	assert.NoError(err)

	decryptedCompressedData, err := aesgcm.Open(nil, nonce, encryptedData, nil)

	// decompress
	unzippedResp := make([]string, 0)

	zipReader, err := zip.NewReader(bytes.NewReader(decryptedCompressedData), int64(len(decryptedCompressedData)))
	assert.NoError(err)

	for _, zipFile := range zipReader.File {
		f, err := zipFile.Open()
		assert.NoError(err)
		defer f.Close()

		unzippedBytes, err := io.ReadAll(f)
		assert.NoError(err)

		unzippedResp = append(unzippedResp, string(unzippedBytes))
	}

	assert.Equal(string(dataB), unzippedResp[0])

}
