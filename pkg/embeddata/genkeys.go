//go:build ignore

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"
)

func main() {
	pubKeyPath := "public/gen.pub"
	privateKeyPath := "private/gen.key"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// generate and write private key as PEM
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	defer privateKeyFile.Close()

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
		log.Fatal(err)
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(pubKeyPath, ssh.MarshalAuthorizedKey(pub), 0655)
	if err != nil {
		log.Fatal(err)
	}
}
