package embedkeys

import (
	"embed"

	gossh "github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh"
)

//go:embed public/id_rsa.pub
var publicKeyFile embed.FS

//go:embed private/id_rsa.key
var privateKeyFile embed.FS

func GetPublicKey() (gossh.PublicKey, error) {
	keyBytes, err := publicKeyFile.ReadFile("public/id_rsa.pub")
	if err != nil {
		return nil, err
	}

	publicKey, _, _, _, err := gossh.ParseAuthorizedKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func GetPrivateKey() (gossh.Signer, error) {
	keyBytes, err := privateKeyFile.ReadFile("private/id_rsa.key")
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return signer, nil
}
