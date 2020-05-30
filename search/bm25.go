package search

import (
	"container/heap"
	"github.com/james-bowman/sparse"
	"io"
	"math"
	"strings"
)

type Bm25 struct {
	K, B float64
	M *sparse.CSR
	Terms []string
	InvertedList map[string][]*documentWrapper
	UseTfIdf bool
	MaxResults int
}

func NewBm25(b, k float64) *Bm25 {
	return &Bm25{
		B: b,
		K: k,
		Terms: make([]string, 0),
		InvertedList: make(map[string][]*documentWrapper, 0),
		UseTfIdf: false,
		MaxResults: 10,
	}
}

func (bm *Bm25) Load(r io.Reader, documents []*Document) error {
	if bm.M == nil {
		bm.M = sparse.NewDOK(0, 0).ToCSR()
	}
	if len(bm.Terms) == 0 {
		bm.addDocuments(documents)
		bm.InvertedList = nil
	}
	_, err := bm.M.UnmarshalBinaryFrom(r)
	return err
}

func (bm *Bm25) Save(w io.Writer) error {
	_, err := bm.M.MarshalBinaryTo(w)
	return err
}

func (bm *Bm25) Build(documents []*Document) {
	bm.addDocuments(documents)

	numberOfWords := len(bm.InvertedList)
	numberOfDocuments := len(documents)

	dokMat := sparse.NewDOK(numberOfDocuments, numberOfWords)

	n := float64(numberOfDocuments)
	avdl := avgDocumentLength(documents)


	for termId, term := range bm.Terms {
		for _, wrapper := range bm.InvertedList[term] {
			// fmt.Print(0)

			if bm.UseTfIdf {
				dokMat.Set(wrapper.d.Row, termId, float64(wrapper.ctr))
				continue
			}
			val := float64(wrapper.ctr)
			// fmt.Print(" ", 1)
			alpha := 1 - bm.B + (bm.B * float64(len(wrapper.d.Terms)) / avdl)
			// fmt.Print(" ", 2)
			tf := 1.0
			if bm.K > 0 {
				tf = val * (1 + (1 / bm.K)) / (alpha + (val / bm.K))
			}
			// fmt.Print(" ", 3)

			df := float64(len(bm.InvertedList[term])) // TODO: check the value
			// fmt.Print(" ", 4)

			v := tf * math.Log2(n / df)
			// fmt.Print(" ", 5)

			dokMat.Set(wrapper.d.Row, termId, v)
			// fmt.Print(" ", 6)
			// fmt.Printf("\n")

		}
		if termId % 1000 == 0 {
			// fmt.Printf("Finished %v terms\n", termId)
		}
	}

	bm.M = dokMat.ToCSR()


}

func (bm *Bm25) addDocuments(documents []*Document) {
	for _, document := range documents {
		for _, term := range document.Terms {
			term = strings.ToLower(term)
			if _, ok := bm.InvertedList[term]; !ok {
				bm.InvertedList[term] = make([]*documentWrapper, 1)
				// TODO: Fix
				bm.InvertedList[term][0] = &documentWrapper{d: document, ctr: 0}
				bm.Terms = append(bm.Terms, term)
			}
			if bm.InvertedList[term][len(bm.InvertedList[term]) - 1].d.ID == document.ID {
				bm.InvertedList[term][len(bm.InvertedList[term]) - 1].ctr++
				// continue
			} else {
				bm.InvertedList[term] = append(bm.InvertedList[term], &documentWrapper{d: document, ctr: 1})
			}
		}
	}
}

func (bm *Bm25) Search(terms []string) []int {
	if bm.M == nil {
		panic("Bm25 is not initialised")
	}
	searchVec := bm.buildSearchVec(terms)

	var searchResult sparse.CSR
	searchResult.Mul(bm.M, searchVec)

	pq := make(PriorityQueue, searchResult.NNZ())

	rows, _ := searchResult.Dims()

	ctr := 0
	for i := 0; i < rows; i++ {
		if searchResult.At(i, 0) == 0 {
			continue
		}
		pq[ctr] = &Item{
			value:    i,
			priority: searchResult.At(i, 0),
			index: ctr,
		}
		ctr++
	}

	heap.Init(&pq)

	maxLength := bm.MaxResults
	if pq.Len() < maxLength {
		maxLength = pq.Len()
	}

	docs := make([]int, maxLength)

	for i := 0; i < maxLength; i++ {
		item := heap.Pop(&pq).(*Item)
		docs[i] = item.value
	}

	return docs
}

func (bm *Bm25) SearchFromString(query string, f SplitFunc) []int {
	return bm.Search(f(query))
}

func (bm *Bm25) buildSearchVec(terms []string) *sparse.CSC {
	t1 := make(map[string]bool)

	for _, st := range terms {
		t1[st] = true
	}

	cleanedSearchTerms := make([]string, len(t1))
	ctr := 0
	for term, _ := range t1 {
		cleanedSearchTerms[ctr] = strings.ToLower(term)
		ctr++
	}

	searchVec := sparse.NewDOK(len(bm.Terms), 1).ToCSC()

	for i, term := range bm.Terms {
		for _, st := range cleanedSearchTerms {
			if term == st {
				searchVec.Set(i, 0, float64(1))
			}
		}
	}
	return searchVec
}

func avgDocumentLength(documents []*Document) (avg float64) {
	for _, document := range documents {
		avg += float64(len(document.Terms))
	}
	return avg/float64(len(documents))
}

type documentWrapper struct {
	ctr int
	d *Document
}