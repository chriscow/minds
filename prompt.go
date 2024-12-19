package minds

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/chriscow/minds/internal/utils"
	"gopkg.in/yaml.v2"
)

type PromptHeader struct {
	Name    string
	Version string
	Format  string
	SHA256  string
}

type Prompt struct {
	Header   PromptHeader
	Template *template.Template
}

func (p Prompt) Execute(data interface{}) (string, error) {
	var result strings.Builder
	if err := p.Template.Execute(&result, data); err != nil {
		return "", err
	}
	return result.String(), nil
}

func CreateTemplate(fs embed.FS, filepath string) (Prompt, error) {
	var prompt Prompt
	content, err := fs.ReadFile(filepath)
	if err != nil {
		return prompt, err
	}

	header, body, err := extractYAMLHeader(string(content))
	if err != nil {
		return prompt, err
	}
	sha, err := utils.SHA256Hash([]byte(body))
	if err != nil {
		return prompt, err
	}
	if header.SHA256 != "" && header.SHA256 != sha {
		return prompt, fmt.Errorf("SHA256 mismatch. Bump the version and update the SHA256")
	} else if header.SHA256 == "" {
		header.SHA256 = sha
		if err := SavePromptTemplate(filepath, header, body); err != nil {
			return prompt, err
		}
	}
	prompt.Header = header

	tmpl, err := template.New(header.Name).Parse(body)
	if err != nil {
		return prompt, err
	}
	prompt.Template = tmpl

	return prompt, nil
}

func SavePromptTemplate(filepath string, header PromptHeader, body string) error {
	file, err := os.Create(path.Join("assets", filepath))
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
	parts := strings.SplitN(templateStr, delimiter, 3)
	if len(parts) != 3 {
		return header, "", fmt.Errorf("invalid template format")
	}

	meta := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	if err := yaml.Unmarshal([]byte(meta), &header); err != nil {
		return header, "", err
	}

	return header, body, nil
}
