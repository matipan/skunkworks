package main

import (
	"context"
	"encoding/json"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/invopop/jsonschema"
	"go.opentelemetry.io/otel/codes"
)

// An Anthropic model name
type ModelName = string

func (m Anthropic) claudeQuery(ctx context.Context) (msg *anthropic.Message, rerr error) {
	ctx, span := Tracer().Start(ctx, "[🤖] 💭")
	defer func() {
		if rerr != nil {
			span.SetStatus(codes.Error, rerr.Error())
		}
		span.End()
	}()
	key, err := m.APIKey.Plaintext(ctx)
	if err != nil {
		return nil, err
	}
	client := anthropic.NewClient(
		option.WithAPIKey(key),
	)

	tools := []anthropic.ToolParam{
		{
			Name:        anthropic.F("dagger"),
			Description: anthropic.F("Execute a dagger script. <prerequisite>read the dagger manual</prerequisite>"),
			InputSchema: anthropic.F(generateSchema[daggerCommand]()),
		},
	}

	for _, manual := range m.Sandbox.Manuals {
		tools = append(tools, anthropic.ToolParam{
			Name:        anthropic.F(manual.Key),
			Description: anthropic.F(manual.Description),
			InputSchema: anthropic.F(generateSchema[question]()),
		})
	}

	return client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model:     anthropic.F(m.Model),
		MaxTokens: anthropic.Int(2048),
		Messages:  anthropic.F(m.loadHistory(ctx)),
		Tools:     anthropic.F(tools),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(m.SystemPrompt),
		}),
	})
}

func (m Anthropic) loadHistory(ctx context.Context) []anthropic.MessageParam {
	if m.HistoryJSON == "" {
		return nil
	}
	history := []anthropic.MessageParam{}
	err := json.Unmarshal([]byte(m.HistoryJSON), &history)
	if err != nil {
		panic(err)
	}
	return history
}

func (m Anthropic) saveHistory(history []anthropic.MessageParam) Anthropic {
	data, err := json.Marshal(history)
	if err != nil {
		panic(err)
	}
	m.HistoryJSON = string(data)
	return m
}

type question struct {
	Question string `json:"question"`
}

type daggerCommand struct {
	Command string `json:"command"`
}

func generateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}
