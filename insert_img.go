package pdfmix

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func insertImage(p *PdfData, imgr io.Reader, pageIndex int, x float64, y float64) error {
	img, err := ioutil.ReadAll(imgr)
	imgObjID, err := p.createImg(img)
	if err != nil {
		return errors.Wrap(err, "")
	}
	resnode, err := p.findResourcesOfPage(pageIndex)
	if err != nil {
		return errors.Wrapf(err, "findResourcesOfPage fail (pageIndex=%d)", pageIndex)
	}
	resnodes := p.objects[resnode.content.refTo]
	xobjnode, ok := resnodes.findOneNodeByName("XObject")
	if !ok {
		//FIXME: หาทางแก้ถ้า pdf ยังไม่มี XObject
		return errors.Wrapf(err, "not found XObject of pageindex=%d", pageIndex)
	}
	//fmt.Printf("--->%#v", xobjnode.content.refTo)
	imgIndex := 0
	imgprefix := "Im"
	xobjnodes := p.objects[xobjnode.content.refTo]
	if xobjnodes == nil {
		//FIXME: ถ้า xobjnodes เป็น nil
		tmps := pdfNodes{}
		p.objects[xobjnode.content.refTo] = &tmps
		xobjnodes = &tmps
	}

	for _, prop := range *xobjnodes {
		if prop.key.use == constNodeKeyUseName &&
			strings.Contains(prop.key.name, imgprefix) {
			index := getIndexFromImgPropName(imgprefix, prop.key.name)
			if index > imgIndex {
				imgIndex = index
			}
		}
	}
	imgIndex++ //use next index
	imgName := fmt.Sprintf("%s%d", imgprefix, imgIndex)
	xobjnodes.append(initNodeKeyUseNameAndNodeContentUseRefTo(imgName, imgObjID))
	//imgName = "Im4"
	cc := contenteCacheImg{
		imgName: imgName,
		x:       x,
		y:       y,
		h:       100,
		w:       100,
	}
	//*p.mapPageAndContentCachers[pageIndex] = append(*p.mapPageAndContentCachers[pageIndex], cc)
	p.mapPageAndContentCachers.append(pageIndex, &cc)
	return nil
}

func getIndexFromImgPropName(imgprefix string, propname string) int {
	idx := strings.Replace(propname, imgprefix, "", -1)
	i, err := strconv.Atoi(idx)
	if err != nil {
		return 0
	}
	return i
}
