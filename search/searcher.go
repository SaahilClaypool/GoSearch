package search

type Searcher interface {
	Search(query string) (SearchResults, error)
}

type SearchResults struct {
    Query string
	Results []SearchResult
}

type SearchResult struct {
	Url     string
	Title   string
	Summary string
}
