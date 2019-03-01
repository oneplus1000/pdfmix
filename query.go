package pdfmix

type query struct {
	pdfdata *PdfData
}

func newQuery(p *PdfData) *query {
	var q query
	q.pdfdata = p
	return &q
}

func (q *query) findDict(keyname string, val string) ([]queryResult, error) {
	var results []queryResult
	for objID, nodes := range q.pdfdata.objects {
		for _, node := range *nodes {
			if node.key.use == constNodeKeyUseName &&
				node.key.name == keyname &&
				node.content.use == constNodeContentUseString &&
				node.content.str == val {

				var result queryResult
				result.objID = objID
				result.node = node
				results = append(results, result)

			}
		}
	}
	return results, nil
}

func (q *query) findPdfNodeByKeyName(id objectID, keyname string) (*pdfNode, error) {

	if nodes, ok := q.pdfdata.objects[id]; ok {
		for _, node := range *nodes {
			if node.key.name == keyname {
				return &node, nil
			}
		}
		return nil, ErrKeyNameNotFound
	}
	return nil, ErrObjectIDNotFound
}

func (q *query) findIndexByKeyName(id objectID, keyname string) (int, error) {

	if nodes, ok := q.pdfdata.objects[id]; ok {
		for i, node := range *nodes {
			if node.key.name == keyname {
				return i, nil
			}
		}
		return 0, ErrKeyNameNotFound
	}
	return 0, ErrObjectIDNotFound
}

type queryResult struct {
	objID objectID
	node  pdfNode
}
