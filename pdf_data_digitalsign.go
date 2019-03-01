package pdfmix

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

var idContnetType = []int{1, 2, 840, 113549, 1, 9, 3}
var idPkcs7Data = []int{1, 2, 840, 113549, 1, 7, 1}
var idMessageDigest = []int{1, 2, 840, 113549, 1, 9, 4}
var idSHA256 = []int{2, 16, 840, 1, 101, 3, 4, 2, 1}
var idPKCS7SignedData = []int{1, 2, 840, 113549, 1, 7, 2}
var idRSA = []int{1, 2, 840, 113549, 1, 1, 1}

const fakeByteRange = "[0000000000 0000000000 0000000000 0000000000]"

func (p *PdfData) setDigitalSigner(ds DigitalSigner) {
	p.ds = ds
}

func (p *PdfData) buildAndBytesWithSign() ([]byte, error) {
	err := p.preSigning()
	if err != nil {
		return nil, errors.Wrap(err, "p.preSigning(...) fail")
	}
	data, err := p.postSigning()
	if err != nil {
		return nil, errors.Wrap(err, "p.signing(...) fail")
	}
	return data, nil
}

//preSigning setup digital signatures ทำพวก insert dic ต่างๆลงไป
func (p *PdfData) preSigning() error {

	pageIndex := 0 // page default
	//rect := Position{} //rect default ของ annot เวลาไม่ visible

	catalogID, err := p.findCatalogID()
	if err != nil {
		return errors.Wrapf(err, "not found Catalog")
	}

	refPage, err := p.findRefPageNodeByPageIndex(pageIndex)
	if err != nil {
		return errors.Wrapf(err, "not found page (index = %d)", pageIndex)
	}

	refAnnotID := p.newRealID()
	refSigID := p.newRealID()

	annot, err := p.createSignAnnot(refPage, refSigID)
	if err != nil {
		return errors.Wrap(err, "cannot create /Annot for digital signatures")
	}

	//TODO ยังไม่ support กรณีที่มี annot แล้วนะ
	refAnnotIDItemsID := p.newFakeID()
	refAnnotIDItems := pdfNodes{}
	p.objects[refAnnotIDItemsID] = &refAnnotIDItems
	refAnnotIDItems.append(initNodeKeyUseIndexAndNodeContentUseRefTo(refAnnotID))
	refAnnot := initNodeKeyUseNameAndNodeContentUseRefTo("Annots", refAnnotIDItemsID)

	p.objects[refPage.content.refTo].append(refAnnot)
	p.objects[refAnnotID] = &annot

	//sig
	sig, err := p.createSig()
	if err != nil {
		return errors.Wrap(err, "createSig fail")
	}
	p.objects[refSigID] = &sig
	p.buildinfo.setRefID(buildInfoSigID, refSigID)

	acroForm, err := p.acroForm(refAnnotID)
	if err != nil {
		return errors.Wrapf(err, "not found Catalog")
	}

	catalog := p.objects[catalogID]
	catalog.append(acroForm)

	return nil
}

func (p *PdfData) postSigning() ([]byte, error) {
	ds := p.ds
	sig, all, err := p.bodyForSig()
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	hash := sha256.Sum256(sig)

	authset := authSet{
		ContentTypeSeq: contentTypeSeq{
			ContentTypeID: idContnetType,
			ContentTypeDataSet: contentTypeDataSet{
				Pkcs7DataID: idPkcs7Data,
			},
		},
		DigestSet: digestSet{
			MessageDigestID: idMessageDigest,
			SecondDigestSet: secondDigestSet{
				Data: hash[:],
			},
		},
	}

	der, err := asn1.Marshal(authset)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	der[0] = asn1.TagSet | 0x20

	//fmt.Printf("%x\n\n", der)
	//privateKey, certificate := ds.Cert()
	signedcontent, err := ds.Sign(der) //precessCert(der, privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	encodedSig, err := encodedPKCS7(signedcontent, authset, ds.Cert())
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	signContent, err := p.sigContentWithData(encodedSig)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	sigContentEmpty, _ := p.sigContent()
	result := bytes.Replace(all, []byte(sigContentEmpty), []byte(signContent), -1)
	return result, nil
}

func (p *PdfData) bodyForSig() ([]byte, []byte, error) {

	err := p.build()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	var offsetOfSigStart, offsetOfSigEnd, endOfFile int
	p.bytesCallBack = func(buff *bytes.Buffer, currRealObjID objectID, status int) {
		refSigID, ok := p.buildinfo.refIDByKey(buildInfoSigID)
		if !ok {
			return
		}
		if status == statusStartObj && currRealObjID.compare(refSigID) {
			offsetOfSigStart = buff.Len()
		} else if status == statusEndObj && currRealObjID.compare(refSigID) {
			offsetOfSigEnd = buff.Len()
		} else if status == statusEndOfFile {
			endOfFile = buff.Len()
		}
	}
	rawAll, err := p.bytes()
	if err != nil {
		return nil, nil, errors.Wrap(err, "")
	}
	p.bytesCallBack = nil //clear bytesCallBack
	rawSig := rawAll[offsetOfSigStart:offsetOfSigEnd]
	sigContentEmpty, _ := p.sigContent()
	start := offsetOfSigStart + bytes.Index(rawSig, []byte(sigContentEmpty))
	stop := offsetOfSigStart + bytes.LastIndex(rawSig, []byte(sigContentEmpty)) + len(sigContentEmpty)
	realByteRange := fmt.Sprintf("[%d %d %d %d ]", 0, start, stop-1, endOfFile-stop)
	length := len(fakeByteRange)
	for len(realByteRange) < length {
		realByteRange = realByteRange + " "
	}

	rawSig = bytes.Replace(rawSig, []byte(fakeByteRange), []byte(realByteRange), -1)
	rawSig = bytes.Replace(rawSig, []byte(sigContentEmpty), []byte{}, -1)

	tmpA := make([]byte, len(rawAll[0:offsetOfSigStart]))
	copy(tmpA, rawAll[0:offsetOfSigStart])
	tmpB := make([]byte, len(rawAll[offsetOfSigEnd:]))
	copy(tmpB, rawAll[offsetOfSigEnd:])
	result := append(tmpA, rawSig...)
	result = append(result, tmpB...)

	rawAll = bytes.Replace(rawAll, []byte(fakeByteRange), []byte(realByteRange), -1)
	//fmt.Printf("\n%s\n%d...%d ...%d\n\n", string(result), len(result), len(rawAll), len(sigContentEmpty))
	return result, rawAll, nil
}

func encodedPKCS7(data []byte, secondDigest authSet, cert *x509.Certificate) ([]byte, error) {

	secondDigestDer, err := asn1.Marshal(secondDigest)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	whole := wholeSeq{
		SignedData: idPKCS7SignedData,
		BodyTag: bodyTag{
			BodySeq: bodySeq{
				Version: 1,
				DigestAlgorithms: digestAlgorithmsSet{
					AlgosSeq: algosSeq{
						Dal:  idSHA256,
						Null: asn1.RawValue{Tag: asn1.TagNull},
					},
				},
				ContentinfoSeq: contentinfoSeq{
					V: idPkcs7Data,
				},
				DercertificatesTag: dercertificatesTag{
					DercertificatesSet: dercertificatesSet{
						Raw: cert.Raw,
					},
				},
				SignerinfoSet: signerinfoSet{
					SignerinfoSeq: signerinfoSeq{
						SignerVersion: 1,
						SignCertSeq: signCertSeq{
							IssuerSeq: issuerSeq{
								IssuerCont: cert.RawIssuer,
							},
							SerialNumber: cert.SerialNumber.Int64(),
						},
						DigestAlgorithmOidSeq: digestAlgorithmOidSeq{
							DigestAlgorithmOid: idSHA256,
							Null:               asn1.RawValue{Tag: asn1.TagNull},
						},
						SecondDigestTag: secondDigestTag{
							SecondDigest: secondDigestDer,
						},
						DigestEncryptionAlgorithmSeq: digestEncryptionAlgorithmSeq{
							DigestEncryptionAlgorithmOid: idRSA,
							Null:                         asn1.RawValue{Tag: asn1.TagNull},
						},
						Digest: data,
					},
				},
			},
		},
	}

	a, err := asn1.Marshal(whole)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	return a, nil
}

/*
func precessCert(data []byte, ds DigitalSigner) ([]byte, error) {

	signer, err := newSignerFromKey(privateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	//data = []byte("HI2")
	dataSigned, err := signer.Sign(data)
	if err != nil {
		return nil, errors.Wrapf(err, "")
	}

	return dataSigned, nil
}*/

const estimatedContentSize = 8191 //TODO แก้ทีหลังด้วย

func (p *PdfData) sigContentWithData(data []byte) (string, error) {
	var buff bytes.Buffer
	buff.WriteString("<")
	i := 0
	for _, d := range data {
		s := fmt.Sprintf("%x", d)
		if len(s) == 1 {
			buff.WriteString("0")
		}
		buff.WriteString(s)
		//fmt.Printf("%x  ", d)
		i++
	}
	s := buff.Len()
	i = 0
	for i < estimatedContentSize-s {
		buff.WriteString("0")
		i++
	}
	buff.WriteString(">")
	return buff.String(), nil
}

func (p *PdfData) sigContent() (string, error) {
	var buff bytes.Buffer
	buff.WriteString("<")
	for i := 0; i < estimatedContentSize; i++ {
		buff.WriteString("0")
	}
	buff.WriteString(">")
	return buff.String(), nil
}

func (p *PdfData) acroForm(annotID objectID) (pdfNode, error) {

	acroFormID := p.newFakeID()
	arco := initNodeKeyUseNameAndNodeContentUseRefTo("AcroForm", acroFormID)
	arcoForm := pdfNodes{}
	p.objects[acroFormID] = &arcoForm

	acroFormItemsID := p.newFakeID()
	annotItems := pdfNodes{}
	p.objects[acroFormItemsID] = &annotItems
	annotItems.append(initNodeKeyUseIndexAndNodeContentUseRefTo(annotID))

	fields := initNodeKeyUseNameAndNodeContentUseRefTo("Fields", acroFormItemsID)
	arcoForm.append(fields)

	sigFlags := initNodeKeyUseNameAndNodeContentUseString("SigFlags", "3")
	arcoForm.append(sigFlags)

	return arco, nil
}

func formatPdfDateTime(t time.Time) string {
	return "D:" + t.UTC().Format("20060102150405") + "+07'00'"
}

func (p *PdfData) createSig() (pdfNodes, error) {

	appID := p.newFakeID()
	nameID := p.newFakeID()

	var sig pdfNodes
	sigType := initNodeKeyUseNameAndNodeContentUseString("Type", "/Sig")
	sigFilter := initNodeKeyUseNameAndNodeContentUseString("Filter", "/Adobe.PPKLite")
	sigSubFilter := initNodeKeyUseNameAndNodeContentUseString("SubFilter", "/adbe.pkcs7.detached")
	//sigName := initNodeKeyUseNameAndNodeContentUseString("Name", "(oneplus x)")
	sigLocation := initNodeKeyUseNameAndNodeContentUseString("Location", "(Bangkok, Thailand)")
	sigReason := initNodeKeyUseNameAndNodeContentUseString("Reason", "(Verified by OIC)")

	sigPropBuild := initNodeKeyUseNameAndNodeContentUseRefTo("Prop_Build", appID)

	app := pdfNodes{}
	p.objects[appID] = &app
	appApp := initNodeKeyUseNameAndNodeContentUseRefTo("App", nameID)
	app.append(appApp)

	name := pdfNodes{}
	p.objects[nameID] = &name
	nameName := initNodeKeyUseNameAndNodeContentUseString("Name", "(edmon engine)")
	name.append(nameName)

	content, err := p.sigContent()
	if err != nil {
		return pdfNodes{}, errors.Wrap(err, "createSigContent fail")
	}
	sigContents := initNodeKeyUseNameAndNodeContentUseString("Contents", content)
	byteRange := initNodeKeyUseNameAndNodeContentUseString("ByteRange", fakeByteRange)
	m := initNodeKeyUseNameAndNodeContentUseString("M", "("+formatPdfDateTime(time.Now())+")")
	contactInfo := initNodeKeyUseNameAndNodeContentUseString("ContactInfo", "()")

	sig.append(sigType)
	sig.append(sigFilter)
	sig.append(sigSubFilter)
	//sig.append(sigName)
	sig.append(sigLocation)
	sig.append(sigReason)
	sig.append(sigPropBuild)
	sig.append(m)
	sig.append(contactInfo)
	sig.append(byteRange)
	sig.append(sigContents)

	return sig, nil
}

func (p *PdfData) createSignAnnot(
	refPage pdfNode,
	refSigID objectID,
) (
	pdfNodes,
	error,
) {

	var annot pdfNodes
	annotType := initNodeKeyUseNameAndNodeContentUseString("Type", "/Annot")
	annotP := initNodeKeyUseNameAndNodeContentUseRefTo("P", refPage.content.refTo)
	annotSubtype := initNodeKeyUseNameAndNodeContentUseString("Subtype", "/Widget")
	annotV := initNodeKeyUseNameAndNodeContentUseRefTo("V", refSigID)
	annotT := initNodeKeyUseNameAndNodeContentUseString("T", "(Signature1)")
	annotFT := initNodeKeyUseNameAndNodeContentUseString("FT", "/Sig")
	annotF := initNodeKeyUseNameAndNodeContentUseString("F", "4")
	annotDA := initNodeKeyUseNameAndNodeContentUseString("DA", "(/Helvetica 0 Tf 0 g)")

	if ds, ok := p.ds.(DigitalSignWithAnnoter); ok && ds.Annot() != nil {
		//ถ้า ds impl DigitalSignWithAnnoter -> visible sign
		// ap
		boxID, err := p.createAnnotVisibleBox(ds)
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
		nnode := initNodeKeyUseNameAndNodeContentUseRefTo("N", boxID)
		annotNNodes := pdfNodes{}
		annotNNodes.append(nnode)
		nID := p.newFakeID()
		p.objects[nID] = &annotNNodes
		annotAp := initNodeKeyUseNameAndNodeContentUseRefTo("AP", nID)
		//annot.append(annotAp)
		_ = annotAp
		// rect
		rect := ds.Annot().Pos
		annotRect := initNodeKeyUseNameAndNodeContentUseString("Rect", fmt.Sprintf("[%0.2f %0.2f %0.2f %0.2f]", rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H))
		annot.append(annotRect)

	} else {
		annotRect := initNodeKeyUseNameAndNodeContentUseString("Rect", "[0 0 0 0]")
		annot.append(annotRect)
	}

	annot.append(annotType)
	annot.append(annotSubtype)
	annot.append(annotP)
	annot.append(annotV)
	annot.append(annotT)
	annot.append(annotFT)
	annot.append(annotF)
	//annot.append(annotDA)
	_ = annotDA

	return annot, nil
}

//createAnnotVisibleBox โชว text และ img บน annot
func (p *PdfData) createAnnotVisibleBox(ds DigitalSignWithAnnoter) (objectID, error) {

	nodesID := p.newRealID()
	nodes := pdfNodes{}
	p.objects[nodesID] = &nodes

	cropBox := initNodeKeyUseNameAndNodeContentUseString("CropBox", "[0 0 197 70]")
	typ := initNodeKeyUseNameAndNodeContentUseString("Type", "/XObject")
	filter := initNodeKeyUseNameAndNodeContentUseString("Filter", "/FlateDecode")
	formType := initNodeKeyUseNameAndNodeContentUseString("FormType", "1")
	bbox := initNodeKeyUseNameAndNodeContentUseString("BBox", "[0 0 197.0 70.0]")
	mediaBox := initNodeKeyUseNameAndNodeContentUseString("MediaBox", "[0 0 197 70]")
	subtype := initNodeKeyUseNameAndNodeContentUseString("Subtype", "/Form")

	annot := ds.Annot()
	img1ID, err := p.createImg(annot.Img)
	if err != nil {
		return objectID{}, errors.Wrap(err, "createImg(...) fail")
	}
	imgNodes := pdfNodes{}
	img1 := initNodeKeyUseNameAndNodeContentUseRefTo("Img1", img1ID)
	imgNodes.append(img1)
	imgNodesID := p.newFakeID()
	p.objects[imgNodesID] = &imgNodes

	xObjNodes := pdfNodes{}
	xObjNodesID := p.newFakeID()
	p.objects[xObjNodesID] = &xObjNodes
	xObj := initNodeKeyUseNameAndNodeContentUseRefTo("XObject", imgNodesID)
	fontObj := initNodeKeyUseNameAndNodeContentUseString("Font", "<<>>")
	xObjNodes.append(xObj)
	xObjNodes.append(fontObj)

	res := initNodeKeyUseNameAndNodeContentUseRefTo("Resources", xObjNodesID)

	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("q %0.2f 0 0 %0.2f %0.2f %0.2f cm /Img1  Do Q\n", 100.0, 100.0, 10.0, 10.0))
	stm := initNodeKeyUseStreamNodeContentUseStream(buff.Bytes())
	length := initNodeKeyUseNameAndNodeContentUseString("Length", fmt.Sprintf("%d", buff.Len()))

	nodes.append(cropBox)
	nodes.append(typ)
	nodes.append(filter)
	nodes.append(formType)
	nodes.append(bbox)
	nodes.append(res)
	nodes.append(mediaBox)
	nodes.append(subtype)
	nodes.append(stm)
	nodes.append(length)
	return nodesID, nil
}

func (p *PdfData) createXObjectForXObject(maxRealID uint32, maxFakeID uint32) (objectID, uint32, uint32, error) {

	stm := []byte("% DSBlank")
	xobjID := p.newRealID()
	xobj := pdfNodes{}
	xobjType := initNodeKeyUseNameAndNodeContentUseString("Type", "/XObject")
	xobjSubType := initNodeKeyUseNameAndNodeContentUseString("Subtype", "/Form")
	xobjBbox := initNodeKeyUseNameAndNodeContentUseString("BBox", "[0 0 0 0]")
	xobjFromType := initNodeKeyUseNameAndNodeContentUseString("FormType", "1")
	xobjMatrix := initNodeKeyUseNameAndNodeContentUseString("Matrix", "[1 0 0 1 0 0]")
	xobjLength := initNodeKeyUseNameAndNodeContentUseString("Length", fmt.Sprintf("%d", len(stm)))
	xobjFilter := initNodeKeyUseNameAndNodeContentUseString("Filter", "/FlateDecode")
	xobjStream := initNodeKeyUseStreamNodeContentUseStream(stm)

	xobj.append(xobjType)
	xobj.append(xobjSubType)
	xobj.append(xobjBbox)
	xobj.append(xobjFromType)
	xobj.append(xobjMatrix)
	xobj.append(xobjLength)
	xobj.append(xobjFilter)
	xobj.append(xobjStream)
	p.objects[xobjID] = &xobj
	return xobjID, 0, 0, nil
}

func (p *PdfData) findCatalogID() (objectID, error) {
	q := newQuery(p)
	results, err := q.findDict("Type", "/Catalog")
	if err != nil {
		return objectID{}, errors.Wrap(err, "")
	}
	for _, r := range results {
		return r.objID, nil
	}
	return objectID{}, ErrPdfNodeNotFound
}
