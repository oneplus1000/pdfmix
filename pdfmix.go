package pdfmix

import (
	"bytes"
	"io"

	"github.com/oneplus1000/pdf"
	"github.com/pkg/errors"
)

//Read read pdf file into PdfData
func Read(pdffile []byte) (*PdfData, error) {
	byteReader := bytes.NewReader(pdffile)
	pdfReader, err := pdf.NewReader(byteReader, byteReader.Size())
	if err != nil {
		return nil, errors.Wrap(err, "pdf.NewReader(...) fail")
	}
	return unmarshal(pdfReader)
}

/*
FIXME: not finished yet
//AddFont add font file into pdf file
func AddFont(p *PdfData, fontfile []byte) (FontRef, error) {
	return addFontFile(p, fontfile)
}

//AddFontByPath add font file into pdf file by path of fontfile
func AddFontByPath(p *PdfData, fontpath string) (FontRef, error) {
	b, err := ioutil.ReadFile(fontpath)
	if err != nil {
		return FontRefEmpty, errors.Wrapf(err, "ioutil.ReadFile(%s) fail", fontpath)
	}
	return AddFont(p, b)
}

//InsertText insert text to pdf

func InsertText(p *PdfData, fontRef FontRef, text string, pageIndex int, rect *Position, option *TextOption) error {
	return insertText(p, fontRef, text, pageIndex, rect, option)
}*/

//InsertImage insert image
func InsertImage(p *PdfData, imgr io.Reader) error {
	err := insertImage(p, imgr, 0, 10, 10)
	if err != nil {
		return errors.Wrap(err, "insertImage fail")
	}
	return nil
}

//Merge merge pdf b into a
func Merge(a, b *PdfData) error {
	return merge(a, b)
}

//Build create pdf file
func Build(p *PdfData) ([]byte, error) {

	err := p.build()
	if err != nil {
		return nil, errors.Wrap(err, "p.build() fail")
	}

	b, err := p.bytes()
	if err != nil {
		return nil, errors.Wrap(err, "p.bytes() fail")
	}

	return b, nil
}

//BuildWithOption build with option
func BuildWithOption(p *PdfData, option *BuildPdfOption) ([]byte, error) {
	if option.PassProtection != nil {
		err := p.pwdProtection(option.PassProtection)
		if err != nil {
			return nil, errors.Wrap(err, "p.passwordProtection(...) fail")
		}
	}
	if option.DigitalSign != nil {
		p.setDigitalSigner(option.DigitalSign)
		data, err := p.buildAndBytesWithSign()
		if err != nil {
			return nil, errors.Wrap(err, "p.buildAndBytesWithSign(...) fail")
		}
		return data, nil
	}
	return Build(p) //default
}
