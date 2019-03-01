package pdfmix

import "encoding/asn1"

type wholeSeq struct {
	SignedData asn1.ObjectIdentifier
	BodyTag    bodyTag `asn1:"tag:0"`
}

type bodyTag struct {
	BodySeq bodySeq
}

type bodySeq struct {
	Version            int
	DigestAlgorithms   digestAlgorithmsSet `asn1:"set"`
	ContentinfoSeq     contentinfoSeq
	DercertificatesTag dercertificatesTag `asn1:"tag:0"`
	SignerinfoSet      signerinfoSet      `asn1:"set"`
}

type digestAlgorithmsSet struct {
	AlgosSeq algosSeq
}

type algosSeq struct {
	Dal  asn1.ObjectIdentifier
	Null asn1.RawValue
}

type contentinfoSeq struct {
	V asn1.ObjectIdentifier
}

type dercertificatesTag struct {
	DercertificatesSet dercertificatesSet
}

type dercertificatesSet struct {
	Raw asn1.RawContent
}

type signerinfoSet struct {
	SignerinfoSeq signerinfoSeq
}

type signerinfoSeq struct {
	SignerVersion                int
	SignCertSeq                  signCertSeq
	DigestAlgorithmOidSeq        digestAlgorithmOidSeq
	SecondDigestTag              secondDigestTag `asn1:"tag:0"`
	DigestEncryptionAlgorithmSeq digestEncryptionAlgorithmSeq
	Digest                       []byte
}

type digestEncryptionAlgorithmSeq struct {
	DigestEncryptionAlgorithmOid asn1.ObjectIdentifier
	Null                         asn1.RawValue
}

type secondDigestTag struct {
	SecondDigest asn1.RawContent
}

type digestAlgorithmOidSeq struct {
	DigestAlgorithmOid asn1.ObjectIdentifier
	Null               asn1.RawValue
}

type signCertSeq struct {
	IssuerSeq    issuerSeq
	SerialNumber int64
}

type issuerSeq struct {
	IssuerCont asn1.RawContent
}
