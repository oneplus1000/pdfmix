package pdfmix

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/pkg/errors"
)

//PdfData hold data of pdf file
type PdfData struct {
	subsetFonts              map[FontRef](*subsetFont)
	mapPageAndContentCachers pageContentCacherMap
	objects                  map[objectID]*pdfNodes
	bytesCallBack            func(buff *bytes.Buffer, currRealObjID objectID, status int)
	buildinfo                buildInfo
	maxRealID, maxFakeID     uint32
	pwdProtectionInfo        *pwdProtectionInfo
	ds                       DigitalSigner
}

const statusStartObj = 1
const statusEndObj = 2
const statusEndOfFile = 3

func newPdfData() *PdfData {
	var p PdfData
	p.objects = make(map[objectID]*pdfNodes)
	return &p
}

func (p *PdfData) push(myID objectID, node pdfNode) {
	if _, ok := p.objects[myID]; ok {
		p.objects[myID].append(node)
	} else {
		var nodes pdfNodes
		nodes.append(node)
		p.objects[myID] = &nodes
	}
}

//findAllPage หา page ทั้งหมดใน pdf จะ return map[index ของ page]objectID ของ page
func (p *PdfData) findAllPage() (map[int]objectID, error) {

	cats, err := newQuery(p).findDict("Type", "/Catalog")
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	if len(cats) <= 0 {
		return nil, ErrDictNotFound
	}

	cat := cats[0]
	catnodes := p.objects[cat.objID]
	rootpage, ok := catnodes.findOneNodeByName("Pages")
	if !ok {
		return nil, fmt.Errorf("obj %+v not found /Pages", cat.objID)
	} else if rootpage.content.use != constNodeContentUseRefTo {
		return nil, fmt.Errorf("obj %+v have bad /Pages", cat.objID)
	}

	pageIndexs := make(map[int]objectID)
	_, err = p.findAllSubPage(rootpage.content.refTo, 0, &pageIndexs)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return pageIndexs, nil
}

//findAllSubPage ใช้หา page จาก findAllPage แบบ recursive
func (p *PdfData) findAllSubPage(typePageID objectID, offset int, pageIndexs *map[int]objectID) (int, error) {

	obj := p.objects[typePageID]
	//check type
	ty, ok := obj.findOneNodeByName("Type")
	if !ok {
		return 0, fmt.Errorf("obj %+v not found /Type", typePageID)
	}
	if ty.content.use == constNodeContentUseString && ty.content.str == "/Page" {
		//เป็น /page
		(*pageIndexs)[offset] = typePageID
		return offset + 1, nil
	}

	kid, ok := obj.findOneNodeByName("Kids")
	if !ok {
		return 0, fmt.Errorf("obj %+v not found /Kids", typePageID)
	}
	kidsNodes := p.objects[kid.content.refTo]
	for _, kid := range *kidsNodes {
		of, err := p.findAllSubPage(kid.content.refTo, offset, pageIndexs)
		if err != nil {
			return 0, errors.Wrap(err, "")
		}
		offset = of
	}
	return offset, nil
}

//build build pdf
func (p *PdfData) build() error {

	//find all ref
	pageIndexs, err := p.findAllPage()
	if err != nil {
		return errors.Wrap(err, "findAllPage() fail")
	}
	resObjectIDs := make(map[int]objectID)
	contentObjectIDs := make(map[int]objectID)
	for i, kidObjectID := range pageIndexs {

		resNode, err := newQuery(p).findPdfNodeByKeyName(kidObjectID, "Resources")
		if err != nil {
			return errors.Wrapf(err, "newQuery(p).findPdfNodeByKeyName(%+v, \"Resources\")", kidObjectID)
		}
		resObjectIDs[i] = resNode.content.refTo

		contentNode, err := newQuery(p).findPdfNodeByKeyName(kidObjectID, "Contents")
		if err != nil {
			return errors.Wrap(err, "")
		}
		contentObjectIDs[i] = contentNode.content.refTo
	}
	//end find all ref

	err = p.buildSubsetFont(resObjectIDs)
	if err != nil {
		return errors.Wrap(err, "buildSubsetFont(...) fail")
	}

	err = p.buildContent(contentObjectIDs)
	if err != nil {
		return errors.Wrap(err, "buildContent(...) fail")
	}

	return nil
}

func (p *PdfData) buildSubsetFont(resObjectIDs map[int]objectID) error {

	var err error
	var newFontObjectIDs []objectID
	for fontRef, ss := range p.subsetFonts {
		var newFontObjectID objectID
		newFontObjectID, err = p.appendSubsetFont(ss, fontRef)
		if err != nil {
			return errors.Wrap(err, "")
		}
		newFontObjectIDs = append(newFontObjectIDs, newFontObjectID)
	}

	//append subset font to all res
	resIDs := make(map[objectID]bool)
	for _, objectID := range resObjectIDs {
		if _, ok := resIDs[objectID]; !ok {
			resIDs[objectID] = true
		}
	}

	for resID := range resIDs {
		fontNode, err := newQuery(p).findPdfNodeByKeyName(resID, "Font")
		if err == ErrKeyNameNotFound {
			continue
		} else if err != nil {
			return errors.Wrap(err, "")
		}
		fontNodes := p.objects[fontNode.content.refTo]
		fName := "F"
		fIndexMax := 0
		for _, node := range *fontNodes {
			fIndex := 0
			fName, fIndex, err = p.fontnameExtract(node.key.name)
			if err != nil {
				return errors.Wrap(err, "")
			}
			if fIndex > fIndexMax {
				fIndexMax = fIndex
			}
		}

		for i, newFontObjectID := range newFontObjectIDs {
			fontNode := pdfNode{
				key: nodeKey{
					name: fmt.Sprintf("%s%d", fName, fIndexMax+1+i),
					use:  constNodeKeyUseName,
				},
				content: nodeContent{
					use:   constNodeContentUseRefTo,
					refTo: newFontObjectID,
				},
			}
			fontNodes.append(fontNode)
		}
	}

	return nil
}

func (p *PdfData) fontnameExtract(fontname string) (string, int, error) {

	/*
		rex := regexp.MustCompile("[0-9]+")
		if !rex.MatchString(fontname) {
			return "", 0, fmt.Errorf("can not parse %s", fontname)
		}

		fname := rex.ReplaceAllString(fontname, "")

		findex, err := strconv.Atoi(strings.Replace(fontname, fname, "", -1))
		if err != nil {
			return "", 0, errors.Wrap(err, "")
		}

		return fname, findex, nil
	*/
	offset := -1
	rs := []rune(fontname)
	max := len(rs)
	for i := max - 1; i >= 0; i-- {
		c := string(rs[i])
		if _, err := strconv.Atoi(c); err != nil {
			offset = i
			break
		}
	}

	if offset == -1 {
		//fontname is number
		findex, err := strconv.Atoi(fontname)
		if err != nil {
			return "", 0, errors.Wrap(err, "")
		}
		return "", findex, nil
	}

	fName := fontname[0 : offset+1]
	fIndexStr := fontname[offset+1:]
	findex, err := strconv.Atoi(fIndexStr)
	if err != nil {
		return "", 0, errors.Wrap(err, "")
	}

	return fName, findex, nil
}

func (p *PdfData) buildContent(contentObjectIDs map[int]objectID) error {

	mapPageAndBuff := make(map[int]*bytes.Buffer) //map ระหว่าง pageindex กับ buffer ( ของ content)
	//--
	pageindexs := p.mapPageAndContentCachers.allPageIndexs()
	for _, pageIndex := range pageindexs {
		if _, ok := mapPageAndBuff[pageIndex]; !ok {
			mapPageAndBuff[pageIndex] = &bytes.Buffer{}
		}
		ccs := p.mapPageAndContentCachers.contentCachersByPageIndex(pageIndex)
		for _, c := range ccs {
			_, err := c.build(mapPageAndBuff[pageIndex])
			if err != nil {
				return errors.Wrap(err, "")
			}
		}
	}
	/*
		for pageIndex, caches := range p.mapPageAndContentCachers {
			if _, ok := mapPageAndBuff[pageIndex]; !ok {
				mapPageAndBuff[pageIndex] = &bytes.Buffer{}
			}
			for _, cache := range *caches {
				_, err := cache.build(mapPageAndBuff[pageIndex])
				if err != nil {
					return errors.Wrap(err, "")
				}
			}
		}*/
	//--

	for pageIndex, buff := range mapPageAndBuff {
		stm, err := p.getStreamOfContentOfPage(contentObjectIDs, pageIndex) //ดึงข้อมูลเดิมออกมา
		if err != nil {
			return errors.Wrapf(err, "p.getStreamOfContentOfPage(contentObjectIDs, %d) fail", pageIndex)
		}
		stm.Write(buff.Bytes())                                   //เพิ่มข้อมูลเข้าไป
		err = p.replaceStramObj(contentObjectIDs[pageIndex], stm) // ยัดลง pdfdata
		if err != nil {
			return errors.Wrap(err, "p.reWriteStramObj(contentObjectIDs[pageIndex], stm) fail")
		}
	}

	return nil
}

func (p *PdfData) replaceStramObj(id objectID, data *bytes.Buffer) error {

	nodes := p.objects[id]
	isStream := false
	for _, node := range *nodes {
		if node.content.use == constNodeContentUseStream {
			isStream = true
			break
		}
	}

	if !isStream {
		return ErrStreamNotFound
	}
	newNodes := pdfNodes{}
	newNodeLen := pdfNode{
		key: nodeKey{
			use:  constNodeKeyUseName,
			name: "Length",
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: fmt.Sprintf("%d", data.Len()),
		},
	}
	newNodeStm := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseStream,
		},
		content: nodeContent{
			use:    constNodeContentUseStream,
			stream: data.Bytes(),
		},
	}
	newNodes.append(newNodeLen)
	newNodes.append(newNodeStm)
	p.objects[id] = &newNodes

	return nil
}

func (p *PdfData) getStreamOfContentOfPage(contentObjectIDs map[int]objectID, pageIndex int) (*bytes.Buffer, error) {
	var nodes *pdfNodes
	filter := ""
	var stm []byte
	if id, ok := contentObjectIDs[pageIndex]; ok {
		nodes = p.objects[id]
		for _, node := range *nodes {
			if node.key.name == "Filter" {
				filter = node.content.str
			} else if node.content.use == constNodeContentUseStream {
				stm = node.content.stream
			}
		}
	}

	var buff *bytes.Buffer
	if filter == "/FlateDecode" { //zip
		buffZip := bytes.NewBuffer(stm)
		r, err := zlib.NewReader(buffZip)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		defer r.Close()
		buff = bytes.NewBuffer(nil)
		_, err = io.Copy(buff, r)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	} else {
		buff = bytes.NewBuffer(stm)
	}
	return buff, nil
}

//bytes return []byte of pdf file
func (p *PdfData) bytes() ([]byte, error) {
	var buff bytes.Buffer
	var buffTrailer bytes.Buffer
	var realIDs []int
	for objID := range p.objects {
		if objID.isReal {
			realIDs = append(realIDs, int(objID.id))
		}
	}
	sort.Ints(realIDs)
	buff.WriteString("%PDF-1.7")
	xreftable := make(map[int]int) //map[realId] offfset
	for _, realID := range realIDs {
		realObjID := createRealObjectID(uint32(realID))
		if p.bytesCallBack != nil {
			p.bytesCallBack(&buff, realObjID, statusStartObj)
		}
		if realID == 0 { //Root

			sizeIdx, err := newQuery(p).findIndexByKeyName(realObjID, "Size")
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			p.objects[realObjID].remove(sizeIdx)
			p.objects[realObjID].append(pdfNode{
				key: nodeKey{
					name: "Size",
					use:  constNodeKeyUseName,
				},
				content: nodeContent{
					use: constNodeContentUseString,
					str: fmt.Sprintf("%d", len(realIDs)),
				},
			})

			data, err := p.bytesOfNodesByID(realObjID)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			buffTrailer.Write(data)

		} else { //Other

			buff.WriteString("\n")
			//xreftable = append(xreftable, buff.Len())
			xreftable[realID] = buff.Len()
			buff.WriteString(fmt.Sprintf("%d 0 obj\n", realID))
			data, err := p.bytesOfNodesByID(realObjID)
			if err != nil {
				return nil, errors.Wrap(err, "")
			}
			buff.Write(data)
			buff.WriteString("\nendobj\n")

		}
		if p.bytesCallBack != nil {
			p.bytesCallBack(&buff, realObjID, statusEndObj)
		}
	}
	startxref := buff.Len()
	buff.WriteString("\nxref\n")
	buff.Write(p.bytesOfXref(xreftable))
	buff.WriteString("trailer")
	buffTrailer.WriteTo(&buff)
	buff.WriteString("\nstartxref\n")
	buff.WriteString(fmt.Sprintf("%d", startxref))
	buff.WriteString("\n%%EOF\n")
	if p.bytesCallBack != nil {
		p.bytesCallBack(&buff, objectID{}, statusEndOfFile)
	}

	return buff.Bytes(), nil
}

type xrefrow struct {
	offset int
	gen    string
	flag   string
}

//bytesOfXref ( xreftable -> map[realId] offfset)
func (p PdfData) bytesOfXref(xreftable map[int]int) []byte {

	size := 1
	for id := range xreftable {
		if id > size {
			size = id
		}
	}
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("0 %d\r\n", size+1))
	var xrefrows []xrefrow
	xrefrows = append(xrefrows, xrefrow{offset: 0, flag: "f", gen: "65535"})
	lastIndexOfF := 0
	j := 1
	//fmt.Printf("size:%d\n", size)
	for j <= size {
		if linelen, ok := xreftable[j]; ok {
			xrefrows = append(xrefrows, xrefrow{offset: linelen, flag: "n", gen: "00000"})
		} else {
			xrefrows = append(xrefrows, xrefrow{offset: 0, flag: "f", gen: "65535"})
			offset := len(xrefrows) - 1
			xrefrows[lastIndexOfF].offset = offset
			lastIndexOfF = offset
		}
		j++
	}

	for _, xrefrow := range xrefrows {
		buff.WriteString(formatXrefline(xrefrow.offset) + " " + xrefrow.gen + " " + xrefrow.flag + " \n")
	}

	return buff.Bytes()
}

func formatXrefline(n int) string {
	str := strconv.Itoa(n)
	for len(str) < 10 {
		str = "0" + str
	}
	return str
}

func (p PdfData) bytesOfNodesByID(id objectID) ([]byte, error) {

	var buff bytes.Buffer
	nodes := p.objects[id]
	isArray := p.isArrayNodes(nodes)
	indexOfStream, isStream := p.isStream(nodes)
	isSingleValObj := p.isSingleValObjNodes(nodes)
	if isArray {
		buff.WriteString("[")
	} else if isSingleValObj {
		buff.WriteString("\n")
	} else {
		buff.WriteString("\n<<\n")
	}

	if nodes != nil {

		for _, node := range *nodes {
			/*
				if id.id == 37 {
					fmt.Printf("%+v -> %+v\n", node.key, node.content.str)
				}*/
			//key
			if node.key.use == constNodeKeyUseName {
				buff.WriteString(fmt.Sprintf("/%s", node.key.name))
			}
			//content
			buff.WriteString(" ")
			if node.content.use == constNodeContentUseString || node.content.use == constNodeContentUseSingleObj {
				buff.WriteString(fmt.Sprintf("%s", node.content.str))
			} else if node.content.use == constNodeContentUseRefTo {
				if node.content.refTo.isReal {
					buff.WriteString(fmt.Sprintf("%d 0 R", node.content.refTo.id))
				} else {
					data, err := p.bytesOfNodesByID(node.content.refTo)
					if err != nil {
						return nil, errors.Wrap(err, "")
					}
					buff.Write(data)
				}
			} else if node.content.use == constNodeContentUseNull {
				buff.WriteString("<<>>")
			}

			if !isArray && !isSingleValObj {
				buff.WriteString("\n")
			}
		}

	} //end nodes != nil

	if isArray {
		buff.WriteString(" ]")
	} else if isSingleValObj {
		buff.WriteString("")
	} else {
		buff.WriteString(">>")
	}

	if isStream && indexOfStream != -1 {
		err := p.writeStream(nodes, id, indexOfStream, &buff)
		if err != nil {
			return nil, errors.Wrap(err, "writeStream(...) fail")
		}
	}

	return buff.Bytes(), nil
}

func (p PdfData) writeStream(nodes *pdfNodes, id objectID, indexOfStream int, buff *bytes.Buffer) error {

	sm := (*nodes)[indexOfStream].content.stream
	if p.pwdProtectionInfo != nil {
		objKey := createPwdProtectionObjectKey(id.id, p.pwdProtectionInfo.encryptionKey)
		cip, err := rc4Cip(objKey, sm)
		if err != nil {
			return errors.Wrap(err, "rc4Cip(...) fail")
		}
		sm = cip
	}

	buff.WriteString("\nstream\n")
	buff.Write(sm)
	if len(sm) > 0 && sm[len(sm)-1] != 0xA {
		buff.WriteString("\n")
	}
	buff.WriteString("endstream")

	return nil
}

func (p PdfData) isArrayNodes(nodes *pdfNodes) bool {
	if nodes == nil {
		return false
	}
	for _, node := range *nodes {
		if node.key.use == constNodeKeyUseIndex {
			return true
		}
	}
	return false
}

func (p PdfData) isSingleValObjNodes(nodes *pdfNodes) bool {
	if nodes == nil {
		return false
	}
	for _, node := range *nodes {
		if node.key.use == constNodeKeyUseSingleObj {
			return true
		}
	}
	return false
}

func (p PdfData) isStream(nodes *pdfNodes) (int, bool) {
	if nodes == nil {
		return -1, false
	}

	for i, node := range *nodes {
		if node.key.use == constNodeKeyUseStream {
			return i, true
		}
	}
	return -1, false
}

//findRefPageNodeByPageIndex ค้นหา page (pdfNode) จาก pageIndex
func (p *PdfData) findRefPageNodeByPageIndex(pageIndex int) (pdfNode, error) {
	//FIXME: เปลี่ยนระบบหา Page ไปใช้ p.findAllPage()
	//เริ่ม หา page ที่ต้องการ จาก Type == /Pages
	var pagesObjID objectID
	isFoundPagesObj := false
	for id, nodes := range p.objects {
		if node, found := nodes.findOneNodeByName("Type"); found && node.content.use == constNodeContentUseString && node.content.str == "/Pages" {
			pagesObjID = id
			isFoundPagesObj = true
			break
		}
	}

	if !isFoundPagesObj {
		return pdfNode{}, errors.New("not found /Pages")
	}

	var ok bool
	var pagesObj *pdfNodes
	if pagesObj, ok = p.objects[pagesObjID]; !ok {
		return pdfNode{}, errors.New("not found /Pages")
	}

	//เข้าไปดูใน Kids
	var kidsNode pdfNode
	if kidsNode, ok = pagesObj.findOneNodeByName("Kids"); !ok {
		return pdfNode{}, errors.New("not found /Kids in /Pages")
	}
	//fmt.Printf(">>%+v\n", kidsNode)

	if kidsNode.content.use != constNodeContentUseRefTo {
		return pdfNode{}, errors.New("not support yet")
	}

	var pageObjs *pdfNodes
	if pageObjs, ok = p.objects[kidsNode.content.refTo]; !ok {
		return pdfNode{}, errors.New("not found ref to /Page")
	}

	for i, pageObj := range *pageObjs {
		//แต่ละ page
		if i == pageIndex { //เจอ page ที่ต้องการ
			return pageObj, nil
		}
	}
	return pdfNode{}, ErrPdfNodeNotFound
}

//findResourcesOfPage pageIndex เริ่มที่ 0
func (p PdfData) findResourcesOfPage(pageIndex int) (pdfNode, error) {
	page, err := p.findRefPageNodeByPageIndex(pageIndex)
	if err != nil {
		return pdfNode{}, errors.Wrap(err, "")
	}
	pagenodes := p.objects[page.content.refTo]
	resnode, ok := pagenodes.findOneNodeByName("Resources")
	if !ok {
		//FIXME: หาทางแก้ถ้า pdf ยังไม่มี Resources
		return pdfNode{}, errors.Wrapf(err, "not found Resources of pageindex=%d", pageIndex)
	}
	return resnode, nil
}
