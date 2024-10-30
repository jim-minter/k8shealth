package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math"
	"math/big"
	"os"
	"time"
)

// helper to create CA and signed certificates

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := os.MkdirAll("certs", 0777); err != nil {
		return err
	}

	_, _, err := cert(os.Args[1])
	return err
}

func key(commonname string) (*rsa.PrivateKey, error) {
	filename := "certs/" + commonname + ".key"

	if b, err := os.ReadFile(filename); err == nil {
		for {
			var block *pem.Block
			block, b = pem.Decode(b)

			if block == nil {
				break
			}

			if block.Type == "RSA PRIVATE KEY" {
				if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
					return key, nil
				}
			}
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	if err = os.WriteFile(filename, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0600); err != nil {
		return nil, err
	}

	return key, nil
}

func ca() (*rsa.PrivateKey, *x509.Certificate, error) {
	filename := "certs/@ca.crt"

	cakey, err := key("@ca")
	if err != nil {
		return nil, nil, err
	}

	if b, err := os.ReadFile(filename); err == nil {
		for {
			var block *pem.Block
			block, b = pem.Decode(b)

			if block == nil {
				break
			}

			if block.Type == "CERTIFICATE" {
				if cacert, err := x509.ParseCertificate(block.Bytes); err == nil {
					return cakey, cacert, nil
				}
			}
		}
	}

	sn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)

	template := &x509.Certificate{
		SerialNumber:          sn,
		Issuer:                pkix.Name{CommonName: "ca"},
		Subject:               pkix.Name{CommonName: "ca"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	b, err := x509.CreateCertificate(rand.Reader, template, template, &cakey.PublicKey, cakey)
	if err != nil {
		return nil, nil, err
	}

	if err = os.WriteFile(filename, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b}), 0666); err != nil {
		return nil, nil, err
	}

	cacert, err := x509.ParseCertificate(b)
	if err != nil {
		return nil, nil, err
	}

	return cakey, cacert, nil
}

func cert(commonname string) (*rsa.PrivateKey, *x509.Certificate, error) {
	filename := "certs/" + commonname + ".crt"

	key, err := key(commonname)
	if err != nil {
		return nil, nil, err
	}

	if b, err := os.ReadFile(filename); err == nil {
		for {
			var block *pem.Block
			block, b = pem.Decode(b)

			if block == nil {
				break
			}

			if block.Type == "CERTIFICATE" {
				if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
					return key, cert, nil
				}
			}
		}
	}

	cakey, cacert, err := ca()
	if err != nil {
		return nil, nil, err
	}

	sn, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.AddDate(1, 0, 0)
	if notAfter.After(cacert.NotAfter) {
		notAfter = cacert.NotAfter
	}

	template := &x509.Certificate{
		SerialNumber:          sn,
		Issuer:                pkix.Name{CommonName: "ca"},
		Subject:               pkix.Name{CommonName: commonname},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		BasicConstraintsValid: true,
		DNSNames:              []string{commonname},
	}

	b, err := x509.CreateCertificate(rand.Reader, template, cacert, &key.PublicKey, cakey)
	if err != nil {
		return nil, nil, err
	}

	if err = os.WriteFile(filename, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b}), 0666); err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(b)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, nil
}
