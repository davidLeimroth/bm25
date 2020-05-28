package main

import (
	"encoding/json"
	"fmt"
	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

var BM25 = true

func main() {

	BM25 = true
	// searchMat, termList, _ := buildSearchMatrix("../../data/corpus_jsons/")
	searchMat, termList, _ := buildSearchMatrix("./testdata/")
	// searchMat, termList, _ := buildSearchMatrix("./testdata2/")
	// searchMat, termList, _ := buildSearchMatrix("./corpus_jsons/")

	// docList, err := readFiles("./testdata2")
	// if err != nil {
	// 	fmt.Println("Could not read files", err)
	// 	return
	// }

	searchTerms := []string{"umsatzsteuer", "world", "surfing", "internet"}

	searchVec := buildSearchVec(searchTerms, termList)

	opts := mat.Excerpt(5)
	fmt.Println(mat.Formatted(searchMat.T(), opts))
	fmt.Println(mat.Formatted(searchVec.T(), opts))
	fmt.Println(searchVec.NNZ())

	var searchResult sparse.CSR

	searchResult.Mul(searchMat, searchVec)


	fmt.Println(mat.Formatted(searchResult.T(), opts))
	fmt.Println(searchResult.Dims())

}

func readFiles(rootDir string) ([]Document, error) {
	documentPaths := make([]string, 0)

	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		documentPath, err := filepath.Abs(path)
		if err != nil {
			log.Println(err)
			return err
		}
		documentPaths = append(documentPaths, documentPath)
		return nil
	})
	log.Printf("There are %v documents", len(documentPaths))

	docList := make([]Document, len(documentPaths))

	for i, path := range documentPaths {
		file, err := os.OpenFile(path, os.O_RDWR, os.ModeTemporary)
		if err != nil {
			panic(err)
		}
		dec := json.NewDecoder(file)

		doc := Document{}
		err = dec.Decode(&doc)
		if err != nil {
			return nil, err
		}

		_ = file.Close()
		doc.Ctr = i
		docList[i] = doc
	}

	return docList, nil
}


func buildSearchVec(terms []string, tl []string) *sparse.CSC {
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

	searchVec := sparse.NewDOK(len(tl), 1)

	rows, cols := searchVec.Dims()

	fmt.Printf("seachVec %v x %v\n", rows, cols)

	for i, term := range tl {
		for _, st := range cleanedSearchTerms {
			if term == st {
				searchVec.Set(i, 0, float64(1))
			}
		}
	}
	return searchVec.ToCSC()
}


func buildSearchMatrix(rootDir string) (*sparse.CSR, []string, []string) {

	documentPaths := make([]string, 0)

	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}
		// documentPath, err := filepath.Abs(filepath.Join(path, info.Name()))
		documentPath, err := filepath.Abs(path)
		if err != nil {
			log.Println(err)
			return err
		}
		documentPaths = append(documentPaths, documentPath)
		return nil
	})
	log.Printf("There are %v documents", len(documentPaths))

	invertedLists := WordMap{
		w: make(map[string][]*DocumentWrapper),
	}
	termList := TermList{}

	docLengths := make([]int, len(documentPaths))

	for i, path := range documentPaths {
		file, err := os.OpenFile(path, os.O_RDWR, os.ModeTemporary)
		if err != nil {
			panic(err)
		}
		dec := json.NewDecoder(file)

		doc := Document{}
		err = dec.Decode(&doc)
		if err != nil {
			panic(err)
		}

		file.Close()

		docLengths[i] = len(doc.GetWords())
		doc.Ctr = i

		invertedLists.AddDocument(&doc, (*[]string)(&termList))
	}

	iv := invertedLists.GetInternalMap()

	numberOfWords := len(iv)
	numberOfDocuments := len(documentPaths)


	fmt.Printf("There are %v words in %v documents\n", numberOfWords, numberOfDocuments)
	fmt.Printf("%v * %v  = %v\n", numberOfWords, numberOfDocuments, numberOfWords * numberOfDocuments)

	dokMat := sparse.NewDOK(numberOfDocuments, numberOfWords)

	rows, cols := dokMat.Dims()
	fmt.Println(dokMat.NNZ())
	fmt.Printf("Done %v x %v\n", rows, cols)

	n := float64(len(documentPaths))
	avdl := 0.0
	for _, l := range docLengths {
		avdl += float64(l)
	}
	avdl = avdl/float64(len(docLengths))

	b, k := 0.75, 1.75

	for termId, term := range termList {
		fmt.Println(term)
		for _, doc := range iv[term] {
			if !BM25 {
				dokMat.Set(doc.Doc.Ctr, termId, float64(doc.Ctr))
				continue
			}

			val := float64(doc.Ctr)
			alpha := 1 - b + (b * float64(docLengths[doc.Doc.Ctr]) / avdl)
			tf := 1.0
			if k > 0 {
				tf = val * (1 + (1 / k)) / (alpha + (val / k))
			}
			df := float64(len(iv[term])) // TODO: check the value

			v := tf * math.Log2(n / df)
			dokMat.Set(doc.Doc.Ctr, termId, v)
		}
		if termId % 1000 == 0 {
			fmt.Printf("Finished %v terms\n", termId)
		}
	}

	// Query like: A * searchVector

	return dokMat.ToCSR(), termList, documentPaths

}

type DocumentWrapper struct {
	Doc *Document
	Ctr int
}

type Document struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Body []string `json:"body"`
	Meta struct {
		DocumentType string `json:"ch:documentType"`
	} `json:"meta"`
	Ctr int
}

func (d Document) String() string {
	text := strings.Builder{}
	for _, s := range d.Body {
		_, err := text.WriteString(s + " ")
		if err != nil {
			panic(err)
		}
	}

	return strings.TrimSpace(text.String())
}

func (d Document) GetWords() []string {
	return strings.Split(d.String(), " ")
}

type TermList []string

type WordMap struct {
	w map[string][]*DocumentWrapper
}

func (w WordMap) GetInternalMap() map[string][]*DocumentWrapper {
	return w.w
}

func (w WordMap) AddDocument(d *Document, tl *[]string) {
	words := d.GetWords()
	for _, word := range words {
		word = strings.ToLower(word)
		if _, ok := w.w[word]; !ok {
			w.w[word] = make([]*DocumentWrapper, 1)
			w.w[word][0] = &DocumentWrapper{
				Doc: d,
				Ctr: 0,
			}
			*tl = append(*tl, word)
		}
		if w.w[word][len(w.w[word]) - 1].Doc.ID == d.ID {
			w.w[word][len(w.w[word]) - 1].Ctr++
			continue
		}
		w.w[word] = append(w.w[word], &DocumentWrapper{Doc: d, Ctr:1})
	}
}