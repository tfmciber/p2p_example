// main.go
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"github.com/libp2p/go-libp2p/core/crypto"
	"golang.org/x/term"
)

// The data struct for the decoded data
// Notice that all fields must be exportable!
type identity struct {

	// defining struct variables
	PrivObj []byte
	Public  string
}

func setID() ([]byte, string) {
	priv, pub, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}
	//crypto.UnmarshalEd25519PrivateKey()
	privM, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		panic(err)
	}

	return privM, fmt.Sprintf("%x", pub)
}

func initPriv(filename string) crypto.PrivKey {

	fmt.Println("[*] Init Private Identity")
	_, err := os.ReadFile(filename)

	var pub string
	var pObj []byte
	if err != nil {
		fmt.Println("Profile not found, Creating new")

		bytePassword, err := requestPwd(false)
		if err != nil {
			fmt.Print(err)
		}
		key := hex.EncodeToString(newSHA256([]byte(bytePassword)))
		pObj, pub = setID()

		writeKeys(filename, pObj, pub, key)
	} else {
		fmt.Println("Using values from profile file ", filename)

		for {
			bytePassword, err := requestPwd(true)
			if err != nil {
				fmt.Print(err)
			}
			key := hex.EncodeToString(newSHA256([]byte(bytePassword)))
			pObj, pub, err = readKeys(filename, key)
			if err == nil {
				break
			}
		}
	}
	priv, err := crypto.UnmarshalPrivateKey(pObj)
	if err != nil {
		fmt.Print(err)
	}
	return priv

}

func requestPwd(ver bool) ([]byte, error) {

	var bytePassword1 []byte
	var bytePassword2 []byte
	var err error
	for {
		fmt.Println("Enter password")
		bytePassword1, err = term.ReadPassword(int(syscall.Stdin))

		if err != nil {
			return nil, err
		}
		if ver {
			return bytePassword1, err
		}
		fmt.Println("Confirm password")
		bytePassword2, err = term.ReadPassword(int(syscall.Stdin))

		if err != nil {
			return nil, err
		}
		if bytes.Compare(bytePassword1, bytePassword2) == 0 {
			return bytePassword1, err
		}

	}

}
func readKeys(filename string, key string) ([]byte, string, error) {
	// Let's first read the `config.json` file
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error during readfile(): ", err)
	}

	// Now let's unmarshall the data into `payload`
	var payload identity
	err = json.Unmarshal(content, &payload)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	payload.PrivObj, err = decrypt(payload.PrivObj, key)

	return payload.PrivObj, payload.Public, err

}

func writeKeys(filename string, PrivObj []byte, public string, key string) {

	id := identity{encrypt(PrivObj, key), public}
	file, err := json.Marshal(id)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		log.Fatal("Error Writing file: ", err)
	}

}

func encrypt(plaintext []byte, keyString string) []byte {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)

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
func decrypt(ciphertext []byte, keyString string) ([]byte, error) {

	key, _ := hex.DecodeString(keyString)

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

func newSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
