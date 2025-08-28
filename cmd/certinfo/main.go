package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/tjarkko/go-demo/internal/pki"
)

type Context struct {
	Debug bool
}

type PrintCmd struct {
	FilePath string `arg:"" name:"cert-file" help:"Cert file." type:"existingfile"`
}

// pemToRSA turns a PEM-encoded RSA public key into an rsa.PublicKey value.
// Intended for use on startup, so panics if any part of the decoding fails.
/*func pemToRSA(pemtxt string) *rsa.PublicKey {
	var pubkey *rsa.PublicKey
	block, _ := pem.Decode([]byte(pemtxt))
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	pubkey = cert.PublicKey.(*rsa.PublicKey)
	return pubkey
}*/

/*func pemToRSA(pemData []byte) (pubKey *rsa.PublicKey, issuer pkix.Name) {
	block, _ := pem.Decode(pemData)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	issuer = cert.Issuer
	pubKey = cert.PublicKey.(*rsa.PublicKey)
	return
}

func pemToX509(pemData []byte) *x509.Certificate {
	block, _ := pem.Decode(pemData)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return cert
}*/

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

func (p *PrintCmd) Run(ctx *Context) error {
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		fail(err)
	}

	blocks := pki.ReadPEMBlocks(data)
	if len(blocks) == 0 {
		// Maybe DER
		blocks = [][]byte{data}
	}

	for i, b := range blocks {
		cert, err := pki.TryParseCert(b)
		if err != nil {
			fmt.Printf("#%d: not a certificate: %v\n\n", i+1, err)
			continue
		}
		fmt.Printf("===== Certificate #%d =====\n", i+1)
		pki.PrintCertInfo(cert)
		fmt.Println()
	}

	return nil
}

var cli struct {
	Debug bool `help:"Enable debug mode."`

	Print PrintCmd `cmd:"" help:"Print cert."`
}

func main() {
	ctx := kong.Parse(&cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run(&Context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}
