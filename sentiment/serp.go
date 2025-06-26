package sentiment

import (
	"fmt"
	"os"
	"strings"

	g "github.com/serpapi/google-search-results-golang"
)

type SerpApi struct {
	api_key string
}

func NewSerpApi() (*SerpApi, error) {
	api_key, is_present := os.LookupEnv("SEARCH_KEY")

	if !is_present {
		return nil, fmt.Errorf("Missing env var SEARCH_KEY")
	}

	return &SerpApi{
		api_key: api_key,
	}, nil
}

func (s *SerpApi) search(ticker string) (string, error) {
	parameter := map[string]string{
		"q":             fmt.Sprintf("%s stock news", ticker),
		"google_domain": "google.com",
	}

	search := g.NewGoogleSearch(parameter, s.api_key)
	results, err := search.GetJSON()
	if err != nil {
		return "", err
	}

	org_results, ok := results["organic_results"].([]interface{})
	if !ok || len(results) == 0 {
		return "No recent news found.", nil
	}

	var snippets []string
	for _, result := range org_results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := resultMap["title"].(string)
		snippet, _ := resultMap["snippet"].(string)
		link, _ := resultMap["link"].(string)

		formatted := fmt.Sprintf("- %s\n%s\n(%s)", title, snippet, link)
		snippets = append(snippets, formatted)

		if len(snippets) == 5 {
			break
		}
	}

	if len(snippets) == 0 {
		return "No recent news found.", nil
	}

	return strings.Join(snippets, "\n"), nil
}
