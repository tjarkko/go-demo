package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/tjarkko/go-demo/internal/pki"
)

var addr = flag.String("addr", ":8080", "http service address")

var templ = template.Must(template.New("qr").Parse(templateStr))

func main() {
	flag.Parse()
	http.Handle("/", http.HandlerFunc(CertInfo))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func CertInfo(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		// Show the upload form
		templ.Execute(w, "")
		return
	}

	file, header, err := req.FormFile("cert")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := make([]byte, header.Size)
	file.Read(data)

	blocks := pki.ReadPEMBlocks(data)
	if len(blocks) == 0 {
		// Maybe DER
		blocks = [][]byte{data}
	}

	var certInfos []string
	for i, b := range blocks {
		cert, err := pki.TryParseCert(b)
		if err != nil {
			certInfos = append(certInfos, fmt.Sprintf("#%d: not a certificate: %v", i+1, err))
			continue
		}
		certInfo := fmt.Sprintf("===== Certificate #%d =====\n", i+1)
		certInfo += pki.GetCertInfoString(cert)
		certInfos = append(certInfos, certInfo)
	}

	// Join all certificate information with double newlines
	allCertInfo := strings.Join(certInfos, "\n\n")
	templ.Execute(w, allCertInfo)
}

const templateStr = `
<html>
<head>
<title>Display X.509 certificate info</title>
<style>
body {
    font-family: Arial, sans-serif;
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
    background-color: #f5f5f5;
}
.container {
    background-color: white;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.upload-form {
    margin-bottom: 20px;
    padding: 20px;
    border: 2px dashed #ccc;
    border-radius: 8px;
    text-align: center;
}
.upload-form:hover {
    border-color: #007bff;
}
.cert-info {
    background-color: #f8f9fa;
    border: 1px solid #dee2e6;
    border-radius: 4px;
    padding: 15px;
    font-family: 'Courier New', monospace;
    white-space: pre-wrap;
    overflow-x: auto;
    margin-top: 20px;
}
input[type="file"] {
    margin: 10px 0;
}
button {
    background-color: #007bff;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 4px;
    cursor: pointer;
}
button:hover {
    background-color: #0056b3;
}
</style>
</head>
<body>
<div class="container">
    <h1>X.509 Certificate Information</h1>
    
    <div class="upload-form">
        <form method="POST" enctype="multipart/form-data" action="/">
            <label for="cert"><strong>Upload certificate:</strong></label><br>
            <input type="file" id="cert" name="cert"
                    accept=".pem,.crt,.cer,.der,
                            application/x-pem-file,
                            application/x-x509-ca-cert,
                            application/pem-certificate-chain,
                            application/pkix-cert,
                            application/octet-stream">
            <br>
            <button type="submit">Analyze Certificate</button>
        </form>
    </div>

    {{if .}}
    <div class="cert-info">
{{.}}
    </div>
    {{end}}
</div>
</body>
</html>
`
