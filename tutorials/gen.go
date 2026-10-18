//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
)

const template = `
	"%s": GeneratedTutorial{
		title: "%s",
		content: %#v,

		code: []func() granio.CanvasObject{
			%s
		},
	},
`

func main() {
	w, _ := os.Create("data_gen.go")
	_, _ = io.WriteString(w,
		`package tutorials

import (
	"image"
	"image/color"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

var tutorials = map[string]GeneratedTutorial{`)

	processDirectories(w)
	io.WriteString(w, "}\n")
	w.Close()

	cmd := exec.Command("go", "fmt", "data_gen.go")
	err := cmd.Run()
	if err != nil {
		granio.LogError("failed to tidy output", err)
	}
}

func processDirectories(w io.Writer) {
	dirs := []string{"canvas", "widgets"}

	for _, dir := range dirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			granio.LogError("Failed to list directory", err)
			return
		}

		for _, file := range files {
			processFile(filepath.Join(dir, file.Name()), w)
		}
	}
}

func processFile(path string, w io.Writer) {
	title, contents := parseMarkdown(path)

	codes := ""
	for i, str := range contents {
		if i%2 == 0 {
			continue
		}

		code := str
		varName := strings.Split(code, " ")[0]
		codes += fmt.Sprintf(`	func() granio.CanvasObject {
		%s
		return %s
	},
`, code, varName)
	}

	fmt.Fprintf(w, template, path, title, contents, codes)
}

func parseMarkdown(path string) (title string, sections []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		granio.LogError("Failed to parse tutorial: "+path, err)
		return "", []string{}
	}

	sections = strings.Split(string(data), "```")

	first := strings.Split(sections[0], "\n")
	title = strings.TrimSpace(first[0][1:])

	sections[0] = sections[0][len(first[0])+1:]

	for i, s := range sections {
		sections[i] = strings.TrimSpace(s)
	}

	return title, sections
}
