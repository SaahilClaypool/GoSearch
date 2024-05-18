package search

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var gid int = 0
var convos map[int]*Conversation = make(map[int]*Conversation)

func CreateConverser(llm LLM, searcher Searcher) Converser {
	return Converser{
		searcher: searcher,
		llm:      llm,
	}
}

type Converser struct {
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

func (c *Converser) CreateConversation() *Conversation {
	id := gid
	gid++
	rounds := make([]Round, 0, 1)
	convos[id] = &Conversation{
		Id:     id,
		Rounds: &rounds,
	}
	return convos[id]
}

func (ctx *Converser) GetConversation(id int) *Conversation {
	return convos[id]
}

func (c Converser) Reply(id int, input string) (string, *Conversation, error) {
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
	prompt, err := renderT("reply.txt", struct {
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

func (c *Converser) getQuery(convo *Conversation, input string) (string, error) {

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
	c.llm.GetResult(s, &r)

	return r.Query, nil

}

func (conversation Conversation) format() (string, error) {
	return renderT("conversation.txt", conversation)
}

var templates *template.Template

func renderT(templateName string, data any) (string, error) {
	if templates == nil {
		templates = parseTemplates()
	}
	var tpl bytes.Buffer
	if err := templates.ExecuteTemplate(&tpl, templateName, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func parseTemplates() *template.Template {
	templ := template.New("")
	err := filepath.Walk("./search/prompts", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".txt") {
			_, err := templ.ParseFiles(path)
			if err != nil {
				log.Println(err)
				panic("exit")
			}
		}

		return err
	})

	fmt.Println(templ.DefinedTemplates())
	if err != nil {
		panic(err)
	}

	return templ
}
