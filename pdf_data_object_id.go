package pdfmix

func (p *PdfData) newRealID() objectID {
	p.maxRealID++
	return createRealObjectID(p.maxRealID)
}

func (p *PdfData) newFakeID() objectID {
	p.maxFakeID++
	return createFakeObjectID(p.maxFakeID)
}

func createRealObjectID(id uint32) objectID {
	var o objectID
	o.id = id
	o.isReal = true
	return o
}

func createFakeObjectID(id uint32) objectID {
	var o objectID
	o.id = id
	o.isReal = false
	return o
}
