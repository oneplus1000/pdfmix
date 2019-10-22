package pdfmix

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestReadPdf(t *testing.T) {
	//printDebugStream = true
	//printDebug = true
	//fp := "/Users/oneplus/Desktop/xxx.pdf"
	fp := "testing/out/pdf_from_gopdf_itext3 copy.pdf"
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	pdfd, err := Read(b)
	if err != nil {
		t.Fatal(err)
	}
	_ = pdfd
}

func TestFindResourcesOfPage(t *testing.T) {
	//printDebugStream = true
	//printDebug = true
	fp := "testing/pdf/jpg.pdf"
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	pdfd, err := Read(b)
	if err != nil {
		t.Fatal(err)
	}
	node, err := pdfd.findResourcesOfPage(0)
	if err != nil {
		t.Fatal(err)
	}
	if node.content.refTo.id != 7 {
		t.Fatal("wrong Resources id")
	}
}
func testInsertImage(t *testing.T, logo, fp, fpout string) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	pdfd, err := Read(b)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(logo)
	if err != nil {
		t.Fatal(err)
	}
	err = InsertImage(pdfd, f)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Build(pdfd)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("%s", result)
	//_ = result
	_, err = Read(result)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(fpout, result, 0666)
	if err != nil {
		t.Fatal(err)
	}
}
func TestInsertImage(t *testing.T) {
	logo := "testing/img/logo.jpg"
	//printDebugStream = true
	//printDebug = true
	/*
		fp := "testing/pdf/jpg.pdf"
		fpout := "testing/out/jpg.pdf"
		testInsertImage(t, logo, fp, fpout)
	*/
	fp := "testing/pdf/pdf_from_gopdf.pdf"
	fpout := "testing/out/pdf_from_gopdf-img.pdf"
	testInsertImage(t, logo, fp, fpout)
}
