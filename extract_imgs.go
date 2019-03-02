package pdfmix

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

func extractImgs(p *PdfData) error {
	pages, err := p.findAllPage()
	if err != nil {
		return errors.Wrapf(err, "")
	}
	_ = pages

	//fmt.Printf("size:%d", len(pages))

	for i, page := range pages {
		fmt.Printf("%d\n", i)
		err := extractImg(p, page, i)
		if err != nil {
			return errors.Wrapf(err, "")
		}
		//break //just for test
	}

	return nil
}

func extractImg(p *PdfData, objID objectID, pageIndex int) error {
	//fmt.Printf("%+v\n\n", p.objects[objID])

	page, ok := p.objects[objID]
	if !ok {
		return fmt.Errorf("Page not found")
	}

	//fmt.Printf("%+v\n\n", page)

	resRef, ok := page.findOneNodeByName("Resources")
	if !ok {
		//fmt.Printf("xobj=%+v\n\n", xobj)
		return fmt.Errorf("/Resources not found in /Page")
	}

	res, ok := p.objects[resRef.content.refTo]
	if !ok {
		return fmt.Errorf("Resources not found")
	}

	//_ = res
	//fmt.Printf("%+v\n\n", res)

	xobjRef, ok := res.findOneNodeByName("XObject")
	if !ok {
		return fmt.Errorf("/XObject not found in /Resources")
	}

	_ = xobjRef
	xobj, ok := p.objects[xobjRef.content.refTo]
	if !ok {
		return fmt.Errorf("XObject not found")
	}

	//_ = xobj
	//fmt.Printf("%+v\n\n", xobj)

	for _, childRef := range *xobj {
		//fmt.Printf("%+v\n", childRef)
		child, ok := p.objects[childRef.content.refTo]
		if !ok {
			return fmt.Errorf("%+v not found", childRef.key)
		}
		if isImage(child) { //เอาเฉพาะที่เป็นรูป

			imgType, ok := imageType(child)
			if !ok {
				return fmt.Errorf("Filter not found")
			}

			if imgType == "png" {
				pic, _ := readStream(child)
				err := ioutil.WriteFile(fmt.Sprintf("testing/out/x_%d.png", pageIndex), pic, 0777)
				if err != nil {
					return errors.Wrapf(err, "")
				}
			} else if imgType == "jpg" {
				pic, _ := readStream(child)
				err := ioutil.WriteFile(fmt.Sprintf("testing/out/x_%d.jpg", pageIndex), pic, 0777)
				if err != nil {
					return errors.Wrapf(err, "")
				}
			} else if imgType == "tif" {
				//pic, _ := readStream(child)

			}
		}
		//fmt.Printf("%+v\n\n", child)
	}

	return nil
}

func imageType(obj *pdfNodes) (string, bool) {
	for _, c := range *obj {
		if c.key.name == "Filter" {
			if c.content.str == "/CCITTFaxDecode" {
				return "tif", true
			} else if c.content.str == "/DCTDecode" {
				return "jpg", true
			} else {
				return "png", true
			}
		}
	}
	return "", false
}

func isImage(obj *pdfNodes) bool {
	for _, c := range *obj {
		if c.key.name == "Subtype" && c.content.str == "/Image" {
			return true
		}
	}
	return false
}

func readStream(obj *pdfNodes) ([]byte, int) {
	for _, c := range *obj {
		if c.key.use == 3 && len(c.content.stream) > 0 { //เอา stream ออกมา
			return c.content.stream, len(c.content.stream)
		}
	}
	return nil, 0
}

//https://github.com/gettalong/hexapdf/blob/e0b7f345e9474e2bf07a920b67b41007f095ab37/lib/hexapdf/type/image.rb
//https://code.i-harness.com/en/q/284f6a
//https://github.com/hhrutter/pdfcpu
