package disk

import (
	"crypto/sha256"
	"encoding/json"
	"golang.org/x/mod/sumdb/dirhash"
	"io"
	"log"
	"os"
	"path/filepath"
	"search/model"
	"search/search"
	"strings"
)

func GetDocuments(rootDir string) ([]*search.Document, error) {
	documentPaths := make([]string, 0)

	if !FileExists(rootDir) {
		return nil, ErrNotFound
	}

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

	docList := make([]*search.Document, len(documentPaths))

	for i, path := range documentPaths {
		file, err := os.OpenFile(path, os.O_RDWR, os.ModeTemporary)
		if err != nil {
			panic(err)
		}
		dec := json.NewDecoder(file)

		doc := model.Document{}
		err = dec.Decode(&doc)
		if err != nil {
			return nil, err
		}

		_ = file.Close()

		docList[i] = search.NewDocumentFromString(strings.Join(doc.Body, " "), doc.ID, i, search.SplitAtSpace)
	}

	return docList, nil
}

func GetDirHash(rootDir string) (string, error) {
	filenames, err := dirhash.DirFiles(rootDir, rootDir)
	if err != nil {
		return "", err
	}

	hash, err := dirhash.Hash1(filenames, func(s string) (closer io.ReadCloser, err error) {
		f, err := os.Open(s)
		if err != nil {
			return nil, err
		}
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}
		h1 := HashReadCloser{}
		_, _ = h1.Write([]byte(h.Sum(nil)))
		return &h1, nil
	})
	if err != nil {
		return "", err
	}

	return hash, nil
}

type HashReadCloser struct {
	data []byte
	at int
}

func (h *HashReadCloser) Write(p []byte) (n int, err error) {
	h.data = append(h.data, p...)
	return len(p), nil
}

func (h *HashReadCloser) Read(p []byte) (n int, err error) {
	i := 0
	for ; i < len(p); i++ {
		if h.at + i >= len(h.data) {
			return i-1, io.EOF
		}
		p[i] = h.data[h.at + i]
	}
	h.at += i
	return i, nil
}

func (h *HashReadCloser) Close() error {
	return nil
}