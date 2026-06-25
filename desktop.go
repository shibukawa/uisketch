package main

import (
	"context"
	"embed"
	"os"
	"sync"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"uisketch/internal/projectfiles"
	"uisketch/internal/renderapi"
)

//go:embed all:desktop_assets
var desktopAssets embed.FS

type DesktopApp struct {
	ctx   context.Context
	files *projectfiles.Service
	mu    sync.RWMutex
	dirty bool
}

func main() {
	app := &DesktopApp{}
	err := wails.Run(&options.App{
		Title:  "UI Sketch",
		Width:  1280,
		Height: 860,
		AssetServer: &assetserver.Options{
			Assets: desktopAssets,
		},
		OnStartup: app.startup,
		OnBeforeClose: func(ctx context.Context) bool {
			if !app.IsDirty() {
				return false
			}
			choice, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
				Type:          runtime.QuestionDialog,
				Title:         "Discard unsaved changes?",
				Message:       "Unsaved changes will be discarded if this window closes.",
				Buttons:       []string{"Discard", "Cancel"},
				DefaultButton: "Cancel",
				CancelButton:  "Cancel",
			})
			return err == nil && choice != "Discard"
		},
		Bind: []interface{}{app},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *DesktopApp) startup(ctx context.Context) {
	a.ctx = ctx
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}
	a.files, _ = projectfiles.New(root)
}

func (a *DesktopApp) ListFiles() ([]projectfiles.ProjectFile, error) {
	return a.files.ListFiles()
}

func (a *DesktopApp) ReadFile(path string) (projectfiles.FileResponse, error) {
	return a.files.Read(path)
}

func (a *DesktopApp) WriteFile(path, source, revision string) (projectfiles.FileResponse, error) {
	return a.files.Write(path, source, revision)
}

func (a *DesktopApp) CreateFile(path, source string) (projectfiles.FileResponse, error) {
	return a.files.Create(path, source)
}

func (a *DesktopApp) RenderSource(source string) renderapi.Response {
	return renderapi.RenderYAML(source)
}

func (a *DesktopApp) SetDirty(dirty bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dirty = dirty
}

func (a *DesktopApp) IsDirty() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.dirty
}
