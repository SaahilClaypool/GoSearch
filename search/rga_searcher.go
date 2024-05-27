package search

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type RgaConfig struct {
	Exe string
}

func (rc *RgaConfig) Search(query string) (SearchResults, error) {
	r := SearchResults{
		Results: []SearchResult{},
	}
	es, err := rc.call(query)
	if err != nil {
		return r, err
	}
	var cur *SearchResult = nil
	for _, v := range es {
		if v.Type == "begin" {
			if cur != nil {
				r.Results = append(r.Results, *cur)
			}
			cur = &SearchResult{Title: v.Data.Path.Text, Url: v.Data.Path.Text, Summary: ""}
			continue
		}
		assert(cur != nil, "Should have started with begin")
		if v.Type == "context" || v.Type == "match" {
			cur.Summary += v.Data.Lines.Text
			continue
		}
		if cur != nil {
			r.Results = append(r.Results, *cur)
		}
		cur = nil
	}
	return r, nil
}

func (rc *RgaConfig) call(query string) ([]rgEvent, error) {
	cmd := exec.Command(rc.Exe, "--json", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	strout := string(output)
	es, err := parseEvent(strout)
	return es, nil
}

func parseEvent(val string) ([]rgEvent, error) {
	lines := strings.Split(val, "\n")
	events := make([]rgEvent, 0, len(lines))
	for _, line := range strings.Split(val, "\n") {
		var e rgEvent
		err := json.Unmarshal([]byte(line), &e)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

type rgEvent struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines,omitempty"`
		LineNumber     *int       `json:"line_number,omitempty"`
		AbsoluteOffset int        `json:"absolute_offset,omitempty"`
		Submatches     []struct{} `json:"submatches,omitempty"`
	} `json:"data"`
}

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
