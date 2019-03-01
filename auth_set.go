package pdfmix

import "encoding/asn1"

type authSet struct {
	ContentTypeSeq contentTypeSeq
	DigestSet      digestSet
}

type contentTypeSeq struct {
	ContentTypeID      asn1.ObjectIdentifier
	ContentTypeDataSet contentTypeDataSet `asn1:"set"`
}

type digestSet struct {
	MessageDigestID asn1.ObjectIdentifier
	SecondDigestSet secondDigestSet `asn1:"set"`
}

type contentTypeDataSet struct {
	Pkcs7DataID asn1.ObjectIdentifier
}

type secondDigestSet struct {
	Data []byte
}
