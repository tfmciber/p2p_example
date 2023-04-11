package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

//derive key using argon2id

func (c *P2Papp) NewID(password string, filename string) {

	key, salt := c.DeriveKey([]byte(password), nil)

	sk := c.setID()

	ciphertext := c.encrypt(sk, key)

	//append salt to ciphertext
	ciphertext = append(ciphertext, salt...)

	c.writeKeys(filename, ciphertext)

	c.priv, _ = crypto.UnmarshalPrivateKey(sk)

}

func (c *P2Papp) OpenID(data []byte, password string) string {

	//extract salt from ciphertext
	salt := data[len(data)-16:]
	//remove salt from ciphertext
	data = data[:len(data)-16]

	pwd := []byte(password)
	key, _ := c.DeriveKey(pwd, salt)
	sk, err := c.decrypt(data, key)

	if err != nil {
		fmt.Println("Error decrypting key: ", err)
		return err.Error()
	}
	priv, err := crypto.UnmarshalPrivateKey(sk)
	if err != nil {
		return err.Error()
	}
	c.priv = priv
	return ""

}

func (c *P2Papp) DeriveKey(password []byte, salt []byte) ([]byte, []byte) {

	//if salt not provided, generate one
	if salt == nil {
		salt = make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, salt); err != nil {
			panic(err)
		}
	}
	key := argon2.Key(password, salt, 3, 32*1024, 4, 32)
	return key, salt
}

func (c *P2Papp) setID() []byte {
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}

	privM, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	return privM
}

func (c *P2Papp) requestPwd() []byte {

	fmt.Println("Enter password")
	pwd, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Panic(err)
	}
	return pwd

}
func (c *P2Papp) ReadKeys(filename string) []byte {
	// Let's first read the `config.json` file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Print("Error reading file: ", err)
		return nil
	}

	//read id as base64 byte arry

	id, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {

		log.Fatal("Error during base64 decode: ", err)
	}
	return id

}

func (c *P2Papp) writeKeys(filename string, data []byte) {

	// encode id as base 64 byte array

	idstr := []byte(base64.StdEncoding.EncodeToString(data))

	err := os.WriteFile(filename, idstr, 0644)
	if err != nil {
		log.Fatal("Error Writing file: ", err)
	}

}

func (c *P2Papp) encrypt(plaintext []byte, key []byte) []byte {

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext
}
func (c *P2Papp) decrypt(ciphertext []byte, key []byte) ([]byte, error) {

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}