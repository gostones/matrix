package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

// RasKeyPair returns private key, public key, or error
func RsaKeyPair(file ...string) ([]byte, []byte, error) {
	k, err := privateKey()
	if err != nil {
		return nil, nil, err
	}

	pb, err := publicKey(&k.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	b := encodePrivateKey(k)

	//
	if len(file) > 0 {
		name := file[0]
		err = saveKey(name, b)
		if err != nil {
			return nil, nil, err
		}
		err = saveKey(name+".pub", pb)
		if err != nil {
			return nil, nil, err
		}
	}

	return b, pb, nil
}

func saveKey(file string, b []byte) error {
	err := ioutil.WriteFile(file, b, 0600)
	if err != nil {
		return err
	}

	return nil
}

func privateKey() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	err = key.Validate()
	if err != nil {
		return nil, err
	}
	return key, nil
}

func encodePrivateKey(key *rsa.PrivateKey) []byte {
	der := x509.MarshalPKCS1PrivateKey(key)

	b := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   der,
	}

	return pem.EncodeToMemory(&b)
}

func publicKey(key *rsa.PublicKey) ([]byte, error) {
	p, err := ssh.NewPublicKey(key)
	if err != nil {
		return nil, err
	}

	return ssh.MarshalAuthorizedKey(p), nil
}
