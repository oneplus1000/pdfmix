package pdfmix

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/signintech/gopdf/fontmaker/core"
)

func (p *PdfData) appendSubsetFont(ssf *subsetFont, fontRef FontRef) (objectID, error) {

	ssfNodes := pdfNodes{}

	var ssfNodesObjectID = p.newRealID()
	p.objects[ssfNodesObjectID] = &ssfNodes

	typeNode := pdfNode{
		key: nodeKey{
			name: "Type",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/Font",
		},
	}

	subtypeNode := pdfNode{
		key: nodeKey{
			name: "Subtype",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/Type0",
		},
	}

	baseFontNode := pdfNode{
		key: nodeKey{
			name: "BaseFont",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/" + p.fontName(fontRef),
		},
	}

	encodingNode := pdfNode{
		key: nodeKey{
			name: "Encoding",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/Identity-H",
		},
	}

	descendantFontsNodeItemRefID := p.newFakeID()

	descendantFontsNode := pdfNode{
		key: nodeKey{
			name: "DescendantFonts",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: descendantFontsNodeItemRefID,
		},
	}

	toUnicodeRefID := p.newRealID()
	toUnicodeNodeRef := pdfNode{
		key: nodeKey{
			use:  constNodeKeyUseName,
			name: "ToUnicode",
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: toUnicodeRefID,
		},
	}

	ssfNodes.append(typeNode)
	ssfNodes.append(subtypeNode)
	ssfNodes.append(baseFontNode)
	ssfNodes.append(encodingNode)
	ssfNodes.append(descendantFontsNode)
	ssfNodes.append(toUnicodeNodeRef)

	//tounicode

	err := p.appendToUnicode(ssf, fontRef, toUnicodeRefID)
	if err != nil {
		return ssfNodesObjectID, errors.Wrap(err, "")
	}
	//DescendantFonts
	cidFontRefID := p.newRealID()

	descendantFontsItemNodes := pdfNodes{}
	p.objects[descendantFontsNodeItemRefID] = &descendantFontsItemNodes

	descendantFontsItem0Node := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: cidFontRefID,
		},
	}

	descendantFontsItemNodes.append(descendantFontsItem0Node)

	//CID Font
	err = p.appendCidFont(ssf, fontRef, cidFontRefID)
	if err != nil {
		return ssfNodesObjectID, errors.Wrap(err, "")
	}

	return ssfNodesObjectID, nil
}

func (p *PdfData) appendCidFont(
	ssf *subsetFont,
	fontRef FontRef,
	cidFontRefID objectID,
) error {
	cidFontNodes := pdfNodes{}
	p.objects[cidFontRefID] = &cidFontNodes

	cidtypeNode := pdfNode{
		key: nodeKey{
			name: "Type",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/Font",
		},
	}

	cidSubtypeNode := pdfNode{
		key: nodeKey{
			name: "Subtype",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/CIDFontType2",
		},
	}

	baseFontNode := pdfNode{
		key: nodeKey{
			name: "BaseFont",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/" + p.fontName(fontRef),
		},
	}

	cidSystemInfoNodeRefID := p.newFakeID()
	cidSystemInfoNode := pdfNode{
		key: nodeKey{
			name: "CIDSystemInfo",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: cidSystemInfoNodeRefID,
		},
	}

	wRefID := p.newFakeID()
	wNode := pdfNode{
		key: nodeKey{
			name: "W",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: wRefID,
		},
	}

	fontDescriptorRefID := p.newRealID()
	fontDescriptorNode := pdfNode{
		key: nodeKey{
			name: "FontDescriptor",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: fontDescriptorRefID,
		},
	}

	cidFontNodes.append(cidtypeNode)
	cidFontNodes.append(cidSubtypeNode)
	cidFontNodes.append(cidSystemInfoNode)
	cidFontNodes.append(wNode)
	cidFontNodes.append(fontDescriptorNode)
	cidFontNodes.append(baseFontNode)

	//fontDescriptor
	err := p.appendFontDescriptor(ssf, fontRef, fontDescriptorRefID)
	if err != nil {
		return errors.Wrap(err, "")
	}

	//w
	wNodes := pdfNodes{}
	p.objects[wRefID] = &wNodes

	for _, glyphIndex := range ssf.glyphIndexs {

		width := ssf.glyphIndexToPdfWidth(glyphIndex)

		wItemNode := pdfNode{
			key: nodeKey{
				use: constNodeKeyUseIndex,
			},
			content: nodeContent{
				use: constNodeContentUseString,
				str: fmt.Sprintf("%d[%d]", glyphIndex, width),
			},
		}
		wNodes.append(wItemNode)
	}

	//CID SystemInfo
	cidSystemInfoNodes := pdfNodes{}
	p.objects[cidSystemInfoNodeRefID] = &cidSystemInfoNodes
	orderingNode := pdfNode{
		key: nodeKey{
			name: "Ordering",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "(Identity)",
		},
	}

	registryNode := pdfNode{
		key: nodeKey{
			name: "Registry",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "(Adobe)",
		},
	}

	supplementNode := pdfNode{
		key: nodeKey{
			name: "Supplement",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "0",
		},
	}

	cidSystemInfoNodes.append(orderingNode)
	cidSystemInfoNodes.append(registryNode)
	cidSystemInfoNodes.append(supplementNode)
	return nil
}

func (p *PdfData) appendFontDescriptor(
	ssf *subsetFont,
	fontRef FontRef,
	fontDescriptorRefID objectID,
) error {

	fontDescriptorNodes := pdfNodes{}
	p.objects[fontDescriptorRefID] = &fontDescriptorNodes

	typeNode := pdfNode{
		key: nodeKey{
			name: "Type",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/FontDescriptor",
		},
	}

	ascentNode := pdfNode{
		key: nodeKey{
			name: "Ascent",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.Ascender(), ssf.ttfp.UnitsPerEm())),
		},
	}

	capHeightNode := pdfNode{
		key: nodeKey{
			name: "CapHeight",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.CapHeight(), ssf.ttfp.UnitsPerEm())),
		},
	}

	flagsNode := pdfNode{
		key: nodeKey{
			name: "Flags",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", ssf.ttfp.Flag()),
		},
	}

	fontBoxNodeItemRefID := p.newFakeID()
	fontBoxNode := pdfNode{
		key: nodeKey{
			name: "FontBBox",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: fontBoxNodeItemRefID,
		},
	}

	fontNameNode := pdfNode{
		key: nodeKey{
			name: "FontName",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "/" + p.fontName(fontRef),
		},
	}

	italicAngleNode := pdfNode{
		key: nodeKey{
			name: "ItalicAngle",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", ssf.ttfp.ItalicAngle()),
		},
	}

	stemVNode := pdfNode{
		key: nodeKey{
			name: "ItalicAngle",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: "0",
		},
	}

	xHeightNode := pdfNode{
		key: nodeKey{
			name: "XHeight",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.XHeight(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontFile2RefID := p.newRealID()
	fontFile2RefNode := pdfNode{
		key: nodeKey{
			name: "FontFile2",
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: fontFile2RefID,
		},
	}

	err := p.appendFontFile2(ssf, fontRef, fontFile2RefID) //fontfile2
	if err != nil {
		return errors.Wrap(err, "")
	}

	fontDescriptorNodes.append(typeNode)
	fontDescriptorNodes.append(ascentNode)
	fontDescriptorNodes.append(capHeightNode)
	fontDescriptorNodes.append(flagsNode)
	fontDescriptorNodes.append(fontBoxNode)
	fontDescriptorNodes.append(fontNameNode)
	fontDescriptorNodes.append(italicAngleNode)
	fontDescriptorNodes.append(stemVNode)
	fontDescriptorNodes.append(xHeightNode)
	fontDescriptorNodes.append(fontFile2RefNode)

	//fontbox
	fontBoxNodeItemNodes := pdfNodes{}
	p.objects[fontBoxNodeItemRefID] = &fontBoxNodeItemNodes

	fontBoxItemXMinNode := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.XMin(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxItemYMinNode := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.YMin(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxItemXMaxNode := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.XMax(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxItemYMaxNode := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", toPdfUnit(ssf.ttfp.YMax(), ssf.ttfp.UnitsPerEm())),
		},
	}

	fontBoxNodeItemNodes.append(fontBoxItemXMinNode)
	fontBoxNodeItemNodes.append(fontBoxItemYMinNode)
	fontBoxNodeItemNodes.append(fontBoxItemXMaxNode)
	fontBoxNodeItemNodes.append(fontBoxItemYMaxNode)

	return nil
}

//convert unit
func toPdfUnit(val int, unitsPerEm uint) int {
	return core.Round(float64(float64(val) * 1000.00 / float64(unitsPerEm)))
}

func (p PdfData) fontName(f FontRef) string {
	return string(f)
}
