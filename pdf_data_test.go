package pdfmix

import (
	"testing"

	"github.com/pkg/errors"
)

func TestFontnameExtract(t *testing.T) {
	err := testFontnameExtract("C2_0", "C2_", 0)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = testFontnameExtract("C2", "C", 2)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = testFontnameExtract("2", "", 2)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = testFontnameExtract("F", "F", 0)
	if err == nil {
		t.Fatalf("must error")
	}
}

func testFontnameExtract(fontName string, needName string, needIndex int) error {
	p := PdfData{}
	name, index, err := p.fontnameExtract(fontName)
	if err != nil {
		return err
	}
	if name != needName {
		return errors.New("wrong name")
	}
	if index != needIndex {
		return errors.New("wrong index")
	}
	return nil
}
