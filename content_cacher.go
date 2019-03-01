package pdfmix

import "io"

type contentCacher interface {
	build(w io.Writer) (int64, error)
}

//contenteCacheText
