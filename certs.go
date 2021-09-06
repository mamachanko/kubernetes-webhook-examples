package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type CertConfig struct {
	Organization string
	DNSNames     []string
	CommonName   string
	OutputPath   string
	Passphrase   string
}

func main() {
	createCert(&CertConfig{
		Organization: "mamachanko.com",
		DNSNames: []string{
			"webhook-example-go",
			"webhook-example-go.webhook-examples",
			"webhook-example-go.webhook-examples.svc",
		},
		CommonName: "webhook-example-go.webhook-examples.svc",
		OutputPath: "./k8s/certs/webhook-example-go/",
		Passphrase: "verysecret",
	})
	createCert(&CertConfig{
		Organization: "mamachanko.com",
		DNSNames: []string{
			"webhook-example-java",
			"webhook-example-java.webhook-examples",
			"webhook-example-java.webhook-examples.svc",
		},
		CommonName: "webhook-example-java.webhook-examples.svc",
		OutputPath: "./k8s/certs/webhook-example-java/",
		Passphrase: "verysecret",
	})
}

// createCert creates a TLS key, TLS certificate and PEM-encoded CA bundle.
// (inspired by https://gist.github.com/velotiotech/2e0cfd15043513d253cad7c9126d2026#file-initcontainer_main-go)
func createCert(config *CertConfig) {
	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2020),
		Subject: pkix.Name{
			Organization: []string{config.Organization},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     config.DNSNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: []string{config.Organization},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode the server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	err = os.MkdirAll(config.OutputPath, 0777)
	if err != nil {
		panic(err)
	}
	tlsCrtFile := filepath.Join(config.OutputPath, "tls.crt")
	err = WriteFile(tlsCrtFile, serverCertPEM)
	if err != nil {
		panic(err)
	}

	tlsKeyFile := filepath.Join(config.OutputPath, "tls.key")
	err = WriteFile(tlsKeyFile, serverPrivKeyPEM)
	if err != nil {
		panic(err)
	}

	caBundleFile := filepath.Join(config.OutputPath, "cabundle.pem")
	err = WriteFile(caBundleFile, caPEM)
	if err != nil {
		panic(err)
	}

	keystoreFile := filepath.Join(config.OutputPath, "keystore.p12")
	command := exec.Command(
		"openssl",
		"pkcs12",
		"-export",
		"-inkey", tlsKeyFile,
		"-in", tlsCrtFile,
		"-certfile", caBundleFile,
		"-out", keystoreFile,
		"-password", fmt.Sprintf("pass:%s", config.Passphrase),
	)
	if command.Run() != nil {
		fmt.Println(err)
	}

	fmt.Printf("created certificate files in %s\n", config.OutputPath)
}

// WriteFile writes data in the file at the given path
func WriteFile(filepath string, sCert *bytes.Buffer) error {
	f, err := os.Create(filepath)
	if err != nil {
		//return err
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(sCert.Bytes())
	if err != nil {
		return err
	}
	return nil
}
