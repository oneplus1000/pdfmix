package pdfmix

const buildInfoSigID = "SIGID"

type buildInfo struct {
	refIDs map[string]objectID
}

func (b *buildInfo) setRefID(key string, objID objectID) {
	if b.refIDs == nil {
		b.refIDs = make(map[string]objectID)
	}
	b.refIDs[key] = objID
}

func (b *buildInfo) refIDByKey(key string) (objectID, bool) {
	if b.refIDs == nil {
		return objectID{}, false
	}
	objID, ok := b.refIDs[key]
	return objID, ok
}
