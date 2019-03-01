package pdfmix

import (
	"crypto/x509"

	"github.com/oneplus1000/pkcs11"
	"github.com/pkg/errors"
)

//DigitalSignHsm digital sign for hsm
type DigitalSignHsm struct {
	cert  *x509.Certificate
	sw    SoftHsmWrap
	annot *AnnotInfo
}

//Cert get cert information
func (d *DigitalSignHsm) Cert() *x509.Certificate {
	return d.cert
}

//Sign sign data
func (d *DigitalSignHsm) Sign(data []byte) ([]byte, error) {
	//pkcs11.CKM_SHA256_RSA_PKCS
	pkv, err := d.sw.GetPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	d.sw.Ctx().SignInit(d.sw.Session(), []*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_SHA256_RSA_PKCS, nil)}, pkv)
	hash, err := d.sw.Ctx().Sign(d.sw.Session(), data)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return hash, nil
}

//OpenShm open
func (d *DigitalSignHsm) OpenShm(
	LibraryPath string,
	CertLabel string,
	PrivLabel string,
	PIN string,
	SlotIndex int,
) error {
	d.sw = SoftHsmWrap{
		LibraryPath: LibraryPath,
		PIN:         PIN,
		CertLabel:   CertLabel,
		PrivLabel:   PrivLabel,
		SlotIndex:   SlotIndex,
	}
	err := d.sw.Open()
	if err != nil {
		return errors.Wrap(err, "")
	}

	certraw, err := d.sw.GetSigningCertificate()
	if err != nil {
		return errors.Wrap(err, "")
	}

	cert, err := x509.ParseCertificate(certraw)
	if err != nil {
		return errors.Wrap(err, "")
	}
	d.cert = cert

	return nil
}

//CloseHsm close hsm
func (d *DigitalSignHsm) CloseHsm() {
	d.sw.Close()
}

//SetAnnot set annto
func (d *DigitalSignHsm) SetAnnot(a *AnnotInfo) error {
	d.annot = a
	return nil
}

//Annto get antto
func (d *DigitalSignHsm) Annot() *AnnotInfo {
	return d.annot
}
