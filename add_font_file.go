package pdfmix

import "github.com/pkg/errors"

func addFontFile(p *PdfData, fontfile []byte) (FontRef, error) {

	hash, err := hashSha1(fontfile)
	if err != nil {
		return FontRefEmpty, errors.Wrap(err, "hashSha1(...) fail")
	}

	if p.subsetFonts == nil {
		p.subsetFonts = make(map[FontRef](*subsetFont))
	}

	subsetFont := newSubsetFont(fontfile)
	err = subsetFont.init()
	if err != nil {
		return FontRefEmpty, errors.Wrap(err, "subsetFont.init() fail")
	}

	p.subsetFonts[FontRef(hash)] = subsetFont
	fontRef := FontRef(hash)
	return fontRef, nil
}
