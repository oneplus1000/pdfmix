package pdfmix

import (
	"io/ioutil"
	"testing"
)

func TestExtractImgs(t *testing.T) {
	pdfSrc := "testing/pdf/7thGARDENSM.pdf"
	//pdfDest := "testing/out/pdf_from_gopdf_pass.pdf"
	b, err := ioutil.ReadFile(pdfSrc)
	if err != nil {
		t.Fatal(err)
	}
	pdfd, err := Read(b)
	if err != nil {
		t.Fatal(err)
	}
	err = ExtractImgs(pdfd)
	if err != nil {
		t.Fatalf("%+v", err)
	}
}
