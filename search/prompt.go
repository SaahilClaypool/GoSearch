package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

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
	templ := template.New("").Funcs(
		template.FuncMap{
			"toJson": func(v any) string {
				b, err := json.Marshal(v)
				if err != nil {
					log.Printf("error marshaling to json :%v\n", err)
					return ""
				}
				return string(b)
			},
		},
	)
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

type JRequest[I, T any] struct {
	Overview string
	Examples []JEx[I, T]
	Req      JEx[I, T]
}

type JEx[I, T any] struct {
	Directions *string
	Input      I
	Output     *T
}

func makeJsonRequest[I, T any](overview string, examples []JEx[I, T], inp JEx[I, T]) JRequest[I, T] {
	if examples == nil {
		examples = make([]JEx[I, T], 0)
	}
	return JRequest[I, T]{
		Overview: overview,
		Examples: examples,
		Req:      inp,
	}
}

func LLMJson[I, T any](llm LLM, overview string, req I, ex []JEx[I, T]) (T, error) {
	var outputStruct = JEx[I, T]{
		Input: req,
	}
	jreq := makeJsonRequest(overview, ex, outputStruct)
	var output T
	prompt, err := renderT("json_request.txt", jreq)
	if err != nil {
		return output, err
	}

	err = llm.GetJson(prompt, &output)
	if err != nil {
		return output, err
	}
	return output, nil
}
