package pdfmix

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

type contenteCacheText struct {
	ssf      *subsetFont
	textRaw  string
	x, y     int
	fontSize int
}

func (c *contenteCacheText) build(w io.Writer) (int64, error) {

	var buffText bytes.Buffer
	var leftRune rune
	var leftRuneIndex uint
	unitsPerEm := int(c.ssf.ttfp.UnitsPerEm())
	for _, currRune := range c.textRaw {

		currRuneIndex, err := c.ssf.getGlyphIndex(currRune)
		if err != nil {
			return 0, errors.Wrapf(err, "c.ssf.getGlyphIndex(%s) fail", string(currRune))
		}

		//kerning
		pairval := convertTTFUnit2PDFUnit(int(c.kerning(leftRune, currRune, leftRuneIndex, currRuneIndex)), unitsPerEm)
		if pairval != 0 {
			buffText.WriteString(fmt.Sprintf(">%d<", (-1)*pairval))
		}

		//write rune index
		buffText.WriteString(fmt.Sprintf("%04X", currRuneIndex))

		//next loop
		leftRune = currRune
		leftRuneIndex = currRuneIndex
	}

	x := 10.0           //FIXME: this hard code
	y := 800.00         //FIXME: this hard code
	fontCountIndex := 2 //FIXME: this hard code
	fontSize := 14      //FIXME: this hard code

	var buff bytes.Buffer
	buff.WriteString("BT\n")
	buff.WriteString(fmt.Sprintf("%0.2f %0.2f TD\n", x, y))
	buff.WriteString("/F" + strconv.Itoa(fontCountIndex) + " " + strconv.Itoa(fontSize) + " Tf\n")
	buff.WriteString("[<")
	buffText.WriteTo(&buff)
	buff.WriteString(">] TJ\n")
	buff.WriteString("ET\n")

	//fmt.Printf("%s\n", buff.String()) //debug

	return buff.WriteTo(w)
}

func (c *contenteCacheText) kerning(leftRune rune, rightRune rune, leftIndex uint, rightIndex uint) int16 {

	pairVal := int16(0)
	if haveKerning, kval := c.ssf.kernValueByLeft(leftIndex); haveKerning {
		if ok, v := kval.ValueByRight(rightIndex); ok {
			pairVal = v
		}
	}
	/*
		if f.funcKernOverride != nil {
			pairVal = f.funcKernOverride(
			leftRune,
			rightRune,
			leftIndex,
			rightIndex,
			pairVal,
			)
		}
	*/
	return pairVal
}

func convertTTFUnit2PDFUnit(n int, upem int) int {
	var ret int
	if n < 0 {
		rest1 := n % upem
		storrest := 1000 * rest1
		//ledd2 := (storrest != 0 ? rest1 / storrest : 0);
		ledd2 := 0
		if storrest != 0 {
			ledd2 = rest1 / storrest
		} else {
			ledd2 = 0
		}
		ret = -((-1000*n)/upem - int(ledd2))
	} else {
		ret = (n/upem)*1000 + ((n%upem)*1000)/upem
	}
	return ret
}
