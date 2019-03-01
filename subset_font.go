package pdfmix

import (
	"github.com/oneplus1000/pdfmix/font"
	"github.com/pkg/errors"
)

type subsetFont struct {
	fontfileRaw []byte
	glyphIndexs map[rune]uint
	ttfp        font.TTFParser
}

func newSubsetFont(fontfile []byte) *subsetFont {
	var s subsetFont
	s.glyphIndexs = make(map[rune]uint)
	s.fontfileRaw = fontfile
	return &s
}

func (s *subsetFont) init() error {
	s.ttfp.SetUseKerning(true)
	err := s.ttfp.ParseByBytes(s.fontfileRaw)
	if err != nil {
		return errors.Wrap(err, "")
	}
	return nil
}

func (s *subsetFont) addChars(text string) error {

	for _, r := range text {
		if _, ok := s.glyphIndexs[r]; ok {
			continue
		}
		glyphIndex, err := s.charCodeToGlyphIndex(r)
		if err != nil {
			return errors.Wrap(err, "")
		}
		s.glyphIndexs[r] = glyphIndex
	}

	return nil
}

func (s *subsetFont) getGlyphIndex(r rune) (uint, error) {

	if idx, ok := s.glyphIndexs[r]; ok {
		return idx, nil
	}

	return 0, ErrRuneNotFound
}

//charCodeToGlyphIndex get glyph index from char code
func (s *subsetFont) charCodeToGlyphIndex(r rune) (uint, error) {

	value := uint64(r)
	if value <= 0xFFFF {
		index, err := s.charCodeToGlyphIndexFormat4(r)
		if err != nil {
			return 0, errors.Wrap(err, "")
		}
		return index, nil
	}

	index, err := s.charCodeToGlyphIndexFormat12(r)
	if err != nil {
		return 0, errors.Wrap(err, "")
	}

	return index, nil

}

func (s *subsetFont) charCodeToGlyphIndexFormat12(r rune) (uint, error) {

	value := uint(r)
	gTbs := s.ttfp.GroupingTables()
	for _, gTb := range gTbs {
		if value >= gTb.StartCharCode && value < gTb.EndCharCode {
			index := (value - gTb.StartCharCode) + gTb.GlyphID
			return index, nil
		}
	}

	return 0, ErrGlyphNotFound
}

func (s *subsetFont) charCodeToGlyphIndexFormat4(r rune) (uint, error) {
	value := uint(r)
	seg := uint(0)
	segCount := s.ttfp.SegCount
	for seg < segCount {
		if value <= s.ttfp.EndCount[seg] {
			break
		}
		seg++
	}

	if value < s.ttfp.StartCount[seg] {
		return 0, nil
	}

	if s.ttfp.IdRangeOffset[seg] == 0 {

		return (value + s.ttfp.IdDelta[seg]) & 0xFFFF, nil
	}
	idx := s.ttfp.IdRangeOffset[seg]/2 + (value - s.ttfp.StartCount[seg]) - (segCount - seg)

	if s.ttfp.GlyphIdArray[int(idx)] == uint(0) {
		return 0, nil
	}

	return (s.ttfp.GlyphIdArray[int(idx)] + s.ttfp.IdDelta[seg]) & 0xFFFF, nil
}

//GlyphIndexToPdfWidth get with from glyphIndex
func (s *subsetFont) glyphIndexToPdfWidth(glyphIndex uint) uint {
	numberOfHMetrics := s.ttfp.NumberOfHMetrics()
	unitsPerEm := s.ttfp.UnitsPerEm()
	if glyphIndex >= numberOfHMetrics {
		glyphIndex = numberOfHMetrics - 1
	}

	width := s.ttfp.Widths()[glyphIndex]
	if unitsPerEm == 1000 {
		return width
	}
	return width * 1000 / unitsPerEm
}

//KernValueByLeft find kern value from kern table by left
func (s *subsetFont) kernValueByLeft(left uint) (bool, *font.KernValue) {

	k := s.ttfp.Kern()
	if k == nil {
		return false, nil
	}

	if kval, ok := k.Kerning[left]; ok {
		return true, &kval
	}

	return false, nil
}
