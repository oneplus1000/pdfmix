package pdfmix

type pageContentCacherMap struct {
	keys           map[int]([]int) //map[index ของ หน้า] index(s)ของ array contentCachers
	contentCachers []contentCacher
}

func (p *pageContentCacherMap) append(pageIndex int, cc contentCacher) {
	p.contentCachers = append(p.contentCachers, cc)
	arrIndex := len(p.contentCachers) - 1

	if p.keys == nil {
		p.keys = make(map[int]([]int))
	}
	arrIndexs := p.keys[pageIndex]
	arrIndexs = append(arrIndexs, arrIndex)
	p.keys[pageIndex] = arrIndexs
}

func (p *pageContentCacherMap) allPageIndexs() []int {
	pageindexs := make([]int, len(p.keys))
	i := 0
	for pageindex := range p.keys {
		pageindexs[i] = pageindex
		i++
	}
	return pageindexs
}

func (p *pageContentCacherMap) contentCachersByPageIndex(pageIndex int) []contentCacher {
	indexs := p.keys[pageIndex]
	ccs := make([]contentCacher, len(indexs))
	for i, index := range indexs {
		ccs[i] = p.contentCachers[index]
	}
	return ccs
}

/*
func (p *pageContentCacherMap) isPageIndexExist(pageIndex int) bool {
	if _, ok := p.keys[pageIndex]; ok {
		return true
	}
	return false
}*/
