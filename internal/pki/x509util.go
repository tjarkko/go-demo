package pki

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

func ReadPEMBlocks(in []byte) [][]byte {
	var out [][]byte
	for {
		var block *pem.Block
		block, in = pem.Decode(in)
		if block == nil {
			break
		}
		// We accept CERTIFICATE and TRUSTED CERTIFICATE
		if block.Type == "CERTIFICATE" || strings.HasSuffix(block.Type, "CERTIFICATE") {
			out = append(out, block.Bytes)
		}
	}
	return out
}

func TryParseCert(der []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(der)
}

func PrintCertInfo(c *x509.Certificate) {
	fmt.Printf("Subject:             %s\n", nameToOneLine(c.Subject.String()))
	fmt.Printf("Issuer:              %s\n", nameToOneLine(c.Issuer.String()))
	fmt.Printf("Serial:              %s\n", hexifyBigInt(c.SerialNumber))
	fmt.Printf("Version:             %d (X.509v%d)\n", c.Version, c.Version)
	fmt.Printf("Signature Algorithm: %s\n", c.SignatureAlgorithm)
	fmt.Printf("Public Key:          %s\n", publicKeySummary(c))
	fmt.Printf("Validity:\n")
	fmt.Printf("  Not Before:        %s\n", c.NotBefore.Format(time.RFC3339))
	fmt.Printf("  Not After:         %s\n", c.NotAfter.Format(time.RFC3339))
	fmt.Printf("Is CA:               %t\n", c.IsCA)
	if c.MaxPathLenZero {
		fmt.Printf("Path Len:            0 (MaxPathLenZero)\n")
	} else if c.MaxPathLen > 0 {
		fmt.Printf("Path Len:            %d\n", c.MaxPathLen)
	}

	if ku := keyUsageToStrings(c.KeyUsage); len(ku) > 0 {
		fmt.Printf("Key Usage:           %s\n", strings.Join(ku, ", "))
	}
	if eku := extKeyUsageToStrings(c.ExtKeyUsage); len(eku) > 0 {
		fmt.Printf("Extended Key Usage:  %s\n", strings.Join(eku, ", "))
	}

	// SANs
	var sans []string
	if len(c.DNSNames) > 0 {
		sans = append(sans, "DNS="+strings.Join(c.DNSNames, ","))
	}
	if len(c.EmailAddresses) > 0 {
		sans = append(sans, "Email="+strings.Join(c.EmailAddresses, ","))
	}
	if len(c.IPAddresses) > 0 {
		var ips []string
		for _, ip := range c.IPAddresses {
			ips = append(ips, ip.String())
		}
		sans = append(sans, "IP="+strings.Join(ips, ","))
	}
	if len(c.URIs) > 0 {
		var uris []string
		for _, u := range c.URIs {
			uris = append(uris, safeURI(u))
		}
		sans = append(sans, "URI="+strings.Join(uris, ","))
	}
	if len(sans) > 0 {
		fmt.Printf("Subject Alt Names:   %s\n", strings.Join(sans, " | "))
	}

	// IDs
	if len(c.SubjectKeyId) > 0 {
		fmt.Printf("Subject Key ID:      %s\n", hexColon(c.SubjectKeyId))
	}
	if len(c.AuthorityKeyId) > 0 {
		fmt.Printf("Authority Key ID:    %s\n", hexColon(c.AuthorityKeyId))
	}

	// AIA / CRL / OCSP
	if len(c.OCSPServer) > 0 {
		fmt.Printf("OCSP:                %s\n", strings.Join(c.OCSPServer, ", "))
	}
	if len(c.CRLDistributionPoints) > 0 {
		fmt.Printf("CRL Distribution:    %s\n", strings.Join(c.CRLDistributionPoints, ", "))
	}
	if len(c.IssuingCertificateURL) > 0 {
		fmt.Printf("AIA Issuer URL:      %s\n", strings.Join(c.IssuingCertificateURL, ", "))
	}
	if len(c.PolicyIdentifiers) > 0 {
		var oids []string
		for _, oid := range c.PolicyIdentifiers {
			oids = append(oids, oid.String())
		}
		fmt.Printf("Policy OIDs:         %s\n", strings.Join(oids, ", "))
	}

	// Fingerprints
	sha256fp := sha256.Sum256(c.Raw)
	fmt.Printf("Fingerprint SHA-256: %s\n", hexColon(sha256fp[:]))

	// Basic constraints (already partly covered)
	// Raw extensions (quick peek)
	if len(c.Extensions) > 0 {
		var exts []string
		for _, e := range c.Extensions {
			critical := ""
			if e.Critical {
				critical = " (critical)"
			}
			exts = append(exts, fmt.Sprintf("%s%s", e.Id.String(), critical))
		}
		fmt.Printf("Extensions:          %s\n", strings.Join(exts, ", "))
	}

	// Chain-building hints
	fmt.Printf("Can Verify Chains:   %t\n", canVerifyChains(c))
}

func nameToOneLine(s string) string {
	// OpenSSL-style names are comma-separated already; just squeeze whitespace.
	return strings.Join(strings.Fields(s), " ")
}

// replace hexifyBigInt and types with:
func hexifyBigInt(n *big.Int) string {
	return hexColon(n.Bytes())
}

func hexColon(b []byte) string {
	s := strings.ToUpper(hex.EncodeToString(b))
	var buf strings.Builder
	for i := 0; i < len(s); i += 2 {
		if i > 0 {
			buf.WriteByte(':')
		}
		buf.WriteString(s[i : i+2])
	}
	return buf.String()
}

func publicKeySummary(c *x509.Certificate) string {
	switch pub := c.PublicKey.(type) {
	case *rsa.PublicKey:
		return fmt.Sprintf("RSA (%d bits)", pub.N.BitLen())
	case *ecdsa.PublicKey:
		if pub.Curve != nil {
			return fmt.Sprintf("ECDSA (%s)", pub.Curve.Params().Name)
		}
		return "ECDSA"
	case ed25519.PublicKey:
		return "Ed25519"
	default:
		return fmt.Sprintf("%T", pub)
	}
}

func keyUsageToStrings(ku x509.KeyUsage) []string {
	var out []string
	add := func(b x509.KeyUsage, name string) {
		if ku&b != 0 {
			out = append(out, name)
		}
	}
	add(x509.KeyUsageDigitalSignature, "DigitalSignature")
	add(x509.KeyUsageContentCommitment, "ContentCommitment")
	add(x509.KeyUsageKeyEncipherment, "KeyEncipherment")
	add(x509.KeyUsageDataEncipherment, "DataEncipherment")
	add(x509.KeyUsageKeyAgreement, "KeyAgreement")
	add(x509.KeyUsageCertSign, "CertSign")
	add(x509.KeyUsageCRLSign, "CRLSign")
	add(x509.KeyUsageEncipherOnly, "EncipherOnly")
	add(x509.KeyUsageDecipherOnly, "DecipherOnly")
	return out
}

func extKeyUsageToStrings(eku []x509.ExtKeyUsage) []string {
	names := map[x509.ExtKeyUsage]string{
		x509.ExtKeyUsageAny:                        "Any",
		x509.ExtKeyUsageServerAuth:                 "ServerAuth",
		x509.ExtKeyUsageClientAuth:                 "ClientAuth",
		x509.ExtKeyUsageCodeSigning:                "CodeSigning",
		x509.ExtKeyUsageEmailProtection:            "EmailProtection",
		x509.ExtKeyUsageIPSECEndSystem:             "IPSECEndSystem",
		x509.ExtKeyUsageIPSECTunnel:                "IPSECTunnel",
		x509.ExtKeyUsageIPSECUser:                  "IPSECUser",
		x509.ExtKeyUsageTimeStamping:               "TimeStamping",
		x509.ExtKeyUsageOCSPSigning:                "OCSPSigning",
		x509.ExtKeyUsageMicrosoftServerGatedCrypto: "MS SGC",
		x509.ExtKeyUsageNetscapeServerGatedCrypto:  "Netscape SGC",
	}
	var out []string
	for _, e := range eku {
		if s, ok := names[e]; ok {
			out = append(out, s)
		} else {
			out = append(out, fmt.Sprintf("Unknown(%d)", e))
		}
	}
	return out
}

func safeURI(u *url.URL) string {
	uu := *u
	uu.User = nil // don’t print creds
	return uu.String()
}

func canVerifyChains(c *x509.Certificate) bool {
	// basic sanity: leaf vs CA; not authoritative verification obviously
	return c.IsCA || len(c.DNSNames) > 0 || len(c.IPAddresses) > 0 || len(c.EmailAddresses) > 0
}
