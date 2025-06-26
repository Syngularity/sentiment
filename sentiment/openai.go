package sentiment

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAi struct {
	client   openai.Client
	web_tool *SerpApi
}

func NewOpenAi() (*OpenAi, error) {
	api_key, is_present := os.LookupEnv("OPENAI_KEY")

	if !is_present {
		return nil, fmt.Errorf("Missing env var OPENAI_KEY")
	}

	client := openai.NewClient(
		option.WithAPIKey(api_key),
	)

	web_tool, err := NewSerpApi()
	if err != nil {
		return nil, err
	}

	return &OpenAi{
		client:   client,
		web_tool: web_tool,
	}, nil
}

func (o *OpenAi) Sentiment(ctx context.Context, ticker string, sse *SSEWriter) {

	webResults, err := o.web_tool.search(ticker)
	if err != nil {
		sse.Error(err)
		return
	}

	systemPrompt := `You are a professional stock market news analyst. 
        Given web search results about a stock, summarize:
        - What the company does (1-2 lines)
        - Today's main catalyst or news moving the stock
        - Sentiment (Bullish, Bearish, Neutral) and why
        - Possible intraday price action or volatility range estimate`

	userPrompt := fmt.Sprintf(`Ticker: %s
        
        Web Search Results:
        %s
        
        Give your analysis:`, ticker, webResults)

	stream := o.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Seed:  openai.Int(0),
		Model: openai.ChatModelGPT4o,
	})
	defer stream.Close()

	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)

		if content, ok := acc.JustFinishedContent(); ok {
			log.Println("Content stream finished:", content)
		}

		// if using tool calls
		if tool, ok := acc.JustFinishedToolCall(); ok {
			log.Println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
		}

		if refusal, ok := acc.JustFinishedRefusal(); ok {
			log.Println("Refusal stream finished:", refusal)
		}

		if len(chunk.Choices) > 0 {
			sse.WriteEvent(chunk.Choices[0].Delta.Content)
		}
	}

	if err := stream.Err(); err != nil {
		sse.Error(err)
	} else {
		sse.Done()
	}
}
