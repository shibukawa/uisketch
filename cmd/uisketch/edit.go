package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"uisketch/internal/projectfiles"
)

type projectFileService struct {
	files *projectfiles.Service
}

type writeRequest struct {
	Path     string `json:"path"`
	Source   string `json:"source"`
	Revision string `json:"revision"`
}

func runEdit(args []string) error {
	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.SetOutput(stderr)
	project := fs.String("project", ".", "project root")
	host := fs.String("host", "127.0.0.1", "host")
	port := fs.String("port", "0", "port")
	noOpen := fs.Bool("no-open", false, "do not open browser")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("expected at most one file argument")
	}
	files, err := projectfiles.New(*project)
	if err != nil {
		return err
	}
	service := &projectFileService{files: files}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/files", service.handleFiles)
	mux.HandleFunc("/api/file", service.handleFile)
	mux.Handle("/", editorStaticHandler())

	listener, err := net.Listen("tcp", net.JoinHostPort(*host, *port))
	if err != nil {
		return err
	}
	address := "http://" + listener.Addr().String() + "/"
	if fs.NArg() == 1 {
		rel, err := service.cleanPath(fs.Arg(0))
		if err != nil {
			_ = listener.Close()
			return err
		}
		address += "?mode=local&file=" + url.QueryEscape(rel)
	} else {
		address += "?mode=local"
	}
	fmt.Fprintf(stderr, "uisketch edit project=%s url=%s\n", files.Root(), address)
	if !*noOpen {
		openBrowser(address)
	}
	return http.Serve(listener, mux)
}

func editorStaticHandler() http.Handler {
	dist := filepath.Join("packages", "web", "dist")
	if stat, err := os.Stat(dist); err == nil && stat.IsDir() {
		return http.FileServer(http.Dir(dist))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Error(w, "web editor is not built; run npm run build:web first", http.StatusServiceUnavailable)
			return
		}
		http.NotFound(w, r)
	})
}

func (s *projectFileService) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	files, err := s.files.ListFiles()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"files": files})
}

func (s *projectFileService) handleFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleRead(w, r)
	case http.MethodPut:
		s.handleWrite(w, r, false)
	case http.MethodPost:
		s.handleWrite(w, r, true)
	default:
		writeAPIError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *projectFileService) handleRead(w http.ResponseWriter, r *http.Request) {
	file, err := s.files.Read(r.URL.Query().Get("path"))
	if err != nil {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, file)
}

func (s *projectFileService) handleWrite(w http.ResponseWriter, r *http.Request, create bool) {
	var req writeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid json request")
		return
	}
	var file projectfiles.FileResponse
	var err error
	if create {
		file, err = s.files.Create(req.Path, req.Source)
	} else {
		file, err = s.files.Write(req.Path, req.Source, req.Revision)
	}
	if err != nil {
		writeAPIError(w, statusForFileError(err), err.Error())
		return
	}
	writeJSON(w, file)
}

func (s *projectFileService) cleanPath(input string) (string, error) {
	return s.files.CleanPath(input)
}

func statusForFileError(err error) int {
	if errors.Is(err, projectfiles.ErrRevisionConflict) {
		return http.StatusConflict
	}
	if errors.Is(err, os.ErrNotExist) {
		return http.StatusNotFound
	}
	if errors.Is(err, os.ErrPermission) {
		return http.StatusForbidden
	}
	return http.StatusBadRequest
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func openBrowser(address string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", address)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", address)
	default:
		cmd = exec.Command("xdg-open", address)
	}
	_ = cmd.Start()
}
