package concept

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Screen struct {
	ID      string
	Type    string
	Title   string
	Source  Source
	Actions []string
	Path    string
}

type Source struct {
	Type     string
	Location string
}

func LoadScreen(path string) (*Screen, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fm, err := frontmatter(string(body))
	if err != nil {
		return nil, err
	}

	screen := &Screen{Path: path}
	var section string
	for _, raw := range strings.Split(fm, "\n") {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)
		if indent == 0 {
			section = ""
		}
		if strings.HasPrefix(trimmed, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			if section == "actions" && value != "" {
				screen.Actions = append(screen.Actions, value)
			}
			continue
		}
		key, value, ok := splitKeyValue(trimmed)
		if !ok {
			continue
		}
		if indent == 0 && value == "" {
			section = key
			continue
		}
		switch {
		case indent == 0 && key == "id":
			screen.ID = value
		case indent == 0 && key == "type":
			screen.Type = value
		case indent == 0 && key == "title":
			screen.Title = value
		case section == "source" && key == "type":
			screen.Source.Type = value
		case section == "source" && key == "location":
			screen.Source.Location = value
		}
	}
	if screen.Type != "screen" {
		return nil, fmt.Errorf("screen concept %s has type %q, want screen", path, screen.Type)
	}
	if screen.ID == "" {
		return nil, fmt.Errorf("screen concept %s is missing id", path)
	}
	return screen, nil
}

func (s *Screen) SourcePath() string {
	if filepath.IsAbs(s.Source.Location) {
		return s.Source.Location
	}
	return filepath.Join(filepath.Dir(s.Path), s.Source.Location)
}

func frontmatter(content string) (string, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return "", fmt.Errorf("missing frontmatter")
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", fmt.Errorf("unterminated frontmatter")
	}
	return rest[:end], nil
}

func splitKeyValue(line string) (string, string, bool) {
	before, after, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(before), unquote(strings.TrimSpace(after)), true
}

func countIndent(s string) int {
	n := 0
	for _, r := range s {
		if r != ' ' {
			break
		}
		n++
	}
	return n
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
