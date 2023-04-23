package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks     string
	SuffixArray       *suffixarray.Index
	ChapterMap        map[int]string
	ChapterIndexArray []int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	s.organizeTableOfContents()
	return nil
}

func (s *Searcher) organizeTableOfContents() {
	results := s.SearchContents()
	idx := results[0]
	contentString := s.CompleteWorks[idx : idx+1970]

	s.getChapters(contentString)
}

func (s *Searcher) SearchContents() []int {
	idxs := s.SuffixArray.Lookup([]byte("Contents"), -1)
	return idxs
}

func (s *Searcher) getChapters(contents string) {

	titles := strings.Split(contents, "\n")
	tempMap := make(map[int]string)
	tempArray := []int{}
	for _, title := range titles {
		title = strings.TrimSpace(title)

		if title != "" && title != "Contents" {
			index := 1
			if title == "THE TRAGEDY OF ANTONY AND CLEOPATRA" {
				title = "ANTONY AND CLEOPATRA"
			} else if title == "THE LIFE OF KING HENRY THE FIFTH" {
				title = "THE LIFE OF KING HENRY V"
				index = 0
			} else if title == "THE TRAGEDY OF MACBETH" {
				title = "MACBETH"
			} else if title == "THE TRAGEDY OF OTHELLO, MOOR OF VENICE" {
				title = "OTHELLO, THE MOOR OF VENICE"
				index = 0
			} else if title == "TWELFTH NIGHT; OR, WHAT YOU WILL" {
				title = "TWELFTH NIGHT: OR, WHAT YOU WILL"
				index = 0
			}
			curr := s.getChapter(title, index)
			tempMap[curr] = title
			tempArray = append(tempArray, curr)
		}
	}
	s.ChapterMap = tempMap
	sort.Ints(tempArray)
	s.ChapterIndexArray = tempArray
}

func (s *Searcher) getChapter(title string, index int) int {
	idxs := s.SuffixArray.Lookup([]byte(title), -1)
	sort.Ints(idxs)
	return idxs[index]
}

func (s *Searcher) binarySearch(needle int) string {

	low := 0
	high := len(s.ChapterIndexArray) - 1

	for low <= high {
		median := (low + high) / 2

		if s.ChapterIndexArray[median] < needle {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == 0 {
		return "Pre-chapter"
	}
	index := s.ChapterIndexArray[low-1]
	return s.ChapterMap[index]
}

func (s *Searcher) getAllIdxs(query string) []int {
	query_uppercase := strings.ToUpper(query)
	query_lowercase := strings.ToLower(query)
	query_capitalize := strings.Title(query_lowercase)

	idxs := s.SuffixArray.Lookup([]byte(query_lowercase), -1)
	upperIdxs := s.SuffixArray.Lookup([]byte(query_uppercase), -1)
	capitalIdxs := s.SuffixArray.Lookup([]byte(query_capitalize), -1)
	idxs = append(idxs, upperIdxs...)
	idxs = append(idxs, capitalIdxs...)
	sort.Ints(idxs)

	return idxs
}

func (s *Searcher) Search(query string) [][]string {

	idxs := s.getAllIdxs(query)

	results := [][]string{}

	for _, idx := range idxs {
		_, ok := s.ChapterMap[idx] //is the actual contents title
		if ok || idx < s.ChapterIndexArray[0] {
			continue
		}

		currExcerpt := s.CompleteWorks[idx-250 : idx+250]
		chapter := s.binarySearch(idx)
		row := []string{chapter, currExcerpt}
		results = append(results, row)
	}
	return results
}
