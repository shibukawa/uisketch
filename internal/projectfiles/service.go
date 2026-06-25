package projectfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Service struct {
	root string
}

type ProjectFile struct {
	Path string `json:"path"`
}

type FileResponse struct {
	Path     string `json:"path"`
	Source   string `json:"source"`
	Revision string `json:"revision"`
}

var ErrRevisionConflict = errors.New("file changed on disk; reload before overwriting")

func New(root string) (*Service, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	return &Service{root: resolved}, nil
}

func (s *Service) Root() string {
	return s.root
}

func (s *Service) ListFiles() ([]ProjectFile, error) {
	var files []ProjectFile
	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "dist" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if isEditableSketchPath(rel, path) {
			files = append(files, ProjectFile{Path: rel})
		}
		return nil
	})
	return files, err
}

func (s *Service) Read(path string) (FileResponse, error) {
	rel, err := s.CleanPath(path)
	if err != nil {
		return FileResponse{}, err
	}
	body, revision, err := s.readFile(rel)
	if err != nil {
		return FileResponse{}, err
	}
	return FileResponse{Path: rel, Source: string(body), Revision: revision}, nil
}

func (s *Service) Write(path, source, revision string) (FileResponse, error) {
	rel, err := s.CleanPath(path)
	if err != nil {
		return FileResponse{}, err
	}
	if err := s.writeFile(rel, []byte(source), revision); err != nil {
		return FileResponse{}, err
	}
	return s.Read(rel)
}

func (s *Service) Create(path, source string) (FileResponse, error) {
	rel, err := s.CleanPath(path)
	if err != nil {
		return FileResponse{}, err
	}
	if err := s.createFile(rel, []byte(source)); err != nil {
		return FileResponse{}, err
	}
	return s.Read(rel)
}

func (s *Service) CleanPath(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("path is required")
	}
	var rel string
	var err error
	if filepath.IsAbs(input) {
		rel, err = filepath.Rel(s.root, filepath.Clean(input))
	} else {
		rel = filepath.Clean(filepath.FromSlash(input))
	}
	if err != nil {
		return "", err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes project root")
	}
	return filepath.ToSlash(rel), nil
}

func (s *Service) readFile(rel string) ([]byte, string, error) {
	body, err := os.ReadFile(filepath.Join(s.root, filepath.FromSlash(rel)))
	if err != nil {
		return nil, "", err
	}
	return body, revisionFor(body), nil
}

func (s *Service) writeFile(rel string, body []byte, revision string) error {
	_, currentRevision, err := s.readFile(rel)
	if err != nil {
		return err
	}
	if revision == "" || revision != currentRevision {
		return ErrRevisionConflict
	}
	return os.WriteFile(filepath.Join(s.root, filepath.FromSlash(rel)), body, 0644)
}

func (s *Service) createFile(rel string, body []byte) error {
	if !strings.HasSuffix(rel, ".uisketch.md") {
		return fmt.Errorf("new files must use .uisketch.md")
	}
	full := filepath.Join(s.root, filepath.FromSlash(rel))
	if _, err := os.Stat(full); err == nil {
		return fmt.Errorf("file already exists")
	}
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return err
	}
	return os.WriteFile(full, body, 0644)
}

func isEditableSketchPath(rel, full string) bool {
	if strings.HasSuffix(rel, ".uisketch.md") {
		return true
	}
	if !strings.HasSuffix(rel, ".md") {
		return false
	}
	body, err := os.ReadFile(full)
	if err != nil {
		return false
	}
	text := string(body)
	return strings.Contains(text, "type: uisketch") || strings.Contains(text, "```uisketch")
}

func revisionFor(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
