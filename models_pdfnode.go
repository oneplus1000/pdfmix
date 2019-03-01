package pdfmix

import "fmt"

type objectID struct {
	isReal bool
	id     uint32
	//fromRealID uint32
}

var objectIDEmpty = objectID{}

func (o objectID) String() string {
	if o.isReal {
		return fmt.Sprintf("%d", o.id)
	}
	return fmt.Sprintf("%df", o.id)
}

func (o objectID) compare(id objectID) bool {
	if id.isReal == o.isReal && id.id == o.id {
		return true
	}
	return false
}

type pdfNodes []pdfNode

func (p *pdfNodes) len() int {
	return len(*p)
}

func (p *pdfNodes) append(n pdfNode) {
	*p = append(*p, n)
}

func (p *pdfNodes) remove(index int) {
	*p = append((*p)[:index], (*p)[index+1:]...)
}

//findOneNodeByName find node by name return only first one
func (p *pdfNodes) findOneNodeByName(name string) (pdfNode, bool /*true = found*/) {
	for _, node := range *p {
		if node.key.use == constNodeKeyUseName && node.key.name == name {
			return node, true
		}
	}
	return pdfNode{}, false
}

type pdfNode struct {
	key     nodeKey
	content nodeContent
}

func (p pdfNode) clone() pdfNode {
	return p
}

const (
	constNodeKeyUseName      = 1
	constNodeKeyUseIndex     = 2 //สำหรับ array
	constNodeKeyUseStream    = 3
	constNodeKeyUseSingleObj = 4
)

type nodeKey struct {
	use   int // 1 = name , 2 = index , 3 = stream , 4 = single obj
	name  string
	index int
}

const (
	constNodeContentUseString    = 1
	constNodeContentUseRefTo     = 2
	constNodeContentUseStream    = 3
	constNodeContentUseSingleObj = 4
	constNodeContentUseNull      = 5
)

type nodeContent struct {
	use    int // 1 = str , 2 refTo , 3 = stream , 4 = single obj
	str    string
	refTo  objectID
	stream []byte
}
