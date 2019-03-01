package pdfmix

//BuildPdfOption option for BuildPdfWithOption
type BuildPdfOption struct {
	DigitalSign    DigitalSigner
	PassProtection *PasswordInfo
}
