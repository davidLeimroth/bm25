package config

import (
	"encoding/json"
	"io"
	"os"
)

type SearchConfig struct {
	DataDir string `json:"dataDir"`
	ModelDir string `json:"modelDir"`
	ModelName string `json:"modelName"`
	StateDir string `json:"stateDir"`
	Debug bool `json:"debug"`
	RebuildOnChange bool `json:"rebuildOnChange"`
}

func NewSearchConfig(r io.Reader) (c *SearchConfig, err error) {
	c = &SearchConfig{}
	dec := json.NewDecoder(r)
	err = dec.Decode(c)
	return c, err
}

func NewSearchConfigFromFile(filename string)  (c *SearchConfig, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewSearchConfig(f)
}