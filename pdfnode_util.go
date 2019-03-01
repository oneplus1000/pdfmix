package pdfmix

func initNodeKeyUseIndexAndNodeContentUseRefTo(id objectID) pdfNode {
	n := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: id,
		},
	}
	return n
}

func initNodeKeyUseNameAndNodeContentUseString(nodeName string, contentStr string) pdfNode {
	n := pdfNode{
		key: nodeKey{
			name: nodeName,
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use: constNodeContentUseString,
			str: contentStr,
		},
	}
	return n
}

func initNodeKeyUseNameAndNodeContentUseRefTo(nodeName string, refObjID objectID) pdfNode {
	n := pdfNode{
		key: nodeKey{
			name: nodeName,
			use:  constNodeKeyUseName,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: refObjID,
		},
	}
	return n
}

func initNodeKeyUseStreamNodeContentUseStream(data []byte) pdfNode {

	n := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseStream,
		},
		content: nodeContent{
			use:    constNodeContentUseStream,
			stream: data,
		},
	}
	return n
}

func initNodeKeyUseIndexNodeContentUseRefTo(refObjID objectID) pdfNode {
	n := pdfNode{
		key: nodeKey{
			use: constNodeKeyUseIndex,
		},
		content: nodeContent{
			use:   constNodeContentUseRefTo,
			refTo: refObjID,
		},
	}
	return n
}
