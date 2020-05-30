package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"search/adapter/disk"
	"search/config"
	"search/model/state"
	"search/search"
	"time"
)

func main() {
	// Load the config
	conf, err := config.NewSearchConfigFromFile("./.config/search_config.json")
	if err != nil {
		panic(fmt.Errorf("couldn't read the config: %v", err))
	}

	// Load the app state
	appStateFilename := filepath.Join(conf.StateDir, "app_state.bin")
	appState := state.NewAppState()
	_ = appState.LoadFromFile(appStateFilename)

	searchFilename := filepath.Join(conf.ModelDir, conf.ModelName)
	rootDir := conf.DataDir

	err = os.MkdirAll(".data/", os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("couldn't create the .data/ directory: %v", err))
	}
	// load the documents
	documents, err := disk.GetDocuments(rootDir)
	if err != nil {
		fmt.Println(fmt.Errorf("the requested data directory does not exist: %v", err))
		os.Exit(1)
	}

	log.Printf("There are %v documents", len(documents))

	h, err := disk.GetDirHash(rootDir)
	if err != nil {
		fmt.Println(fmt.Errorf("couln't create the directories hash: %v", err))
	}

	bm25 := search.NewBm25(0.75, 1.75)
	if disk.FileExists(searchFilename) &&
		len(documents) == appState.NumberOfDocuments && appState.NumberOfDocuments > 0 &&
		h == appState.Hash && appState.Hash != "" {
		// load the old matrix
		f, err := os.Open(searchFilename)
		if err != nil {
			panic(err)
		}
		fmt.Println("Loading the search index from", searchFilename)
		t1 := time.Now()
		if err := bm25.Load(f, documents); err != nil {
			panic(err)
		}
		fmt.Printf("It took %.3f seconds\n", time.Since(t1).Seconds())
		_ = f.Close()
	} else {
		// bm25.UseTfIdf = true
		fmt.Println("Building a new index")
		t1 := time.Now()
		bm25.Build(documents)
		fmt.Printf("It took %.3f seconds\n", time.Since(t1).Seconds())

		// save the current state to disk
		f, err := os.Create(searchFilename)
		if err != nil {
			fmt.Println(fmt.Errorf("could not create the save file: %v", err))
		} else {
			fmt.Println("Saving the new search index")
			t1 = time.Now()
			if err = bm25.Save(f); err != nil {
				fmt.Println(fmt.Errorf("couldn't write the current matrix to disk': %v", err))
			}
			fmt.Printf("It took %.3f seconds\n", time.Since(t1).Seconds())
		}

	}

	appState.NumberOfDocuments = len(documents)
	if h != "" {
		appState.Hash = h
	}

	if err = appState.SaveToFile(appStateFilename); err != nil {
		fmt.Println(fmt.Errorf("couldn't save the app's state to disk: %v", err))
	}

	query := "test hello neuregelung: umsatzsteuer"

	fmt.Printf("searching for: '%v'\n", query)

	t1 := time.Now()
	results := bm25.SearchFromString(query, search.SplitAtSpace)
	fmt.Printf("It took %.5f seconds\n", time.Since(t1).Seconds())
	fmt.Println("Results:")
	for i, documentIndex := range results {
		fmt.Printf("\t%v.\t%v\n", i+1, documents[documentIndex].ID)
	}
}
