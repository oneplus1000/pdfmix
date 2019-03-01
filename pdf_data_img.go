package pdfmix

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (p *PdfData) createImg(img []byte) (objectID, error) {
	info, err := parseImg(bytes.NewReader(img))
	if err != nil {
		return objectID{}, errors.Wrap(err, "")
	}
	imgNodes := pdfNodes{}
	p.buildImgProp(&imgNodes, info)
	length := initNodeKeyUseNameAndNodeContentUseString("Length", fmt.Sprintf("%d", len(info.data)))
	stm := initNodeKeyUseStreamNodeContentUseStream(info.data)
	imgNodes.append(length)
	imgNodes.append(stm)
	imgNodesID := p.newRealID()
	p.objects[imgNodesID] = &imgNodes
	return imgNodesID, nil
}

func (p *PdfData) buildImgProp(nodes *pdfNodes, imginfo imgInfo) error {
	typ := initNodeKeyUseNameAndNodeContentUseString("Type", "/XObject")
	subType := initNodeKeyUseNameAndNodeContentUseString("Subtype", "/Image")
	width := initNodeKeyUseNameAndNodeContentUseString("Width", fmt.Sprintf("%d", imginfo.w))
	height := initNodeKeyUseNameAndNodeContentUseString("Height", fmt.Sprintf("%d", imginfo.h))

	nodes.append(typ)
	nodes.append(subType)
	nodes.append(width)
	nodes.append(height)

	if isColspaceIndexed(imginfo) {
		size := len(imginfo.pal)/3 - 1
		rgbID, err := p.createDeviceRGB(imginfo)
		if err != nil {
			return errors.Wrap(err, "")
		}
		colorSpace := initNodeKeyUseNameAndNodeContentUseString("ColorSpace", fmt.Sprintf("[/Indexed /DeviceRGB %d %d 0 R]\n", size, rgbID.id))
		nodes.append(colorSpace)
	} else {
		colorSpace := initNodeKeyUseNameAndNodeContentUseString("ColorSpace", fmt.Sprintf("/%s", imginfo.colspace))
		nodes.append(colorSpace)
		if imginfo.colspace == "DeviceCMYK" {
			decode := initNodeKeyUseNameAndNodeContentUseString("Decode", "[1 0 1 0 1 0 1 0]")
			nodes.append(decode)
		}
	}

	bitsPerComponent := initNodeKeyUseNameAndNodeContentUseString("BitsPerComponent", fmt.Sprintf("%s", imginfo.bitsPerComponent))
	nodes.append(bitsPerComponent)
	if strings.TrimSpace(imginfo.filter) != "" {
		filter := initNodeKeyUseNameAndNodeContentUseString("Filter", fmt.Sprintf("/%s", imginfo.filter))
		nodes.append(filter)
	}

	if strings.TrimSpace(imginfo.decodeParms) != "" {
		decodeParms := initNodeKeyUseNameAndNodeContentUseString("DecodeParms", fmt.Sprintf("<<%s>>", imginfo.decodeParms))
		nodes.append(decodeParms)
		//buffer.WriteString(fmt.Sprintf("/DecodeParms <<%s>>\n", imginfo.decodeParms))
	}

	if imginfo.trns != nil && len(imginfo.trns) > 0 {
		j := 0
		max := len(imginfo.trns)
		var trns bytes.Buffer
		for j < max {
			trns.WriteString(fmt.Sprintf("%d", imginfo.trns[j]))
			trns.WriteString(" ")
			trns.WriteString(fmt.Sprintf("%d", imginfo.trns[j]))
			trns.WriteString(" ")
			j++
		}
		mark := initNodeKeyUseNameAndNodeContentUseString("/Mask", trns.String())
		nodes.append(mark)
	}

	if haveSMask(imginfo) {
		//FIXME smark
		//buffer.WriteString(fmt.Sprintf("/SMask %d 0 R\n", imginfo.smarkObjID+1))
	}

	return nil
}

func (p *PdfData) createDeviceRGB(imginfo imgInfo) (objectID, error) {
	rgbNodes := pdfNodes{}
	length := initNodeKeyUseNameAndNodeContentUseString("Length", strconv.Itoa(len(imginfo.pal)))
	stm := initNodeKeyUseStreamNodeContentUseStream(imginfo.pal)
	rgbNodes.append(length)
	rgbNodes.append(stm)
	id := p.newRealID()
	p.objects[id] = &rgbNodes
	return id, nil
}

func (p *PdfData) createSMask(imginfo imgInfo) (objectID, error) {

	return objectID{}, nil
}
