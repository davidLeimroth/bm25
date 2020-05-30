package state

import (
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
	"search/adapter/disk"
)

type AppState struct {
	NumberOfDocuments int
	Hash string
}

func NewAppState() *AppState {
	return &AppState{}
}

func (a *AppState) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(a)
}

func (a *AppState) LoadFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return a.Load(f)
}

func (a *AppState) Save(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(a)
}

func (a *AppState) SaveToFile(filename string) (err error) {
	var f *os.File
	if !disk.FileExists(filename) {
		fpath, _ := filepath.Split(filename)
		if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
			return err
		}
		if f, err = os.Create(filename); err != nil {
			return err
		}
	} else {
		if f, err = os.OpenFile(filename, os.O_RDWR, os.ModePerm); err != nil {
			return err
		}
	}
	defer f.Close()
	return a.Save(f)
}