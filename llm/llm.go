package llm

import (
	"github.com/kardolus/chatgpt-cli/client"
	"github.com/kardolus/chatgpt-cli/config"
	"github.com/kardolus/chatgpt-cli/configmanager"
	"github.com/kardolus/chatgpt-cli/history"
	"github.com/kardolus/chatgpt-cli/http"
	"github.com/kardolus/maps/utils"
	"regexp"
	"strings"
)

const (
	promptFile = "query_prompt.txt"
	filterFile = "filter_prompt.txt"
	inputQuery = "input query: "
)

//go:generate mockgen -destination=clientmocks_test.go -package=llm_test github.com/kardolus/maps/llm LLMClient
type LLMClient interface {
	ProvideContext(context string)
	Query(input string) (string, int, error)
	Stream(input string) error
	ListModels() ([]string, error)
	WithContextWindow(window int) *client.Client
	WithServiceURL(url string) *client.Client
}

// Ensure client.Client implements LLMClient interface
var _ LLMClient = &client.Client{}

//go:generate mockgen -destination=readermocks_test.go -package=llm_test github.com/kardolus/maps/llm FileReader
type FileReader interface {
	FileToBytes(fileName string) ([]byte, error)
}

// Ensure Utils implements FileReader interface
var _ FileReader = &utils.Utils{}

type LLM struct {
	client     LLMClient
	fileReader FileReader
}

func New(client LLMClient, fileReader FileReader) *LLM {
	return &LLM{
		client:     client,
		fileReader: fileReader,
	}
}

func NewChatGPTClient() (*client.Client, error) {
	hs, _ := history.New() // do not error out
	return client.New(http.RealCallerFactory, config.New(), hs, false)
}

func (l *LLM) ClearHistory() error {
	cm := configmanager.New(config.New())

	return cm.DeleteThread(cm.Config.Thread)
}

func (l *LLM) GenerateSubQueries(query string) ([]string, error) {
	bytes, err := l.fileReader.FileToBytes(promptFile)
	if err != nil {
		return nil, err
	}

	l.client.ProvideContext(string(bytes))

	response, _, err := l.client.Query(inputQuery + query)
	if err != nil {
		return nil, err
	}

	return extractSearchQueries(response), nil
}

// GenerateFilter will extract the 'contains' and 'matches' strings from the LLM's response
func (l *LLM) GenerateFilter(query string) ([]string, []string, error) {
	bytes, err := l.fileReader.FileToBytes(filterFile)
	if err != nil {
		return nil, nil, err
	}

	l.client.ProvideContext(string(bytes))

	response, _, err := l.client.Query(inputQuery + query)
	if err != nil {
		return nil, nil, err
	}

	contains, matches := extractContainsAndMatches(response)

	return contains, matches, nil
}

// extractContainsAndMatches will extract the 'contains' and 'matches' strings using regex
func extractContainsAndMatches(input string) ([]string, []string) {
	var containsList, matchesList []string

	// Define regex patterns to match the 'contains' and 'matches' lines
	containsRegex := regexp.MustCompile(`(?i)contains:\s*([^\n]+)`)
	matchesRegex := regexp.MustCompile(`(?i)matches:\s*([^\n]+)`)

	// Search for the 'contains' line
	containsMatches := containsRegex.FindStringSubmatch(input)
	if len(containsMatches) > 1 {
		containsList = strings.Split(strings.TrimSpace(containsMatches[1]), ",")
		for i, item := range containsList {
			containsList[i] = strings.TrimSpace(item)
		}
	}

	// Search for the 'matches' line
	matchesMatches := matchesRegex.FindStringSubmatch(input)
	if len(matchesMatches) > 1 {
		matchesList = strings.Split(strings.TrimSpace(matchesMatches[1]), ",")
		for i, item := range matchesList {
			matchesList[i] = strings.TrimSpace(item)
		}
	}

	return containsList, matchesList
}

func extractSearchQueries(input string) []string {
	var results []string
	regex := regexp.MustCompile("search\\s*\\[\\d+\\]:\\s*(.*)")

	lines := strings.Split(strings.TrimSpace(input), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		matches := regex.FindStringSubmatch(line)
		if len(matches) > 1 {
			results = append(results, matches[1])
		}
	}

	return results
}
