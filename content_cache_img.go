package pdfmix

import (
	"bytes"
	"fmt"
	"io"
)

type contenteCacheImg struct {
	imgName string
	x, y    float64
	w, h    float64
}

func (c *contenteCacheImg) build(w io.Writer) (int64, error) {
	var buff bytes.Buffer
	//buff.WriteString(fmt.Sprintf("q %0.2f 0 0 %0.2f %0.2f %0.2f cm /%s Do Q\n", c.rect.W, c.rect.H, c.x, c.h-(c.y+c.rect.H), c.imgName))
	buff.WriteString(fmt.Sprintf("q %0.2f 0 0 %0.2f %0.2f %0.2f cm /%s Do Q\n", c.w, c.h, c.x, c.y, c.imgName))
	return buff.WriteTo(w)
}
