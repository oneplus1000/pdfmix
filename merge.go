package pdfmix

import (
	"fmt"

	"github.com/pkg/errors"
)

//ErrCannotFindPdfObjectCatalog  can not find pdf object Catalog
var ErrCannotFindPdfObjectCatalog = errors.New("can not find pdf object Catalog")

//ErrCannotFindPdfObjectPages  can not find pdf object Pages
var ErrCannotFindPdfObjectPages = errors.New("can not find pdf object Pages")

func merge(a, b *PdfData) error {

	maxRealIDOfA, maxFakeIDOfA, err := maxID(a)
	if err != nil {
		return errors.Wrap(err, "")
	}
	//fmt.Printf("%d %d\n", maxRealIDOfA, maxFakeIDOfA)

	//remove Catalog,Trailer b
	tempB, err := shiftID(b, maxRealIDOfA+1, maxFakeIDOfA+1)
	if err != nil {
		return errors.Wrap(err, "")
	}
	//tempB := b

	err = removeTrailer(tempB)
	if err != nil {
		return errors.Wrap(err, "")
	}

	/*err = removeCatalog(tempB)
	if err != nil {
		return errors.Wrap(err, "")
	}*/

	//merge Pages a and b together (into a)
	err = mergePages(a, tempB, maxRealIDOfA)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}

func mergePages(a, b *PdfData, maxRealIDOfA uint32) error {

	results, err := newQuery(a).findDict("Type", "/Pages")
	if err != nil {
		return errors.Wrap(err, "")
	}

	if len(results) <= 0 {
		return ErrCannotFindPdfObjectPages
	}
	pagesOfA := results[0]

	results, err = newQuery(b).findDict("Type", "/Pages")
	if err != nil {
		return errors.Wrap(err, "")
	}

	if len(results) <= 0 {
		return ErrCannotFindPdfObjectPages
	}
	pagesOfB := results[0]

	var kidsIDOfA objectID
	for _, nodeOfA := range *a.objects[pagesOfA.objID] {
		if nodeOfA.key.use == constNodeKeyUseName && nodeOfA.key.name == "Kids" {
			kidsIDOfA = nodeOfA.content.refTo
			break
		}
	}

	var kidsIDOfB objectID
	for _, nodeOfB := range *b.objects[pagesOfB.objID] {
		if nodeOfB.key.use == constNodeKeyUseName && nodeOfB.key.name == "Kids" {
			kidsIDOfB = nodeOfB.content.refTo
			break
		}
	}

	//merge
	for objID, obj := range b.objects {
		if objID != kidsIDOfB {
			a.objects[objID] = obj
		}
	}

	for _, node := range *b.objects[kidsIDOfB] {
		a.objects[kidsIDOfA].append(node)
	}

	count := a.objects[kidsIDOfA].len()
	for i, node := range *a.objects[pagesOfA.objID] {
		if node.key.use == constNodeKeyUseName && node.key.name == "Count" {
			(*a.objects[pagesOfA.objID])[i].content.str = fmt.Sprintf("%d", count)
		}
	}

	return nil
}

func removeTrailer(src *PdfData) error {
	trailerObjID := createRealObjectID(0) //Trailer away 0
	delete(src.objects, trailerObjID)
	return nil
}

func removeCatalog(src *PdfData) error {
	results, err := newQuery(src).findDict("Type", "/Catalog")
	if err != nil {
		return errors.Wrap(err, "")
	}
	if len(results) <= 0 {
		return ErrCannotFindPdfObjectCatalog
	}
	delete(src.objects, results[0].objID)
	return nil
}

func removePages(src *PdfData) error {
	results, err := newQuery(src).findDict("Type", "/Pages")
	if err != nil {
		return errors.Wrap(err, "")
	}
	if len(results) <= 0 {
		return ErrCannotFindPdfObjectPages
	}
	delete(src.objects, results[0].objID)
	return nil
}

func clonePdfData(src *PdfData) *PdfData {
	dest := newPdfData()
	dest.objects = src.objects
	return dest
}

func shiftID(src *PdfData, realIDOffset uint32, fakeIDOffset uint32) (*PdfData, error) {
	dest := newPdfData()
	for srcID := range src.objects {
		var destID objectID
		destID.isReal = srcID.isReal
		if destID.isReal {
			destID.id = srcID.id + realIDOffset
		} else {
			destID.id = srcID.id + fakeIDOffset
		}
		srcNodes := src.objects[srcID]
		size := srcNodes.len()
		for i := 0; i < size; i++ {
			srcNode := (*srcNodes)[i]
			destNode := srcNode.clone()
			if destNode.content.use == constNodeContentUseRefTo {
				if destNode.content.refTo.isReal {
					destNode.content.refTo.id = destNode.content.refTo.id + realIDOffset
				} else {
					destNode.content.refTo.id = destNode.content.refTo.id + fakeIDOffset
				}
			}
			dest.push(destID, destNode)
		}
	}
	return dest, nil
}

func maxID(a *PdfData) (uint32, uint32, error) {
	maxRealID := uint32(0)
	maxFakeID := uint32(0)
	for objID := range a.objects {
		if objID.isReal && objID.id > maxRealID {
			maxRealID = objID.id
		}
		if !objID.isReal && objID.id > maxFakeID {
			maxFakeID = objID.id
		}
	}
	return maxRealID, maxRealID, nil
}
