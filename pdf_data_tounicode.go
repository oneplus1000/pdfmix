package pdfmix

import (
	"bytes"
	"fmt"
	"strconv"
)

func (p *PdfData) appendToUnicode(
	ssf *subsetFont,
	fontRef FontRef,
	toUnicodeRefID objectID,
) error {

	prefix :=
		"/CIDInit /ProcSet findresource begin\n" +
			"12 dict begin\n" +
			"begincmap\n" +
			"/CIDSystemInfo << /Registry (Adobe)/Ordering (UCS)/Supplement 0>> def\n" +
			"/CMapName /Adobe-Identity-UCS def /CMapType 2 def\n"
	suffix := "endcmap CMapName currentdict /CMap defineresource pop end end"

	characterToGlyphIndex := ssf.glyphIndexs

	glyphIndexToCharacter := make(map[int]rune)
	lowIndex := 65536
	hiIndex := -1
	for k, v := range characterToGlyphIndex {
		index := int(v)
		if index < lowIndex {
			lowIndex = index
		}
		if index > hiIndex {
			hiIndex = index
		}
		glyphIndexToCharacter[index] = k
	}

	var buff bytes.Buffer
	buff.WriteString(prefix)
	buff.WriteString("1 begincodespacerange\n")
	buff.WriteString(fmt.Sprintf("<%04X><%04X>\n", lowIndex, hiIndex))
	buff.WriteString("endcodespacerange\n")
	buff.WriteString(fmt.Sprintf("%d beginbfrange\n", len(glyphIndexToCharacter)))
	for k, v := range glyphIndexToCharacter {
		buff.WriteString(fmt.Sprintf("<%04X><%04X><%04X>\n", k, k, v))
	}
	buff.WriteString("endbfrange\n")
	buff.WriteString(suffix)
	buff.WriteString("\n")

	toUnicodeNodes := pdfNodes{}
	p.objects[toUnicodeRefID] = &toUnicodeNodes

	lengthNode := pdfNode{
		key: nodeKey{
			name: "Length",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: strconv.Itoa(buff.Len()),
		},
	}

	streamNode := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseStream,
		},
		content: nodeContent{
			use:    constNodeContentUseStream,
			stream: buff.Bytes(),
		},
	}

	toUnicodeNodes.append(lengthNode)
	toUnicodeNodes.append(streamNode)

	return nil
}
