package template

import (
	"encoding/json"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
)

type AnalysisResult struct {
	ProjectName    string
	Language       string
	Architecture   string
	DetectedVars   []DetectedVar
	FilesToInclude []string
	FilesToIgnore  []string
	TotalFiles     int
	DetectedDeps   []DetectedDependency
	SuggestedInputs []SuggestedInput
	Confidence     float64
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

type DetectedDependency struct {
	Name    string
	Alias   string
	Purpose string
}

type SuggestedInput struct {
	Input   dsl.Input
	Reason  string
	Default string
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

	deps, suggested, extraVars := detectByLanguage(abs, lang)
	vars = append(vars, extraVars...)
	confidence := calculateConfidence(lang, vars, deps, totalFiles)

	return &AnalysisResult{
		ProjectName:    projectName,
		Language:       lang,
		Architecture:   "unknown",
		DetectedVars:   vars,
		FilesToInclude: include,
		FilesToIgnore:  uniqueSorted(append(ignore, defaultIgnoreDisplay()...)),
		TotalFiles:     totalFiles,
		DetectedDeps:   deps,
		SuggestedInputs: suggested,
		Confidence:     confidence,
	}, nil
}

func detectByLanguage(root, lang string) ([]DetectedDependency, []SuggestedInput, []DetectedVar) {
	switch lang {
	case "go":
		return detectGoSignals(root)
	case "typescript":
		return detectNodeSignals(root)
	case "rust":
		return detectRustSignals(root)
	default:
		return nil, nil, nil
	}
}

func detectGoSignals(root string) ([]DetectedDependency, []SuggestedInput, []DetectedVar) {
	goModPath := filepath.Join(root, "go.mod")
	b, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, nil, nil
	}
	content := string(b)
	deps := []DetectedDependency{}
	suggested := []SuggestedInput{}
	vars := []DetectedVar{}

	goVersionRe := regexp.MustCompile(`(?m)^\s*go\s+([0-9]+\.[0-9]+)\s*$`)
	if m := goVersionRe.FindStringSubmatch(content); len(m) == 2 {
		vars = append(vars, DetectedVar{
			ID:          "go_version",
			Description: "Go version",
			Type:        "string",
			SuggestAs:   m[1],
		})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{
				ID:      "go_version",
				Prompt:  "Go version?",
				Type:    "string",
				Default: m[1],
			},
			Reason:  "Detectado en go.mod",
			Default: m[1],
		})
	}
	if strings.Contains(content, "gorm.io/gorm") {
		deps = append(deps, DetectedDependency{Name: "gorm.io/gorm", Alias: "gorm", Purpose: "ORM"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{
				ID:      "orm",
				Prompt:  "ORM?",
				Type:    "enum",
				Options: []string{"gorm", "sqlx", "none"},
				Default: "gorm",
			},
			Reason:  "Detectado: gorm.io/gorm",
			Default: "gorm",
		})
	}
	if strings.Contains(content, "github.com/gin-gonic/gin") {
		deps = append(deps, DetectedDependency{Name: "github.com/gin-gonic/gin", Alias: "gin", Purpose: "HTTP framework"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{
				ID:      "transport",
				Prompt:  "Framework HTTP?",
				Type:    "enum",
				Options: []string{"gin", "echo", "fiber", "none"},
				Default: "gin",
			},
			Reason:  "Detectado: gin-gonic/gin",
			Default: "gin",
		})
	}
	if strings.Contains(content, "github.com/labstack/echo") {
		deps = append(deps, DetectedDependency{Name: "github.com/labstack/echo", Alias: "echo", Purpose: "HTTP framework"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{
				ID:      "transport",
				Prompt:  "Framework HTTP?",
				Type:    "enum",
				Options: []string{"gin", "echo", "fiber", "none"},
				Default: "echo",
			},
			Reason:  "Detectado: labstack/echo",
			Default: "echo",
		})
	}

	mainCount := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr == nil && strings.Contains(string(data), "package main") {
			mainCount++
		}
		return nil
	})
	if mainCount > 0 {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{
				ID:      "is_executable",
				Prompt:  "Es ejecutable (package main)?",
				Type:    "bool",
				Default: true,
			},
			Reason:  "Se detecto package main",
			Default: "true",
		})
	}
	cmdEntries, _ := os.ReadDir(filepath.Join(root, "cmd"))
	binCount := 0
	for _, e := range cmdEntries {
		if e.IsDir() {
			binCount++
		}
	}
	if binCount > 0 {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{
				ID:      "binary_count",
				Prompt:  "Cantidad de binarios en /cmd?",
				Type:    "string",
				Default: fmt.Sprintf("%d", binCount),
			},
			Reason:  "Detectado por carpetas en /cmd",
			Default: fmt.Sprintf("%d", binCount),
		})
	}
	return uniqueDeps(deps), dedupeSuggestedInputs(suggested), vars
}

func detectNodeSignals(root string) ([]DetectedDependency, []SuggestedInput, []DetectedVar) {
	type pkgJSON struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
	}
	var pkg pkgJSON
	b, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil || json.Unmarshal(b, &pkg) != nil {
		return nil, nil, nil
	}
	deps := []DetectedDependency{}
	suggested := []SuggestedInput{}
	vars := []DetectedVar{}

	if strings.TrimSpace(pkg.Name) != "" {
		vars = append(vars, DetectedVar{ID: "project_name", Description: "Nombre del proyecto", Type: "string", SuggestAs: pkg.Name})
	}
	if strings.TrimSpace(pkg.Version) != "" {
		vars = append(vars, DetectedVar{ID: "version", Description: "Version", Type: "string", SuggestAs: pkg.Version})
	}
	if _, ok := pkg.Dependencies["express"]; ok {
		deps = append(deps, DetectedDependency{Name: "express", Alias: "express", Purpose: "Node runtime"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "runtime", Prompt: "Runtime Node?", Type: "enum", Options: []string{"express", "fastify", "none"}, Default: "express"},
			Reason: "Detectado en dependencies.express", Default: "express",
		})
	}
	if _, ok := pkg.Dependencies["fastify"]; ok {
		deps = append(deps, DetectedDependency{Name: "fastify", Alias: "fastify", Purpose: "Node runtime"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "runtime", Prompt: "Runtime Node?", Type: "enum", Options: []string{"express", "fastify", "none"}, Default: "fastify"},
			Reason: "Detectado en dependencies.fastify", Default: "fastify",
		})
	}
	if _, ok := pkg.DevDependencies["typescript"]; ok {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "typescript", Prompt: "Usa TypeScript?", Type: "bool", Default: true},
			Reason: "Detectado en devDependencies.typescript", Default: "true",
		})
	}
	if _, ok := pkg.DevDependencies["jest"]; ok {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "include_tests", Prompt: "Incluir tests?", Type: "bool", Default: true},
			Reason: "Detectado en devDependencies.jest", Default: "true",
		})
	}
	if _, ok := pkg.DevDependencies["vitest"]; ok {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "include_tests", Prompt: "Incluir tests?", Type: "bool", Default: true},
			Reason: "Detectado en devDependencies.vitest", Default: "true",
		})
	}
	if _, ok := pkg.Scripts["dev"]; ok {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "has_dev_script", Prompt: "Tiene script dev?", Type: "bool", Default: true},
			Reason: "Detectado en scripts.dev", Default: "true",
		})
	}

	type tsConfig struct {
		CompilerOptions struct {
			Strict any    `json:"strict"`
			OutDir string `json:"outDir"`
		} `json:"compilerOptions"`
	}
	var ts tsConfig
	if tsB, tsErr := os.ReadFile(filepath.Join(root, "tsconfig.json")); tsErr == nil && json.Unmarshal(tsB, &ts) == nil {
		if v, ok := ts.CompilerOptions.Strict.(bool); ok && v {
			suggested = append(suggested, SuggestedInput{
				Input: dsl.Input{ID: "strict_mode", Prompt: "TypeScript strict mode?", Type: "bool", Default: true},
				Reason: "Detectado en tsconfig.compilerOptions.strict", Default: "true",
			})
		}
		if strings.TrimSpace(ts.CompilerOptions.OutDir) != "" {
			suggested = append(suggested, SuggestedInput{
				Input: dsl.Input{ID: "out_dir", Prompt: "Output directory?", Type: "string", Default: ts.CompilerOptions.OutDir},
				Reason: "Detectado en tsconfig.compilerOptions.outDir", Default: ts.CompilerOptions.OutDir,
			})
		}
	}
	return uniqueDeps(deps), dedupeSuggestedInputs(suggested), vars
}

func detectRustSignals(root string) ([]DetectedDependency, []SuggestedInput, []DetectedVar) {
	b, err := os.ReadFile(filepath.Join(root, "Cargo.toml"))
	if err != nil {
		return nil, nil, nil
	}
	content := string(b)
	deps := []DetectedDependency{}
	suggested := []SuggestedInput{}
	vars := []DetectedVar{}

	nameRe := regexp.MustCompile(`(?m)^\s*name\s*=\s*"([^"]+)"\s*$`)
	versionRe := regexp.MustCompile(`(?m)^\s*version\s*=\s*"([^"]+)"\s*$`)
	if m := nameRe.FindStringSubmatch(content); len(m) == 2 {
		vars = append(vars, DetectedVar{ID: "project_name", Description: "Nombre del proyecto", Type: "string", SuggestAs: m[1]})
	}
	if m := versionRe.FindStringSubmatch(content); len(m) == 2 {
		vars = append(vars, DetectedVar{ID: "version", Description: "Version", Type: "string", SuggestAs: m[1]})
	}
	if strings.Contains(content, "\n[workspace]") || strings.HasPrefix(strings.TrimSpace(content), "[workspace]") {
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "workspace", Prompt: "Es workspace multi-crate?", Type: "bool", Default: true},
			Reason: "Detectado bloque [workspace]", Default: "true",
		})
	}
	if strings.Contains(content, "axum") {
		deps = append(deps, DetectedDependency{Name: "axum", Alias: "axum", Purpose: "HTTP transport"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "transport", Prompt: "Transport layer?", Type: "enum", Options: []string{"axum", "actix", "none"}, Default: "axum"},
			Reason: "Detectado en Cargo.toml: axum", Default: "axum",
		})
	}
	if strings.Contains(content, "actix-web") {
		deps = append(deps, DetectedDependency{Name: "actix-web", Alias: "actix", Purpose: "HTTP transport"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "transport", Prompt: "Transport layer?", Type: "enum", Options: []string{"axum", "actix", "none"}, Default: "actix"},
			Reason: "Detectado en Cargo.toml: actix-web", Default: "actix",
		})
	}
	if strings.Contains(content, "serde") {
		deps = append(deps, DetectedDependency{Name: "serde", Alias: "serde", Purpose: "Serialization"})
		suggested = append(suggested, SuggestedInput{
			Input: dsl.Input{ID: "use_serde", Prompt: "Usar serde?", Type: "bool", Default: true},
			Reason: "Detectado en Cargo.toml: serde", Default: "true",
		})
	}
	return uniqueDeps(deps), dedupeSuggestedInputs(suggested), vars
}

func calculateConfidence(lang string, vars []DetectedVar, deps []DetectedDependency, totalFiles int) float64 {
	score := 0.35
	if lang != "unknown" {
		score += 0.25
	}
	if len(vars) > 0 {
		score += 0.2
	}
	if len(deps) > 0 {
		score += 0.15
	}
	if totalFiles > 0 {
		score += 0.05
	}
	if score > 1 {
		score = 1
	}
	return score
}

func uniqueDeps(in []DetectedDependency) []DetectedDependency {
	seen := map[string]struct{}{}
	out := make([]DetectedDependency, 0, len(in))
	for _, d := range in {
		k := d.Name + "|" + d.Alias + "|" + d.Purpose
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, d)
	}
	return out
}

func dedupeSuggestedInputs(in []SuggestedInput) []SuggestedInput {
	seen := map[string]int{}
	out := make([]SuggestedInput, 0, len(in))
	for _, s := range in {
		id := strings.TrimSpace(s.Input.ID)
		if id == "" {
			continue
		}
		if idx, ok := seen[id]; ok {
			// Prefer the latest signal for defaults like runtime/transport.
			out[idx] = s
			continue
		}
		seen[id] = len(out)
		out = append(out, s)
	}
	return out
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
