package serpapi

// This code was substantially borrowed with appreciation from
// github.com/tmc/langchaingo/tools/serpapi
// and modified to fit the needs of the project.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/chriscow/minds"
)

var (
	ErrMissingToken = errors.New("missing the SerpAPI API key, set it in the SERPAPI_API_KEY environment variable")
	ErrNoGoodResult = errors.New("no good search results found")
	ErrAPIError     = errors.New("error from SerpAPI")
)

func New() (minds.Tool, error) {
	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingToken
	}

	return minds.WrapFunction(
		"google_search",
		`Performs a Google Search. Useful for when you need to answer questions 
		about current events. Always one of the first options when you need to 
		find information on internet. Input should be a search query.`,
		struct {
			Input string `json:"input" description:"A detailed search query."`
		}{},
		serpSearch,
	)
}

func serpSearch(ctx context.Context, args []byte) ([]byte, error) {
	var params struct {
		Input string `json:"input"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}

	result, err := query(ctx, params.Input)
	if err != nil {
		if errors.Is(err, ErrNoGoodResult) {
			return []byte("No relevant Google Search results were found"), nil
		}

		return []byte(""), err
	}

	return []byte(strings.Join(strings.Fields(result), " ")), nil
}

func query(ctx context.Context, query string) (string, error) {
	const _url = "https://serpapi.com/search"
	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		return "", ErrMissingToken
	}

	params := make(url.Values)
	query = strings.ReplaceAll(query, " ", "+")
	params.Add("q", query)
	params.Add("google_domain", "google.com")
	params.Add("gl", "us")
	params.Add("hl", "en")
	params.Add("api_key", apiKey)

	reqURL := fmt.Sprintf("%s?%s", _url, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request in serpapi: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing response in serpapi: %w", err)
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return "", fmt.Errorf("coping data in serpapi: %w", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return "", fmt.Errorf("unmarshal data in serpapi: %w", err)
	}

	return processResponse(result)
}

func processResponse(res map[string]interface{}) (string, error) {
	if errorValue, ok := res["error"]; ok {
		return "", fmt.Errorf("%w: %v", ErrAPIError, errorValue)
	}
	if res := getAnswerBox(res); res != "" {
		return res, nil
	}
	if res := getSportResult(res); res != "" {
		return res, nil
	}
	if res := getKnowledgeGraph(res); res != "" {
		return res, nil
	}
	if res := getOrganicResult(res); res != "" {
		return res, nil
	}

	return "", ErrNoGoodResult
}

func getAnswerBox(res map[string]interface{}) string {
	answerBox, answerBoxExists := res["answer_box"].(map[string]interface{})
	if answerBoxExists {
		if answer, ok := answerBox["answer"].(string); ok {
			return answer
		}
		if snippet, ok := answerBox["snippet"].(string); ok {
			return snippet
		}
		snippetHighlightedWords, ok := answerBox["snippet_highlighted_words"].([]interface{})
		if ok && len(snippetHighlightedWords) > 0 {
			return fmt.Sprintf("%v", snippetHighlightedWords[0])
		}
	}

	return ""
}

func getSportResult(res map[string]interface{}) string {
	sportsResults, sportsResultsExists := res["sports_results"].(map[string]interface{})
	if sportsResultsExists {
		if gameSpotlight, ok := sportsResults["game_spotlight"].(string); ok {
			return gameSpotlight
		}
	}

	return ""
}

func getKnowledgeGraph(res map[string]interface{}) string {
	knowledgeGraph, knowledgeGraphExists := res["knowledge_graph"].(map[string]interface{})
	if knowledgeGraphExists {
		if description, ok := knowledgeGraph["description"].(string); ok {
			return description
		}
	}

	return ""
}

func getOrganicResult(res map[string]interface{}) string {
	organicResults, organicResultsExists := res["organic_results"].([]interface{})

	if organicResultsExists && len(organicResults) > 0 {
		organicResult, ok := organicResults[0].(map[string]interface{})
		if ok {
			if snippet, ok := organicResult["snippet"].(string); ok {
				return snippet
			}
		}
	}

	return ""
}
