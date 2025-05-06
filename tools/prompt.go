package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type PromptHeader struct {
	Name    string         `yaml:"name,omitempty"`
	Version string         `yaml:"version,omitempty"`
	Format  string         `yaml:"format,omitempty"`
	SHA256  string         `yaml:"sha256,omitempty"`
	Extra   map[string]any `yaml:",inline"`
}

type Prompt struct {
	Header     PromptHeader
	Template   *template.Template
	RawContent string
}

func (p Prompt) Execute(data any) (string, error) {
	var result strings.Builder
	if err := p.Template.Execute(&result, data); err != nil {
		return "", err
	}
	return result.String(), nil
}

func GeneratePrompt(content []byte, data any) (string, error) {
	tmpl, err := CreatePromptTemplate(content)
	if err != nil {
		return "", fmt.Errorf("failed to create template: %w", err)
	}

	return tmpl.Execute(data)
}

func validateTemplateContent(content string) error {
	// Skip validation for code blocks that might contain valid single braces
	// Extract content outside of code blocks (indicated by ```...```)
	contentToCheck := content
	codeBlocks := strings.Split(content, "```")

	// If there are code blocks, we only want to check the non-code parts
	if len(codeBlocks) > 1 {
		// Rebuild the content without code blocks
		contentToCheck = ""
		for i, block := range codeBlocks {
			// Even indices are non-code blocks
			if i%2 == 0 {
				contentToCheck += block
			}
		}
	}

	// Check for unmatched single curly braces in non-code content
	doubleBraceCount := strings.Count(contentToCheck, "{{")
	doubleCloseBraceCount := strings.Count(contentToCheck, "}}")

	// Count total occurrences of { and }
	totalLeftBrace := strings.Count(contentToCheck, "{")
	totalRightBrace := strings.Count(contentToCheck, "}")

	// If there are more single braces than double braces, there must be a single brace
	if totalLeftBrace > doubleBraceCount*2 {
		return fmt.Errorf("template contains single left curly brace '{', use '{{' instead")
	}

	if totalRightBrace > doubleCloseBraceCount*2 {
		return fmt.Errorf("template contains single right curly brace '}', use '}}' instead")
	}

	return nil
}

func CreatePromptTemplate(content []byte) (Prompt, error) {
	var prompt Prompt

	header, body, err := extractYAMLHeader(string(content))
	if err != nil {
		return prompt, err
	}

	// Validate template content for single curly braces
	if err := validateTemplateContent(body); err != nil {
		return prompt, err
	}

	sha, err := sHA256Hash([]byte(body))
	if err != nil {
		return prompt, err
	}
	if header.SHA256 != "" && header.SHA256 != sha {
		return prompt, fmt.Errorf("SHA256 mismatch. Bump the version and update the SHA256")
	} else if header.SHA256 == "" {
		header.SHA256 = sha
		// if err := SavePromptTemplate(filepath, header, body); err != nil {
		// 	return prompt, fmt.Errorf("failed to save prompt template while updating hash: %w", err)
		// }
	}
	prompt.Header = header
	prompt.RawContent = body

	tmpl, err := template.New(header.Name).Parse(body)
	if err != nil {
		return prompt, err
	}
	prompt.Template = tmpl

	return prompt, nil
}

func SavePromptTemplate(filePath string, header PromptHeader, body string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString("---\n"); err != nil {
		return err
	}
	meta, err := yaml.Marshal(header)
	if err != nil {
		return err
	}
	if _, err := file.Write(meta); err != nil {
		return err
	}
	if _, err := file.WriteString("\n---\n"); err != nil {
		return err
	}
	if _, err := file.WriteString(body); err != nil {
		return err
	}
	return nil
}

func extractYAMLHeader(templateStr string) (PromptHeader, string, error) {
	const delimiter = "---"
	var header PromptHeader
	parts := strings.Split(templateStr, delimiter)
	if len(parts) < 3 {
		return header, templateStr, nil
	}

	meta := strings.TrimSpace(parts[1])
	if strings.HasPrefix(meta, "-") {
		// the user put too many dashes or it was not the right format
		return header, templateStr, fmt.Errorf("invalid header format")
	}
	body := strings.Join(parts[2:], delimiter)

	if err := yaml.Unmarshal([]byte(meta), &header); err != nil {
		return header, templateStr, err
	}

	return header, strings.TrimSpace(body), nil
}

func sHA256Hash(data []byte) (string, error) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("write data: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
