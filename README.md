# GoSearch

## Future ideas

- use RGA <https://github.com/phiresky/ripgrep-all> to search all documents in directory & make a quick response
  - use api keys to do whisper inference and parsing of videos

- use llm configured at the url to make chat completions
  - small llm and big llm

Desired behavior

- llm "question"

-> search with rga
-> pass to small LLM and pick relevent bits
-> pass to big LLM just the relevant bits
-> answer question

- audio -> index
- video -> index w/ screenshot
- scrape utility

needs:

- ripgrepall
- whisper.cpp
- (maybe) it runs docker automatically
  - idea: just ship a wrapper that calls docker
