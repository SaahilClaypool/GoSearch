package serper

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/saahilclaypool/GoSearch/search"
)

type Serper struct {
	apiKey string
}

func CreateSearcher(apiKey string) Serper {
	return Serper{apiKey: apiKey}
}

func (s Serper) Search(query string) (search.SearchResults, error) {
	results := search.SearchResults{
		Query: query,
	}
	url := "https://google.serper.dev/search"
	q, err := json.Marshal(SearchParameters{Q: &query})
	if err != nil {
		return results, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(q))
	if err != nil {
		return results, err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-API-KEY", s.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return results, err
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error in fetch: %s", err)
		}
		log.Printf("Serper error: %s\n%s", resp.Status, b)
		return results, errors.New(resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return results, err
	}
	var root RootObject
	err = json.Unmarshal([]byte(b), &root)
	if err != nil {
		return results, err
	}

	for _, r := range root.Organic {
		d := r.Description
		if d == nil || *d == "" {
			d = r.Snippet
		}
		results.Results = append(results.Results, search.SearchResult{Url: *r.Link, Title: *r.Title, Summary: *d})
	}
	return results, err
}

type SearchParameters struct {
	Q      *string `json:"q"`
	// Type   *string `json:"type"`
	// Engine *string `json:"engine"`
}

type SearchInformation struct {
	DidYouMean *string `json:"idDidYouMean"`
}

type AnswerBox struct {
	Snippet            *string   `json:"snippet"`
	SnippetHighlighted []*string `json:"snippetHighlighted"`
	Title              *string   `json:"title"`
	Link               *string   `json:"link"`
}

type OrganicResult struct {
	Title       *string     `json:"title"`
	Link        *string     `json:"link"`
	Description *string     `json:"description"`
	Snippet     *string     `json:"snippet"`
	Date        *string     `json:"date"`
	Sitelinks   []*Sitelink `json:"sitelinks"`
	Position    *int        `json:"position"`
}

type Sitelink struct {
	Title *string `json:"title"`
	Link  *string `json:"link"`
}

type Attributes struct {
	Duration *string `json:"duration"`
	Posted   *string `json:"posted"`
}

type RootObject struct {
	SearchParameters  *SearchParameters  `json:"searchParameters"`
	SearchInformation *SearchInformation `json:"searchInformation"`
	AnswerBox         *AnswerBox         `json:"answerBox"`
	Organic           []*OrganicResult   `json:"organic"`
	PeopleAlsoAsk     []*PeopleAlsoAsk   `json:"peopleAlsoAsk"`
	RelatedSearches   []*RelatedSearches `json:"relatedSearches"`
}

type PeopleAlsoAsk struct {
	Question *string `json:"question"`
	Snippet  *string `json:"snippet"`
	Title    *string `json:"title"`
	Link     *string `json:"link"`
}

type RelatedSearches struct {
	Query *string `json:"query"`
}
