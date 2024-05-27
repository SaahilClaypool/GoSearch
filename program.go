package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/saahilclaypool/GoSearch/search"
)

type Output struct {
	V string `json:"output"`
}

func main() {
	llm := search.CreateLLM("https://api.groq.com/openai/v1", os.Getenv("GROQ_API_KEY"), "llama3-70b-8192", "Exclude all introductions. Answer concisely and use well-formatted markdown.")
	// searcher := serper.CreateClient(os.Getenv("SERPER_API_KEY"))
	searcher := search.RgaConfig{}
	converser := search.CreateApp(llm, &searcher)
	o, _ := search.LLMJson(llm, "add the word apple to the end of each input",
		"This is a longer sentence",
		[]search.JEx[string, Output]{
			{Input: "test", Output: &Output{V: "testapple"}},
			{Input: "this is two", Output: &Output{V: "this is twoapple"}},
		})
	fmt.Printf("output is: %s\n", o.V)
	if len(os.Args) < 2 {
		fmt.Println("TODO: start server")
		return
	}
	if os.Args[1] == "test" {
		// test(converser)
		_ = converser
	}
}

func test(conv search.App) {
	c := conv.CreateConversation()
	prompt := "> "
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break // blank string was entered
		}
		r, _, err := conv.Reply(c.Id, line)
		if err != nil {
			fmt.Printf("Error in reply: %v\n", err)
			break
		}
		fmt.Printf("====================\nQ: %s\nA: %s\n====================\n\n", line, r)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}

func startServer() bool {
	fmt.Println("Starting api server...")
	r := mux.NewRouter()
	r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s\n", fmt.Sprintf("%s %s", r.Method, r.URL))
			h.ServeHTTP(w, r)
		})
	})
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	http.ListenAndServe(":8011", r)
	return false
}
