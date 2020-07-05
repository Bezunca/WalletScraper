package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
)

func savePEMKey(fileName string, key *rsa.PrivateKey) {
	outFile, err := os.Create(fileName)
	if err != nil{
		panic(err)
	}
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	err = pem.Encode(outFile, privateKey)
	if err != nil{
		panic(err)
	}
}

func savePublicPEMKey(fileName string, pubkey rsa.PublicKey) {
	asn1Bytes, err := asn1.Marshal(pubkey)
	if err != nil{
		panic(err)
	}

	var pemkey = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	pemfile, err := os.Create(fileName)
	if err != nil{
		panic(err)
	}
	defer pemfile.Close()

	err = pem.Encode(pemfile, pemkey)
	if err != nil{
		panic(err)
	}
}

func generateRSAKeys(privateKeyFileName, publicKeyFileName string){
	rng := rand.Reader
	key, err := rsa.GenerateKey(rng, 4096)
	if err != nil{
		panic(err)
	}

	savePEMKey(privateKeyFileName, key)
	savePublicPEMKey(publicKeyFileName, key.PublicKey)
}

func loadPrivateKey(fileName string)(*rsa.PrivateKey, error){
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

func loadPublicKey(fileName string)(*rsa.PublicKey, error){
	pub, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	pubPem, _ := pem.Decode(pub)

	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PublicKey(pubPem.Bytes); err != nil {
		return nil, err
	}

	pubKey, ok := parsedKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("key parsing error")
	}

	return pubKey, nil
}

func main(){
	///////////////////////////////// ENCRYPT /////////////////////////////////
	publicKey, err := loadPublicKey("public.pem")
	if err != nil {
		panic(err)
	}
	print(publicKey.E)

	secretMessage := []byte("1234@asdf")
	label := []byte("cei_password")

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, secretMessage, label)
	if err != nil {
		panic(err)
	}

	b64Ciphertext := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Printf("B64 Ciphertext: %s\n", b64Ciphertext)


	///////////////////////////////// DECRYPT /////////////////////////////////
	privateKey, err := loadPrivateKey("private.key")
	if err != nil {
		panic(err)
	}
	print(publicKey.E)

	decodedCiphertext, err := base64.StdEncoding.DecodeString(b64Ciphertext)
	if err != nil {
		panic(err)
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, decodedCiphertext, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from decryption: %s\n", err)
		return
	}

	fmt.Printf("Plaintext: %s\n", plaintext)
}