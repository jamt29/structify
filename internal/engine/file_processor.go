package engine

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jamt29/structify/internal/dsl"
	"github.com/jamt29/structify/internal/template"
)

// ProcessFiles copies/renders files from a template into the request OutputDir.
func ProcessFiles(req *template.ScaffoldRequest) (created []string, skipped []string, err error) {
	if req == nil {
		return nil, nil, fmt.Errorf("request is nil")
	}
	if req.Template == nil {
		return nil, nil, fmt.Errorf("request template is nil")
	}
	if req.Template.Manifest == nil {
		return nil, nil, fmt.Errorf("template manifest is nil")
	}
	if strings.TrimSpace(req.OutputDir) == "" {
		return nil, nil, fmt.Errorf("output dir is empty")
	}
	if req.Variables == nil {
		return nil, nil, fmt.Errorf("variables context is nil")
	}

	srcRoot := filepath.Join(req.Template.Path, "template")
	if _, statErr := os.Stat(srcRoot); statErr != nil {
		if os.IsNotExist(statErr) {
			// Nothing to do: templates may ship empty template/ directories.
			return []string{}, []string{}, nil
		}
		return nil, nil, fmt.Errorf("stat template dir %s: %w", srcRoot, statErr)
	}

	manifest := req.Template.Manifest

	walkErr := filepath.WalkDir(srcRoot, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if p == srcRoot {
			return nil
		}

		rel, err := filepath.Rel(srcRoot, p)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)

		action, whenOk, ruleMatched, err := decideFileAction(relSlash, manifest.Files, req.Variables)
		if err != nil {
			return fmt.Errorf("evaluating file rules for %s: %w", relSlash, err)
		}

		shouldSkip := false
		if ruleMatched {
			switch action {
			case "exclude":
				shouldSkip = whenOk
			case "include":
				shouldSkip = !whenOk
			}
		}
		if shouldSkip {
			skipped = append(skipped, relSlash)
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		interpRel, err := interpolateRelPath(relSlash, req.Variables)
		if err != nil {
			return fmt.Errorf("interpolating path %s: %w", relSlash, err)
		}

		destRel := interpRel
		if strings.HasSuffix(p, ".tmpl") {
			destRel = strings.TrimSuffix(destRel, ".tmpl")
		}
		destAbs := filepath.Join(req.OutputDir, filepath.FromSlash(destRel))

		if d.IsDir() {
			if req.DryRun {
				return nil
			}
			return os.MkdirAll(destAbs, 0o755)
		}

		created = append(created, destRel)
		if req.DryRun {
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
			return err
		}

		if strings.HasSuffix(p, ".tmpl") {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out, err := dsl.InterpolateFile(b, req.Variables)
			if err != nil {
				return fmt.Errorf("interpolating file %s: %w", relSlash, err)
			}
			return os.WriteFile(destAbs, out, 0o644)
		}

		return copyFileBytes(p, destAbs)
	})
	if walkErr != nil {
		return created, skipped, walkErr
	}

	return created, skipped, nil
}

func decideFileAction(relSlash string, rules []dsl.FileRule, ctx dsl.Context) (action string, whenOk bool, matched bool, err error) {
	// last rule wins
	for i := len(rules) - 1; i >= 0; i-- {
		r := rules[i]
		pattern := strings.TrimSpace(r.Include)
		a := "include"
		if pattern == "" {
			pattern = strings.TrimSpace(r.Exclude)
			a = "exclude"
		}
		if strings.TrimSpace(pattern) == "" {
			continue
		}
		ok, err := matchGlob(pattern, relSlash)
		if err != nil {
			return "", false, false, err
		}
		if !ok {
			continue
		}

		cond := strings.TrimSpace(r.When)
		if cond == "" {
			return a, true, true, nil
		}
		ast, err := dsl.NewParser(cond).Parse()
		if err != nil {
			return "", false, false, err
		}
		v, err := dsl.Evaluate(ast, ctx)
		if err != nil {
			return "", false, false, err
		}
		return a, v, true, nil
	}
	return "", true, false, nil
}

func interpolateRelPath(relSlash string, ctx dsl.Context) (string, error) {
	parts := strings.Split(relSlash, "/")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		s, err := dsl.Interpolate(parts[i], ctx)
		if err != nil {
			return "", err
		}
		parts[i] = s
	}
	return path.Join(parts...), nil
}

// matchGlob matches relSlash (always forward slashes) against a glob pattern
// supporting ** (match zero or more path segments), * and exact segments.
func matchGlob(pattern string, relSlash string) (bool, error) {
	pat := filepath.ToSlash(strings.TrimSpace(pattern))
	val := filepath.ToSlash(strings.TrimPrefix(relSlash, "./"))

	psegs := splitPath(pat)
	vsegs := splitPath(val)

	type key struct{ i, j int }
	memo := map[key]bool{}
	seen := map[key]bool{}

	var rec func(i, j int) (bool, error)
	rec = func(i, j int) (bool, error) {
		k := key{i: i, j: j}
		if seen[k] {
			return memo[k], nil
		}
		seen[k] = true

		// end of pattern
		if i == len(psegs) {
			memo[k] = (j == len(vsegs))
			return memo[k], nil
		}

		seg := psegs[i]
		if seg == "**" {
			// match zero segments
			if ok, err := rec(i+1, j); err != nil {
				return false, err
			} else if ok {
				memo[k] = true
				return true, nil
			}
			// match one or more segments
			if j < len(vsegs) {
				ok, err := rec(i, j+1)
				if err != nil {
					return false, err
				}
				memo[k] = ok
				return ok, nil
			}
			memo[k] = false
			return false, nil
		}

		if j >= len(vsegs) {
			memo[k] = false
			return false, nil
		}

		ok, err := path.Match(seg, vsegs[j])
		if err != nil {
			return false, err
		}
		if !ok {
			memo[k] = false
			return false, nil
		}
		v, err := rec(i+1, j+1)
		if err != nil {
			return false, err
		}
		memo[k] = v
		return v, nil
	}

	return rec(0, 0)
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return []string{}
	}
	return strings.Split(p, "/")
}

func copyFileBytes(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

