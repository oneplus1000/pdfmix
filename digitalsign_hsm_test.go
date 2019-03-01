package pdfmix

import (
	"io/ioutil"
	"testing"

	"github.com/oneplus1000/pkcs11"
	"github.com/pkg/errors"
)

func _TestSignPDFByHsm2(t *testing.T) {
	pdfSrc := "testing/pdf/pdf_from_gopdf.pdf"
	pdfDest := "testing/out/pdf_from_gopdf_hsm.pdf"
	err := testSignPDFByHsm(pdfSrc, pdfDest)
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func testSignPDFByHsm(pdfSrc string, pdfDest string) error {

	pdfSrcFile, err := ioutil.ReadFile(pdfSrc)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(...) fail")
	}

	pdfDestFile, err := testSignPDFByHsmWithBytes(pdfSrcFile)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(...) fail")
	}

	err = ioutil.WriteFile(pdfDest, pdfDestFile, 0777)
	if err != nil {
		return errors.Wrap(err, "ioutil.WriteFile(...) fail")
	}

	return nil
}

func testSignPDFByHsmWithBytes(pdfSrcFile []byte) ([]byte, error) {

	var ds DigitalSignHsm
	err := ds.OpenShm("/usr/local/Cellar/softhsm/2.3.0/lib/softhsm/libsofthsm2.so",
		"tpb",
		"tpb",
		"1234",
		1,
	)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	defer ds.CloseHsm()

	pdfData, err := Read(pdfSrcFile)
	if err != nil {
		return nil, errors.Wrap(err, "ReadPdf(...) fail")
	}

	pdfDestFile, err := BuildWithOption(pdfData, &BuildPdfOption{
		DigitalSign: &ds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "BuildPdfWithDigitalSignatures(...) fail")
	}

	return pdfDestFile, nil
}

func _TestPkcs11(t *testing.T) {
	lib := "/usr/local/Cellar/softhsm/2.3.0/lib/softhsm/libsofthsm2.so"
	ctx := pkcs11.New(lib)
	err := ctx.Initialize()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	slots, err := ctx.GetSlotList(true)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	session, err := ctx.OpenSession(slots[0], pkcs11.CKF_SERIAL_SESSION|pkcs11.CKF_RW_SESSION)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	err = ctx.Login(session, pkcs11.CKU_USER, "1234")
	if err != nil {
		t.Fatalf("%+v", err)
	}
}
