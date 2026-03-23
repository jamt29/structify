package template

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type AnalysisResult struct {
	ProjectName    string
	Language       string
	Architecture   string
	DetectedVars   []DetectedVar
	FilesToInclude []string
	FilesToIgnore  []string
	TotalFiles     int
}

type DetectedVar struct {
	ID          string
	Description string
	Type        string
	Occurrences []Occurrence
	SuggestAs   string
}

type Occurrence struct {
	File    string
	Line    int
	Context string
}

func AnalyzeProject(path string) (*AnalysisResult, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs project path: %w", err)
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("stat project path: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("project path is not a directory: %s", abs)
	}

	projectName := filepath.Base(abs)
	lang, err := detectLanguage(abs)
	if err != nil {
		return nil, err
	}

	ignoredSet := defaultIgnoreSet()
	include := []string{}
	ignore := []string{}
	totalFiles := 0

	projectVar := DetectedVar{
		ID:          "project_name",
		Description: "Nombre del proyecto",
		Type:        "string",
		SuggestAs:   projectName,
	}
	moduleVar := DetectedVar{
		ID:          "module_path",
		Description: "Go module path",
		Type:        "string",
	}

	_ = filepath.WalkDir(abs, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == abs {
			return nil
		}

		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			if shouldIgnoreDir(rel) {
				ignore = append(ignore, rel+"/")
				return filepath.SkipDir
			}
			return nil
		}

		totalFiles++
		if shouldIgnoreFile(rel) {
			ignore = append(ignore, rel)
			ignoredSet[rel] = struct{}{}
			return nil
		}

		include = append(include, rel)
		occ, modulePath, pName := scanFileForVars(path, rel, projectName)
		projectVar.Occurrences = append(projectVar.Occurrences, occ...)
		if modulePath != "" && moduleVar.SuggestAs == "" {
			moduleVar.SuggestAs = modulePath
		}
		if pName != "" && projectVar.SuggestAs == "" {
			projectVar.SuggestAs = pName
		}
		return nil
	})

	vars := []DetectedVar{}
	if projectVar.SuggestAs != "" && len(projectVar.Occurrences) > 0 {
		vars = append(vars, projectVar)
	}
	if moduleVar.SuggestAs != "" {
		vars = append(vars, moduleVar)
	}

	return &AnalysisResult{
		ProjectName:    projectName,
		Language:       lang,
		Architecture:   "unknown",
		DetectedVars:   vars,
		FilesToInclude: include,
		FilesToIgnore:  uniqueSorted(append(ignore, defaultIgnoreDisplay()...)),
		TotalFiles:     totalFiles,
	}, nil
}

func detectLanguage(root string) (string, error) {
	if exists(filepath.Join(root, "go.mod")) {
		return "go", nil
	}
	if exists(filepath.Join(root, "package.json")) {
		return "typescript", nil
	}
	if exists(filepath.Join(root, "Cargo.toml")) {
		return "rust", nil
	}
	if exists(filepath.Join(root, "requirements.txt")) || exists(filepath.Join(root, "pyproject.toml")) {
		return "python", nil
	}
	csproj, err := filepath.Glob(filepath.Join(root, "*.csproj"))
	if err == nil && len(csproj) > 0 {
		return "csharp", nil
	}
	return "unknown", nil
}

func scanFileForVars(absFile, relFile, projectName string) ([]Occurrence, string, string) {
	if !isTextFile(relFile) {
		return nil, "", ""
	}

	f, err := os.Open(absFile)
	if err != nil {
		return nil, "", ""
	}
	defer f.Close()

	needleVariants := projectNameVariants(projectName)
	moduleRe := regexp.MustCompile(`^\s*module\s+([^\s]+)\s*$`)
	pkgJSONRe := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	cargoNameRe := regexp.MustCompile(`^\s*name\s*=\s*"([^"]+)"\s*$`)

	sc := bufio.NewScanner(f)
	lineNo := 0
	occ := []Occurrence{}
	modulePath := ""
	projectDetected := ""

	for sc.Scan() {
		lineNo++
		line := sc.Text()
		trim := strings.TrimSpace(line)

		if m := moduleRe.FindStringSubmatch(trim); len(m) == 2 {
			modulePath = m[1]
			for _, v := range needleVariants {
				if strings.Contains(m[1], v) {
					occ = append(occ, Occurrence{File: relFile, Line: lineNo, Context: line})
					break
				}
			}
		}
		if m := pkgJSONRe.FindStringSubmatch(line); len(m) == 2 {
			projectDetected = m[1]
			if containsAny(projectDetected, needleVariants) {
				occ = append(occ, Occurrence{File: relFile, Line: lineNo, Context: line})
			}
		}
		if m := cargoNameRe.FindStringSubmatch(trim); len(m) == 2 {
			projectDetected = m[1]
			if containsAny(projectDetected, needleVariants) {
				occ = append(occ, Occurrence{File: relFile, Line: lineNo, Context: line})
			}
		}

		if looksLikeImportOrModuleContext(line) && containsAny(line, needleVariants) {
			occ = append(occ, Occurrence{File: relFile, Line: lineNo, Context: line})
		}
	}

	return occ, modulePath, projectDetected
}

func looksLikeImportOrModuleContext(line string) bool {
	l := strings.ToLower(line)
	return strings.Contains(l, "module ") ||
		strings.Contains(l, "import ") ||
		strings.Contains(l, `"name"`) ||
		strings.Contains(l, "from ") ||
		strings.Contains(l, "package ")
}

func projectNameVariants(name string) []string {
	base := strings.TrimSpace(name)
	if base == "" {
		return nil
	}
	parts := splitWords(base)
	if len(parts) == 0 {
		return []string{base}
	}
	snake := strings.Join(parts, "_")
	kebab := strings.Join(parts, "-")
	camel := parts[0]
	for i := 1; i < len(parts); i++ {
		camel += strings.Title(parts[i])
	}
	pascal := ""
	for _, p := range parts {
		pascal += strings.Title(p)
	}
	return uniqueSorted([]string{base, snake, kebab, camel, pascal})
}

func splitWords(s string) []string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	raw := strings.Fields(s)
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func containsAny(s string, values []string) bool {
	for _, v := range values {
		if strings.Contains(s, v) {
			return true
		}
	}
	return false
}

func defaultIgnoreSet() map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range defaultIgnoreDisplay() {
		out[v] = struct{}{}
	}
	return out
}

func defaultIgnoreDisplay() []string {
	return []string{
		".git/",
		"node_modules/",
		"vendor/",
		"target/",
		"dist/",
		"build/",
		".next/",
		"__pycache__/",
		".env",
		".env.local",
		".env.*",
		"*.exe",
		"*.dll",
		"*.so",
		"*.jpg",
		"*.png",
		"*.gif",
		"*.ico",
	}
}

func shouldIgnoreDir(rel string) bool {
	for _, d := range []string{".git", "node_modules", "vendor", "target", "dist", "build", ".next", "__pycache__"} {
		if rel == d || strings.HasPrefix(rel, d+"/") {
			return true
		}
	}
	return false
}

func shouldIgnoreFile(rel string) bool {
	name := strings.ToLower(filepath.Base(rel))
	if name == ".env" || name == ".env.local" || strings.HasPrefix(name, ".env.") {
		return true
	}
	for _, suf := range []string{".exe", ".dll", ".so", ".jpg", ".png", ".gif", ".ico"} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isTextFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	for _, suf := range []string{".exe", ".dll", ".so", ".jpg", ".png", ".gif", ".ico", ".pdf", ".zip", ".tar", ".gz"} {
		if strings.HasSuffix(name, suf) {
			return false
		}
	}
	return true
}

func uniqueSorted(in []string) []string {
	m := map[string]struct{}{}
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		m[v] = struct{}{}
	}
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sortStrings(out)
	return out
}

func sortStrings(v []string) {
	for i := 0; i < len(v); i++ {
		for j := i + 1; j < len(v); j++ {
			if v[j] < v[i] {
				v[i], v[j] = v[j], v[i]
			}
		}
	}
}
