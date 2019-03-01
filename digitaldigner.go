package pdfmix

import "crypto/x509"

//DigitalSigner ตัวทีต่อกะ key
type DigitalSigner interface {
	Cert() *x509.Certificate          //return cert
	Sign(data []byte) ([]byte, error) //sign
}

//DigitalSignWithAnnoter DigitalSigner ตัวมี antto ด้วย
type DigitalSignWithAnnoter interface {
	DigitalSigner
	Annot() *AnnotInfo
}

//AnnotInfo antto
type AnnotInfo struct {
	Img       []byte
	Text      string
	PageIndex int
	Pos       Position
}
