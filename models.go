package pdfmix

//AlignLeft left
const AlignLeft = 8 //001000
//AlignTop top
const AlignTop = 4 //000100
//AlignRight right
const AlignRight = 2 //000010
//AlignBottom bottom
const AlignBottom = 1 //000001
//AlignCenter center
const AlignCenter = 16 //010000
//AlignMiddle middle
const AlignMiddle = 32 //100000

//Position rect
type Position struct {
	X, Y float64
	W, H float64
}

type TextOption struct {
	Size  float64
	Align int
}

type FontRef string

var FontRefEmpty = FontRef("")
