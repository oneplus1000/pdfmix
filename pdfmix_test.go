package pdfmix

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestMerge(t *testing.T) {
	path01 := "testing/pdf/png.pdf"
	path02 := "testing/pdf/jpg.pdf"
	b01, err := ioutil.ReadFile(path01)
	if err != nil {
		log.Panicf("%+v", err)
	}

	pdf01, err := Read(b01)
	if err != nil {
		log.Panicf("%+v", err)
	}

	b02, err := ioutil.ReadFile(path02)
	if err != nil {
		log.Panicf("%+v", err)
	}
	pdf02, err := Read(b02)
	if err != nil {
		log.Panicf("%+v", err)
	}

	err = Merge(pdf01, pdf02)
	if err != nil {
		log.Panicf("%+v", err)
	}

	bResult, err := Build(pdf01)
	if err != nil {
		log.Panicf("%+v", err)
	}

	err = ioutil.WriteFile("testing/out/out_example_merge.pdf", bResult, 0644)
	if err != nil {
		log.Panicf("%+v", err)
	}
}
