package views

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileLoader implements the pango2.TemplateLoader interface, and
// allows for loading templates from the local filesystem using multiple search paths.
type FileLoader struct {
	searchDirs       []string
	baseTemplatePath string
}

// Abs calculates the path to a given template, using the list of specified
// search paths. Each search path is check in order, and if a matching file is found,
// we use that search path to construct the final absolute path.
func (fs *FileLoader) Abs(base, name string) string {
	if filepath.IsAbs(name) {
		return name
	}

	// If the template extends another template with prefix "base:" strip the prefix and
	// load the extended template.
	if strings.HasPrefix(name, "base:") {
		return filepath.Join(fs.baseTemplatePath, strings.TrimPrefix(name, "base:"))
	}

	templatePath := ""
	for _, dir := range fs.searchDirs {
		templatePath = filepath.Join(dir, name)
		st, err := os.Stat(templatePath)
		if err == nil && st.Mode().IsRegular() {
			break
		}
	}

	return templatePath
}

// Get returns an io.Reader where the template's content can be read from.
func (fs *FileLoader) Get(p string) (io.Reader, error) {
	buf, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

// NewFileLoader creates a new loader that searches each of the specified paths for matching templates
func NewFileLoader(baseTemplatePath string, paths ...string) (*FileLoader, error) {
	var err error
	searchDirs := make([]string, 0, len(paths))

	for _, p := range paths {
		if !filepath.IsAbs(p) {
			p, err = filepath.Abs(p)
			if err != nil {
				return nil, err
			}
		}
		searchDirs = append(searchDirs, p)
	}

	return &FileLoader{searchDirs: searchDirs, baseTemplatePath: baseTemplatePath}, nil
}
