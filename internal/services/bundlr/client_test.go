package bundlr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPrepareData(t *testing.T) {
	assert := assert.New(t)

	data := `{"testData": ["this", "is", "a", "test"]}`
	dataB, err := json.Marshal(data)
	assert.NoError(err)
	start := "2023-08-16"
	end := "2023-08-17"

	uploader, err := New(&config.Settings{})
	assert.NoError(err)

	compressedData, err := uploader.compress(dataB, start, end)
	assert.NoError(err)

	// generating random 32 byte key for AES-256
	// this will change with PRO-1867 encryption keys created for minted
	key := make([]byte, 32)
	_, err = rand.Read(key)
	assert.NoError(err)

	keyString := hex.EncodeToString(key)
	t.Logf("Key: %s", keyString)

	encryptedData, err := uploader.encrypt(compressedData, key)
	assert.NoError(err)

	// decrypt and check to make sure

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	assert.NoError(err)

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	assert.NoError(err)

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	//Decrypt the data
	decryptedCompressedData, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	assert.NoError(err)

	assert.Equal(decryptedCompressedData, compressedData)

}
