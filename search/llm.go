package search

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type LLM interface {
	Chat(prompt string) (chan (string), error)
	Complete(prompt string) (string, error)
	GetResult(prompt string, obj any) error
}

type OpenAILLM struct {
	baseUrl string
	key     string
	model   string
	system  string
}

func CreateLLM(url string, key string, model string, system string) OpenAILLM {
	return OpenAILLM{
		baseUrl: url,
		key:     key,
		model:   model,
		system:  system,
	}
}

func (llm OpenAILLM) Chat(prompt string) (chan (string), error) {
	url := fmt.Sprintf("%s/chat/completions", llm.baseUrl)
	chat := llm.makePrompt(prompt, RTStream)
	jsonMessage, err := json.Marshal(chat)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", llm.key))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error in fetch: %s", err)
		}
		log.Printf("Error: %s\n%s", resp.Status, b)
		return nil, errors.New(resp.Status)
	}
	events := make(chan string)
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			token := strings.TrimPrefix(line, "data: ")
			if token == "[DONE]" {
				break
			}
			chunk, err := parseDelta(token)
			if err != nil {
				log.Printf("failed to parse chunk %s: %v", token, err)
				break
			}
			events <- chunk
		}
		if err := scanner.Err(); err != nil {
			log.Printf("error reading from SSE stream: %v", err)
		}
		close(events)
		resp.Body.Close()
	}()
	return events, nil
}

func (llm OpenAILLM) makePrompt(prompt string, responseType ResponseType) ChatRequest {
	request := ChatRequest{
		Model:  llm.model,
		Stream: responseType == RTStream,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
	if responseType == RTObject {
		request.ResponseFormat = &Format{
			FormatType: "json_object",
		}
	}
	if llm.system != "" {
		request.Messages = append([]Message{{Role: "system", Content: llm.system}}, request.Messages...)
	}
	return request
}

func (llm OpenAILLM) GetResult(message string, object any) error {
	respBody, err := llm.complete(message, true)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(respBody), &object)
	if err != nil {
		return err
	}
	return nil
}

func (llm OpenAILLM) Complete(message string) (string, error) {
	return llm.complete(message, false)
}

func (llm OpenAILLM) complete(message string, object bool) (string, error) {
	log.Printf("Getting result:\n%s\n", message)
	url := fmt.Sprintf("%s/chat/completions", llm.baseUrl)
	var format ResponseType
	if object {
		format = RTObject
	} else {
		format = RTString
	}
	request := llm.makePrompt(message, format)
	jsonMessage, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
	if err != nil {
		return "", err
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", llm.key))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error in fetch: %v", err)
		}
		log.Printf("HTTP Error: %s\n%s\n", resp.Status, b)
		return "", errors.New(resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("Response result:\n%s\n", string(b))
	respBody, err := parseResponse(string(b))
	if err != nil {
		return "", err
	}
	return respBody, nil
}

func parseDelta(resp string) (string, error) {
	var chunk ChatCompletionChunk
	err := json.Unmarshal([]byte(resp), &chunk)
	if err != nil {
		return "", err
	}
	return chunk.Choices[0].Delta.Content, nil
}

func parseResponse(resp string) (string, error) {
	var chunk ChatCompletion
	err := json.Unmarshal([]byte(resp), &chunk)
	if err != nil {
		return "", err
	}
	return chunk.Choices[0].Message.Content, nil
}

type ChatRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	ResponseFormat *Format   `json:"response_format"`
	Stream         bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Format struct {
	FormatType string `json:"type"`
}

type ChatCompletionChunk struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
}

type Choice struct {
	Index        int         `json:"index"`
	Delta        Delta       `json:"delta"`
	Message      Message     `json:"message"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason interface{} `json:"finish_reason"`
}

type Delta struct {
	Content string `json:"content"`
}

type ChatCompletion struct {
	ID                string   `json:"id"`
	ObjectType        string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ResponseType int

const (
	RTStream ResponseType = iota
	RTString
	RTObject
)
