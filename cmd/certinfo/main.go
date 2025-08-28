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
