package pdfmix

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

func debugWriteTxt(name string, buff *bytes.Buffer) {
	ioutil.WriteFile(name, buff.Bytes(), 0777)
}

func (p *PdfData) moveContent(pageIndex int, x, y float64) error {
	//find nodes of page
	contentID, err := p.findContentIDByPageIndex(pageIndex)
	if err != nil {
		return errors.Wrapf(err, "p.findContentIDByPageIndex(...) fail")
	}

	//find stream in content
	var buffOld bytes.Buffer
	isZip, streamIndex, err := p.getStream(contentID, &buffOld)
	if err != nil {
		return errors.Wrapf(err, "p.getStream(...) fail")
	}

	debugWriteTxt("./testing/out/__old.txt", &buffOld)

	_ = isZip
	_ = streamIndex

	var lineParsers []lineParser

	var buffNew bytes.Buffer
	scr := bufio.NewScanner(&buffOld)
	scr.Split(bufio.ScanLines)
	lineIdx := 0
	for scr.Scan() {
		lc, err := newLineCache(scr.Text(), lineIdx)
		if err != nil {
			return errors.Wrap(err, "")
		}
		for _, lpr := range lineParsers {
			ok, err := lpr.parse(&lc, &buffNew)
			if err != nil {
				return errors.Wrap(err, "")
			} else if ok {
				break //ok ละจบ
			}
		}
		lineIdx++
	}

	return nil
}

type lineParser interface {
	parse(lc *lineCache, w io.Writer) (bool, error)
}

type lineCache struct {
	lineIdx int
	tokens  []string
}

func newLineCache(line string, lineIdx int) (lineCache, error) {
	tmp := strings.TrimSpace(line)
	return lineCache{
		tokens:  strings.Split(tmp, " "),
		lineIdx: lineIdx,
	}, nil
}

type currState struct {
}

func (p *PdfData) findContentIDByPageIndex(pageIndex int) (objectID, error) {
	page, err := p.findRefPageNodeByPageIndex(pageIndex)
	if err != nil {
		return objectIDEmpty, errors.Wrapf(err, "not found page")
	}
	nodesOfPage := p.objects[page.content.refTo]
	var contentID objectID
	var foundContentID = false
	for _, n := range *nodesOfPage {
		if n.key.use == constNodeKeyUseName &&
			n.key.name == "Contents" {
			contentID = n.content.refTo
			foundContentID = true
			break
		}
	}
	if !foundContentID {
		return objectIDEmpty, errors.New("Contents not found")
	}
	return contentID, nil
}

func (p *PdfData) getStream(id objectID, w io.Writer) (bool, int, error) {
	var stream []byte
	streamIndex := 0
	streamFound := false
	isZip := false
	nodes := p.objects[id]
	c := 0
	for i, n := range *nodes {
		if n.key.use == constNodeKeyUseName &&
			n.key.name == "Filter" &&
			n.content.use == constNodeContentUseString &&
			n.content.str == "/FlateDecode" {
			isZip = true
			c++
		} else if n.content.use == constNodeContentUseStream {
			stream = n.content.stream
			streamIndex = i
			streamFound = true
			c++
		}
		if c >= 2 {
			break
		}
	}

	if !streamFound {
		return false, 0, errors.New("not found stream")
	}

	buff := bytes.NewBuffer(stream)
	if isZip {
		reader, err := zlib.NewReader(buff)
		if err != nil {
			return false, 0, errors.Wrap(err, "zlib.NewReader fail")
		}
		defer reader.Close()
		_, err = io.Copy(w, reader)
		if err != nil {
			return false, 0, errors.Wrap(err, "io.Copy(...) fail")
		}
	} else {
		_, err := io.Copy(w, buff)
		if err != nil {
			return false, 0, errors.Wrap(err, "io.Copy(...) fail")
		}
	}
	return isZip, streamIndex, nil
}
