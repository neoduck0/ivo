package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/neoduck0/ivo/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type role string

const (
	roleUser      role = "user"
	roleAssistant role = "assistant"
)

const systemPrompt = `
You are a helpful assistant.
`

type message struct {
	role    role
	content string
}

// Conversation retains successful text turns for one interactive session.
// It is not safe for concurrent use.
type Conversation struct {
	client   openai.Client
	model    string
	messages []message
}

// NewConversation creates an empty conversation using the provider configuration.
func NewConversation(cfg config.Config) *Conversation {
	return &Conversation{
		client: openai.NewClient(
			option.WithAPIKey(cfg.APIKey),
			option.WithBaseURL(cfg.BaseURL),
		),
		model: cfg.Model,
	}
}

// SendMessage sends a user message with the conversation history and returns
// the reply. Only successful turns are retained; requests time out after a minute.
func (c *Conversation) SendMessage(message string) (string, error) {
	ctx, done := context.WithTimeout(context.Background(), time.Minute)
	defer done()
	return c.sendMessage(ctx, message)
}

func (c *Conversation) sendMessage(ctx context.Context, userMessage string) (string, error) {
	input := make(responses.ResponseInputParam, 0, len(c.messages)+1)
	for _, previous := range c.messages {
		input = append(input, responses.ResponseInputItemParamOfMessage(
			previous.content, responses.EasyInputMessageRole(previous.role),
		))
	}
	input = append(input, responses.ResponseInputItemParamOfMessage(
		userMessage, responses.EasyInputMessageRole(roleUser),
	))

	response, err := c.client.Responses.New(ctx, responses.ResponseNewParams{
		Model:        c.model,
		Instructions: openai.String(systemPrompt),
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: input,
		},
	})
	if err != nil {
		return "", err
	}
	if response == nil {
		return "", fmt.Errorf("response is null")
	}

	// Only completed generations are committed. Discard partial output so a
	// retry cannot inherit a truncated assistant turn.
	switch response.Status {
	case responses.ResponseStatusCompleted:
	case responses.ResponseStatusIncomplete:
		reason := response.IncompleteDetails.Reason
		if reason == "" {
			reason = "unknown reason"
		}
		return "", fmt.Errorf("response incomplete: %s", reason)
	default:
		details := string(response.Error.Code)
		if response.Error.Message != "" {
			if details != "" {
				details += ": "
			}
			details += response.Error.Message
		}
		if details != "" {
			return "", fmt.Errorf("response status %q: %s", response.Status, details)
		}
		return "", fmt.Errorf("response status: %q", response.Status)
	}

	// OutputText omits structured refusals. Include both kinds of visible
	// message content in their original order for display and history.
	var reply strings.Builder
	for _, item := range response.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			switch content.Type {
			case "output_text":
				reply.WriteString(content.Text)
			case "refusal":
				reply.WriteString(content.Refusal)
			}
		}
	}
	text := reply.String()
	// Commit the turn only after the request succeeds, so retries do not
	// duplicate failed user messages in the conversation.
	c.messages = append(c.messages,
		message{role: roleUser, content: userMessage},
		message{role: roleAssistant, content: text},
	)
	return text, nil
}
