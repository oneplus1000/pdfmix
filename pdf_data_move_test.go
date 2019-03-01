package pdfmix

import (
	"io/ioutil"
	"testing"
)

func TestMoveContent(t *testing.T) {

}

func TestMove01(t *testing.T) {
	//printDebug = true
	fp := "testing/pdf/pdf_table_docx.pdf"
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	pdfd, err := Read(b)
	if err != nil {
		t.Fatal(err)
	}

	err = pdfd.moveContent(0, 0.0, -100.0)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	data, err := Build(pdfd)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Print(data)
	err = ioutil.WriteFile("./testing/out/move01.pdf", data, 0777)
	if err != nil {
		t.Fatal(err)
	}
}
