package main

import (
	"context"
	"dagger/anthropic/internal/dagger"
	"encoding/json"
	"regexp"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

func New(
	ctx context.Context,
	// Anthropic API key
	apiKey *dagger.Secret,
	// Anthropic model
	// +optional
	// +default="claude-3-5-sonnet-latest"
	model ModelName,
	// A builtin knowledge library, made of text files.
	// First paragraph is the description. The rest is the contents.
	// +optional
	// +defaultPath=./knowledge
	knowledgeDir *dagger.Directory,
	// A system prompt to inject into Claude's context
	// +optional
	// +defaultPath="./system-prompt.txt"
	systemPrompt *dagger.File,
) (Anthropic, error) {
	anthropic := Anthropic{
		APIKey: apiKey,
		Model:  model,
	}
	sandbox, err := NewSandbox().WithUsername("🤖").ImportManuals(ctx, knowledgeDir)
	if err != nil {
		return anthropic, err
	}
	anthropic.Sandbox = sandbox
	prompt, err := systemPrompt.Contents(ctx)
	if err != nil {
		return anthropic, err
	}
	return anthropic.WithSystemPrompt(ctx, prompt), nil
}

type Anthropic struct {
	Model        ModelName      // +private
	APIKey       *dagger.Secret // +private
	HistoryJSON  string         // +private
	Sandbox      Sandbox
	SystemPrompt string
}

func (m Anthropic) WithSandbox(sandbox Sandbox) Anthropic {
	m.Sandbox = sandbox
	return m
}

func (m Anthropic) WithSecret(name string, value *dagger.Secret) Anthropic {
	m.Sandbox = m.Sandbox.WithSecret(name, value)
	return m
}

func (m Anthropic) WithDirectory(dir *dagger.Directory) Anthropic {
	m.Sandbox = m.Sandbox.WithHome(m.Sandbox.Home.WithDirectory(".", dir))
	return m
}

// Configure a remote module as context for the sandbox
func (m Anthropic) WithRemoteModule(address string) Anthropic {
	m.Sandbox = m.Sandbox.WithRemoteModule(address)
	return m
}

// Configure a local module as context for the sandbox
func (m Anthropic) WithLocalModule(module *dagger.Directory) Anthropic {
	m.Sandbox = m.Sandbox.WithLocalModule(module)
	return m
}

func (m Anthropic) History() []string {
	return m.Sandbox.History
}

func (m Anthropic) withReply(ctx context.Context, message *anthropic.Message) Anthropic {
	if len(message.Content) != 0 {
		m.Sandbox = m.Sandbox.WithNote(ctx, "", "")
	}
	hist := m.loadHistory(ctx)
	hist = append(hist, message.ToParam())
	return m.saveHistory(hist)
}

func (m Anthropic) WithToolOutput(ctx context.Context, blockId, content string) Anthropic {
	// Remove all ANSI escape codes (eg. part of raw interactive shell output), to avoid json marshalling failing
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	content = re.ReplaceAllString(content, "")
	hist := m.loadHistory(ctx)
	hist = append(hist, anthropic.NewUserMessage(anthropic.NewToolResultBlock(blockId, content, false)))
	return m.saveHistory(hist)
}

func (m Anthropic) WithPrompt(ctx context.Context, prompt string) Anthropic {
	m.Sandbox = m.Sandbox.WithNote(ctx, prompt, "🧑")
	hist := m.loadHistory(ctx)
	hist = append(hist, anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)))
	return m.saveHistory(hist)
}

func (m Anthropic) WithSystemPrompt(ctx context.Context, prompt string) Anthropic {
	m.SystemPrompt = prompt
	return m
}

func (m Anthropic) Ask(
	ctx context.Context,
	// The message to send the model
	prompt string,
) (out Anthropic, rerr error) {
	m = m.WithPrompt(ctx, prompt)
	for {
		msg, err := m.claudeQuery(ctx)
		if err != nil {
			return m, err
		}

		m = m.withReply(ctx, msg)
		content := msg.Content
		if len(content) == 0 {
			break
		}

		for _, block := range msg.Content {
			switch block.Type {
			case anthropic.ContentBlockTypeToolUse:
				switch block.Name {
				case "dagger":
					cmd := &daggerCommand{}
					if err := json.Unmarshal(block.Input, cmd); err != nil {
						return m, err
					}

					m.Sandbox, err = m.Sandbox.Run(ctx, cmd.Command)
					if err != nil {
						return m, err
					}
					run, err := m.Sandbox.LastRun()
					if err != nil {
						return m, err
					}
					result, err := run.ResultJSON()
					if err != nil {
						return m, err
					}
					m = m.WithToolOutput(ctx, block.ID, result)
				default:
					manual, err := m.Sandbox.Manual(ctx, block.Name)
					if err != nil {
						return m, err
					}
					m = m.WithToolOutput(ctx, block.ID, manual.Contents)
				}
			}
		}
	}
	return m, nil
}
