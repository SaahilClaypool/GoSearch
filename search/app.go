package search

import (
	"fmt"
)

var gid int = 0
var convos map[int]*Conversation = make(map[int]*Conversation)

func CreateApp(llm LLM, searcher Searcher) App {
	return App{
		searcher: searcher,
		llm:      llm,
	}
}

type App struct {
	searcher Searcher
	llm      LLM
}

type Conversation struct {
	Id     int
	Rounds *[]Round
}

type Round struct {
	UserInput     string
	HasSearch     bool
	SearchResults SearchResults
	Response      string
}

func (c *App) CreateConversation() *Conversation {
	id := gid
	gid++
	rounds := make([]Round, 0, 1)
	convos[id] = &Conversation{
		Id:     id,
		Rounds: &rounds,
	}
	return convos[id]
}

func (ctx *App) GetConversation(id int) *Conversation {
	return convos[id]
}

func (c App) Reply(id int, input string) (string, *Conversation, error) {
	convo := convos[id]
	round := Round{UserInput: input}
	query, err := c.getQuery(convo, input)
	if err != nil {
		return "", nil, err
	}
	fmt.Printf("query: %s\n", query)
	searchResults, err := c.searcher.Search(query)
	if err != nil {
		return "", nil, err
	}
	round.HasSearch = true
	round.SearchResults = searchResults
	prompt, err := renderT("summarize.txt", struct {
		Context Conversation
		Round   Round
	}{Context: *convo, Round: round})
	if err != nil {
		return "", nil, err
	}
	*convo.Rounds = append(*convo.Rounds, round)
	response, err := c.llm.Complete(prompt)
	if err != nil {
		return "", nil, err
	}
	return response, nil, nil
}

func (c *App) getQuery(convo *Conversation, input string) (string, error) {

	ctx, err := convo.format()
	if err != nil {
		return "", err
	}
	if len(*convo.Rounds) == 0 {
		ctx = ""
	}
	data := struct {
		Input           string
		PreviousMessage string
	}{
		Input:           input,
		PreviousMessage: ctx,
	}
	s, err := renderT("get_query.txt", data)
	if err != nil {
		return "", err
	}
	r := struct{ Query string }{}
	c.llm.GetJson(s, &r)

	return r.Query, nil

}

func (conversation Conversation) format() (string, error) {
	return renderT("conversation.txt", conversation)
}
