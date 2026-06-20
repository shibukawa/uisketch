package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestLoaderAcceptanceCases(t *testing.T) {
	root := filepath.Clean("../testdata/loader")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read acceptance testdata: %v", err)
	}

	caseName := regexp.MustCompile(`^\d{3}_[a-z0-9]+(?:_[a-z0-9]+)*$`)
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !caseName.MatchString(name) {
			t.Fatalf("loader acceptance case %q does not match NNN_name", name)
		}
		dirs = append(dirs, name)
	}
	sort.Strings(dirs)
	if len(dirs) != 5 {
		t.Fatalf("loader acceptance case count = %d, want 5", len(dirs))
	}

	for i, dir := range dirs {
		wantPrefix := fmt.Sprintf("%03d", i+1)
		if !strings.HasPrefix(dir, wantPrefix+"_") {
			t.Fatalf("loader acceptance case %q is out of sequence, want prefix %s_", dir, wantPrefix)
		}
		t.Run(dir, func(t *testing.T) {
			caseDir := filepath.Join(root, dir)
			input := readFixture(t, filepath.Join(caseDir, "input.md"))
			wantOutput := readFixture(t, filepath.Join(caseDir, "output.md"))
			wantAssets := readOptionalFixture(t, filepath.Join(caseDir, "assets.txt"))

			got, err := Build(input, Options{AssetDir: "assets"})
			if err != nil {
				t.Fatalf("Build returned error: %v", err)
			}
			if got.Markdown+"\n" != wantOutput {
				t.Fatalf("markdown output mismatch\n--- got ---\n%s\n--- want ---\n%s", got.Markdown+"\n", wantOutput)
			}

			var assetLines []string
			for _, assetPath := range AssetPaths(got) {
				body := got.Assets[assetPath]
				if strings.TrimSpace(body) == "" {
					t.Fatalf("asset %q is empty", assetPath)
				}
				assetLines = append(assetLines, assetPath)
				wantAsset := readFixture(t, filepath.Join(caseDir, filepath.FromSlash(assetPath)))
				if body != wantAsset {
					t.Fatalf("asset %s mismatch\n--- got ---\n%s\n--- want ---\n%s", assetPath, body, wantAsset)
				}
			}
			gotAssets := strings.Join(assetLines, "\n")
			if gotAssets != "" {
				gotAssets += "\n"
			}
			if gotAssets != wantAssets {
				t.Fatalf("asset list mismatch\n--- got ---\n%s\n--- want ---\n%s", gotAssets, wantAssets)
			}
		})
	}
}

func readFixture(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}

func readOptionalFixture(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}
