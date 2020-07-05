package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func LoadPrivateKey(fileName string)(*rsa.PrivateKey, error){
	var priv []byte
	priv, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	privPem, _ := pem.Decode(priv)

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPem.Bytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPem.Bytes); err != nil {
			return nil, err
		}
	}

	key, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key parsing error")
	}

	return key, nil
}


func Decrypt(privateKey *rsa.PrivateKey, b64CipherText string, label string) (string, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(b64CipherText)
	if err != nil {
		panic(err)
	}

	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, decodedCiphertext, []byte(label))
	if err != nil {
		return "", fmt.Errorf("Error from decryption: %s\n", err)
	}

	return string(plaintext), nil
}