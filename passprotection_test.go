package pdfmix

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/pkg/errors"
)

func TestPassword(t *testing.T) {
	pdfSrc := "testing/pdf/pdf_from_gopdf.pdf"
	pdfDest := "testing/out/pdf_from_gopdf_pass.pdf"
	err := passPDF(pdfSrc, pdfDest, PermissionsAll, []byte("1234"), []byte("5555"))
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func TestRc4Cip(t *testing.T) {
	passInfo := PasswordInfo{
		Permissions: PermissionsPrint | PermissionsCopy | PermissionsModify,
		OwnerPass:   []byte("1234"),
		UserPass:    []byte("5555"),
	}

	pd := PdfData{}
	_, encryKey, err := pd.createEncryptNodes(&passInfo)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if "df8aa59554" != fmt.Sprintf("%x", encryKey) {
		t.Fatalf("encryKey wrong")
	}

	//var sm bytes.Buffer
	//sm.WriteString("<239423D423AC23A223B0>") //ไรก็ได้
	//id := uint32(10)
	/*
		smCip, err := pd.createContentPwdProtection(&sm, id, encryKey)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if "0cf7a020d91cce8b3cad71da10d4fcc1630f7f6bb59a" != fmt.Sprintf("%x", smCip.Bytes()) {
			t.Fatalf("cip wrong %x", smCip.Bytes())
		}
	*/
}

func passPDF(
	pdfSrc string,
	pdfDest string,
	permissions int,
	ownerPass []byte,
	userPass []byte,
) error {

	pdfSrcFile, err := ioutil.ReadFile(pdfSrc)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(...) fail")
	}
	pdfDestFile, err := passPDFWithBytes(pdfSrcFile, permissions, ownerPass, userPass)
	if err != nil {
		return errors.Wrap(err, "passPDFWithBytes(...) fail")
	}
	err = ioutil.WriteFile(pdfDest, pdfDestFile, 0777)
	if err != nil {
		return errors.Wrap(err, "ioutil.WriteFile(...) fail")
	}
	return nil
}

func passPDFWithBytes(
	pdfSrcFile []byte,
	permissions int,
	ownerPass []byte,
	userPass []byte,
) ([]byte, error) {

	pdfData, err := Read(pdfSrcFile)
	if err != nil {
		return nil, errors.Wrap(err, "ReadPdf(...) fail")
	}

	pp := PasswordInfo{
		Permissions: permissions,
		OwnerPass:   ownerPass,
		UserPass:    userPass,
	}

	pdfDestFile, err := BuildWithOption(pdfData, &BuildPdfOption{
		PassProtection: &pp,
	})

	return pdfDestFile, nil
}
