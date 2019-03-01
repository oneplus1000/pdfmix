package pdfmix

import (
	"io/ioutil"
	"testing"

	"github.com/pkg/errors"
)

func TestSignPDFDeves(t *testing.T) {
	/*
		err := signPDF("testing/pdf/deves.pdf", "testing/out/deves_cert.pdf", "testing/cert/tpb.p12", "Thaipa*9")
		if err != nil {
			t.Fatalf("%+v", err)
		}*/
	err := signPDF("/Users/oneplus/Desktop/test.pdf", "testing/out/deves_cert.pdf", "testing/cert/tpb.p12", "Thaipa*9")
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func TestSignPDF(t *testing.T) {

	err := signPDF("testing/pdf/pdf_from_gopdf.pdf", "testing/out/pdf_from_gopdf_cert.pdf", "testing/cert/tpb.p12", "Thaipa*9")
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func BenchmarkSignPDF(b *testing.B) {

	pdfSrc := "testing/pdf/pdf_from_gopdf.pdf"
	//"testing/out/pdf_from_gopdf_cert.pdf"
	cert := "testing/cert/tpb.p12"

	certFile, err := ioutil.ReadFile(cert)
	if err != nil {
		b.Fatalf("%+v", err)
	}
	pdfSrcFile, err := ioutil.ReadFile(pdfSrc)
	if err != nil {
		b.Fatalf("%+v", err)
	}

	for i := 0; i < b.N; i++ { //ทดสอบว่าได้กี่รอบ
		result, err := signPDFWithBytes(pdfSrcFile, certFile, "Thaipa*9")
		if err != nil {
			b.Fatalf("%+v", err)
		}
		_ = result
	}
}

func signPDF(pdfSrc string, pdfDest string, cert, certpassword string) error {

	certFile, err := ioutil.ReadFile(cert)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(...) fail")
	}
	pdfSrcFile, err := ioutil.ReadFile(pdfSrc)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(...) fail")
	}

	pdfDestFile, err := signPDFWithBytes(pdfSrcFile, certFile, certpassword)
	if err != nil {
		return errors.Wrap(err, "ioutil.ReadFile(...) fail")
	}

	err = ioutil.WriteFile(pdfDest, pdfDestFile, 0777)
	if err != nil {
		return errors.Wrap(err, "ioutil.WriteFile(...) fail")
	}

	return nil
}

func signPDFWithBytes(pdfSrcFile []byte, certFile []byte, certpassword string) ([]byte, error) {
	logo := "testing/img/logo.jpg"
	var ds DigitalSignPkcs12
	err := ds.SetCertFile(certFile, certpassword)
	if err != nil {
		return nil, errors.Wrap(err, "dsitem.SetCertPkcs12(...) fail")
	}
	b, err := ioutil.ReadFile(logo)
	if err != nil {
		return nil, err
	}
	annot := AnnotInfo{
		Img: b,
		Pos: Position{
			X: 100.0,
			Y: 100.0,
			H: 100.0,
			W: 100.0,
		},
	}
	ds.SetAnnot(&annot)
	//_ = annot

	pdfData, err := Read(pdfSrcFile)
	if err != nil {
		return nil, errors.Wrap(err, "ReadPdf(...) fail")
	}
	/*
		flogo, err := os.Open(logo)
		if err != nil {
			return nil, err
		}
			err = InsertImage(pdfData, flogo)
			if err != nil {
				return nil, err
			}
	*/
	pdfDestFile, err := BuildWithOption(pdfData, &BuildPdfOption{
		DigitalSign: &ds,
	})
	if err != nil {
		return nil, errors.Wrap(err, "BuildPdfWithDigitalSignatures(...) fail")
	}

	return pdfDestFile, nil
}
