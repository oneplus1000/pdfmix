package pdfmix

import "io"

//buffer for pdf content
type buffer struct {
	position int
	datas    []byte
}

//Write : write []byte to buffer
func (b *buffer) Write(p []byte) (int, error) {
	for len(b.datas) < b.position+len(p) {
		b.datas = append(b.datas, 0)
	}
	i := 0
	max := len(p)
	for i < max {
		b.datas[i+b.position] = p[i]
		i++
	}
	b.position += i
	return 0, nil
}

//Len : len of buffer
func (b *buffer) Len() int {
	return len(b.datas)
}

//Bytes : get bytes
func (b *buffer) Bytes() []byte {
	return b.datas
}

//Position : get current postion
func (b *buffer) Position() int {
	return b.position
}

//SetPosition : set current postion
func (b *buffer) SetPosition(pos int) {
	b.position = pos
}

//writeUInt32  writes a 32-bit unsigned integer value to w io.Writer
func writeUInt32(w io.Writer, v uint) error {
	a := byte(v >> 24)
	b := byte(v >> 16)
	c := byte(v >> 8)
	d := byte(v)
	_, err := w.Write([]byte{a, b, c, d})
	if err != nil {
		return err
	}
	return nil
}

//writeUInt16 writes a 16-bit unsigned integer value to w io.Writer
func writeUInt16(w io.Writer, v uint) error {

	a := byte(v >> 8)
	b := byte(v)
	_, err := w.Write([]byte{a, b})
	if err != nil {
		return err
	}
	return nil
}

//writeTag writes string value to w io.Writer
func writeTag(w io.Writer, tag string) error {
	b := []byte(tag)
	_, err := w.Write(b)
	if err != nil {
		return err
	}
	return nil
}

//writeBytes writes []byte value to w io.Writer
func writeBytes(w io.Writer, data []byte, offset int, count int) error {

	_, err := w.Write(data[offset : offset+count])
	if err != nil {
		return err
	}
	return nil
}
