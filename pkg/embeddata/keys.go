package embeddata

//go:generate go run genkeys.go

import (
	"embed"

	gossh "github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh"
)

//go:embed public/gen.pub
var publicKeyFile embed.FS

//go:embed private/gen.key
var privateKeyFile embed.FS

func GetPublicKey() (gossh.PublicKey, error) {
	keyBytes, err := publicKeyFile.ReadFile("public/gen.pub")
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
	keyBytes, err := privateKeyFile.ReadFile("private/gen.key")
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return signer, nil
}
