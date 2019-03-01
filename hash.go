package pdfmix

import (
	"crypto/sha1"
	"fmt"

	"github.com/pkg/errors"
)

func hashSha1(b []byte) (string, error) {
	h := sha1.New()
	_, err := h.Write(b)
	if err != nil {
		return "", errors.Wrap(err, " h.Write(...) fail")
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs), nil
}
