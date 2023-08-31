package bundlr

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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
		EthereumSignerPrivateKey: "1234567890123456789123456789123456789123456789123456789123456789",
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
	decryptedCompressedData, err := uploader.decrypt(encryptedData, key, nonce)
	assert.NoError(err)
	assert.Equal(compressedData, decryptedCompressedData)

	// decompress
	original, err := uploader.decompress(decryptedCompressedData)
	assert.NoError(err)

	assert.Equal(string(dataB), original[0])

}
