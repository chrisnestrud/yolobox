package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

func validateProjectFilteringConfig(cfg Config, projectDir string) error {
	if (len(cfg.Exclude) > 0 || len(cfg.CopyAs) > 0) && !cfg.ReadonlyProject {
		return fmt.Errorf("--exclude and --copy-as currently require --readonly-project")
	}
	if _, err := resolveExcludedProjectPaths(cfg.Exclude, projectDir); err != nil {
		return err
	}
	for _, spec := range cfg.CopyAs {
		if _, err := parseCopyAsSpec(spec, projectDir); err != nil {
			return err
		}
	}
	return nil
}

type projectPathInfo struct {
	abs   string
	rel   string
	isDir bool
}

type projectCopyAsSpec struct {
	src string
	dst string
	rel string
}

func resolveHostPath(path string, projectDir string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else if strings.HasPrefix(path, "~/") {
			path = filepath.Join(home, path[2:])
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectDir, path)
	}
	return filepath.Clean(path), nil
}

func resolveProjectRelativeTarget(target string, projectDir string) (string, error) {
	target = filepath.ToSlash(strings.TrimSpace(target))
	if target == "" {
		return "", fmt.Errorf("empty project path")
	}
	if strings.HasPrefix(target, "~") {
		return "", fmt.Errorf("project path %q must be relative to the project root", target)
	}
	cleanTarget := path.Clean(strings.TrimPrefix(target, "./"))
	if cleanTarget == "." || cleanTarget == ".." || strings.HasPrefix(cleanTarget, "../") {
		return "", fmt.Errorf("project path %q escapes the project root", target)
	}
	if path.IsAbs(cleanTarget) {
		return "", fmt.Errorf("project path %q must be relative to the project root", target)
	}
	return filepath.Join(projectDir, filepath.FromSlash(cleanTarget)), nil
}

func normalizeProjectPattern(pattern string) (string, error) {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	if pattern == "" {
		return "", fmt.Errorf("exclude patterns cannot be blank")
	}
	if strings.HasPrefix(pattern, "~") {
		return "", fmt.Errorf("exclude pattern %q must be relative to the project root", pattern)
	}
	cleanPattern := path.Clean(strings.TrimPrefix(pattern, "./"))
	if cleanPattern == "." || cleanPattern == ".." || strings.HasPrefix(cleanPattern, "../") {
		return "", fmt.Errorf("exclude pattern %q escapes the project root", pattern)
	}
	if path.IsAbs(cleanPattern) {
		return "", fmt.Errorf("exclude pattern %q must be relative to the project root", pattern)
	}
	for _, segment := range strings.Split(cleanPattern, "/") {
		if segment == "**" {
			continue
		}
		if _, err := path.Match(segment, ""); err != nil {
			return "", fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
	}
	return cleanPattern, nil
}

func matchProjectPattern(pattern, rel string) bool {
	return matchProjectPatternSegments(strings.Split(pattern, "/"), strings.Split(rel, "/"))
}

func matchProjectPatternSegments(patternSegments, pathSegments []string) bool {
	if len(patternSegments) == 0 {
		return len(pathSegments) == 0
	}
	if patternSegments[0] == "**" {
		if len(patternSegments) == 1 {
			return true
		}
		if matchProjectPatternSegments(patternSegments[1:], pathSegments) {
			return true
		}
		for i := range pathSegments {
			if matchProjectPatternSegments(patternSegments[1:], pathSegments[i+1:]) {
				return true
			}
		}
		return false
	}
	if len(pathSegments) == 0 {
		return false
	}
	matched, err := path.Match(patternSegments[0], pathSegments[0])
	if err != nil || !matched {
		return false
	}
	return matchProjectPatternSegments(patternSegments[1:], pathSegments[1:])
}

func collectProjectPaths(projectDir string) ([]projectPathInfo, error) {
	var paths []projectPathInfo
	err := filepath.WalkDir(projectDir, func(current string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == projectDir {
			return nil
		}
		rel, err := filepath.Rel(projectDir, current)
		if err != nil {
			return err
		}
		paths = append(paths, projectPathInfo{
			abs:   current,
			rel:   filepath.ToSlash(rel),
			isDir: entry.IsDir(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].rel == paths[j].rel {
			return !paths[i].isDir && paths[j].isDir
		}
		return paths[i].rel < paths[j].rel
	})
	return paths, nil
}

func resolveExcludedProjectPaths(patterns []string, projectDir string) ([]projectPathInfo, error) {
	if len(patterns) == 0 {
		return nil, nil
	}
	projectPaths, err := collectProjectPaths(projectDir)
	if err != nil {
		return nil, err
	}

	matchedByTarget := make(map[string]projectPathInfo)
	for _, rawPattern := range patterns {
		pattern, err := normalizeProjectPattern(rawPattern)
		if err != nil {
			return nil, err
		}
		for _, candidate := range projectPaths {
			if matchProjectPattern(pattern, candidate.rel) {
				matchedByTarget[candidate.abs] = candidate
			}
		}
	}

	matches := make([]projectPathInfo, 0, len(matchedByTarget))
	for _, candidate := range matchedByTarget {
		matches = append(matches, candidate)
	}
	sort.Slice(matches, func(i, j int) bool {
		depthI := strings.Count(matches[i].rel, "/")
		depthJ := strings.Count(matches[j].rel, "/")
		if depthI != depthJ {
			return depthI < depthJ
		}
		return matches[i].rel < matches[j].rel
	})

	var pruned []projectPathInfo
	excludedDirs := make(map[string]bool)
	for _, candidate := range matches {
		if hasExcludedAncestor(candidate.rel, excludedDirs) {
			continue
		}
		pruned = append(pruned, candidate)
		if candidate.isDir {
			excludedDirs[candidate.rel] = true
		}
	}
	return pruned, nil
}

func hasExcludedAncestor(rel string, excludedDirs map[string]bool) bool {
	for parent := path.Dir(rel); parent != "." && parent != "/"; parent = path.Dir(parent) {
		if excludedDirs[parent] {
			return true
		}
	}
	return false
}

func parseCopyAsSpec(spec string, projectDir string) (projectCopyAsSpec, error) {
	parts := strings.SplitN(spec, ":", 2)
	if len(parts) != 2 {
		return projectCopyAsSpec{}, fmt.Errorf("invalid copy-as %q; expected src:dst", spec)
	}
	src, err := resolveHostPath(strings.TrimSpace(parts[0]), projectDir)
	if err != nil {
		return projectCopyAsSpec{}, err
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return projectCopyAsSpec{}, fmt.Errorf("copy-as source %q: %w", parts[0], err)
	}
	if srcInfo.IsDir() {
		return projectCopyAsSpec{}, fmt.Errorf("copy-as source %q must be a file, not a directory", parts[0])
	}

	dst, err := resolveProjectRelativeTarget(parts[1], projectDir)
	if err != nil {
		return projectCopyAsSpec{}, err
	}
	dstInfo, err := os.Lstat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return projectCopyAsSpec{}, fmt.Errorf("copy-as destination %q must already exist as a file in the project", parts[1])
		}
		return projectCopyAsSpec{}, fmt.Errorf("copy-as destination %q: %w", parts[1], err)
	}
	if dstInfo.IsDir() {
		return projectCopyAsSpec{}, fmt.Errorf("copy-as destination %q must be a file path, not a directory", parts[1])
	}
	rel, err := filepath.Rel(projectDir, dst)
	if err != nil {
		return projectCopyAsSpec{}, err
	}
	return projectCopyAsSpec{src: src, dst: dst, rel: filepath.ToSlash(rel)}, nil
}

func buildProjectFilterMounts(cfg Config, projectDir string) (string, []string, error) {
	if len(cfg.Exclude) == 0 && len(cfg.CopyAs) == 0 {
		return projectDir, nil, nil
	}

	excludedPaths, err := resolveExcludedProjectPaths(cfg.Exclude, projectDir)
	if err != nil {
		return "", nil, err
	}

	excludedByRel := make(map[string]projectPathInfo, len(excludedPaths))
	specialAncestorDirs := make(map[string]bool)
	for _, excluded := range excludedPaths {
		excludedByRel[excluded.rel] = excluded
		for parent := path.Dir(excluded.rel); parent != "." && parent != "/"; parent = path.Dir(parent) {
			specialAncestorDirs[parent] = true
		}
	}

	copyAsByRel := make(map[string]projectCopyAsSpec, len(cfg.CopyAs))
	for _, rawSpec := range cfg.CopyAs {
		spec, err := parseCopyAsSpec(rawSpec, projectDir)
		if err != nil {
			return "", nil, err
		}
		copyAsByRel[spec.rel] = spec
		for parent := path.Dir(spec.rel); parent != "." && parent != "/"; parent = path.Dir(parent) {
			specialAncestorDirs[parent] = true
		}
	}

	viewRoot, err := os.MkdirTemp(projectDir, ".yolobox-project-view-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp dir for project filtering: %w", err)
	}
	cleanupPaths := []string{viewRoot}
	viewRootBase := filepath.Base(viewRoot)

	var buildDir func(hostDir, relDir string) error
	buildDir = func(hostDir, relDir string) error {
		entries, err := os.ReadDir(hostDir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if relDir == "" && entry.Name() == viewRootBase {
				continue
			}
			relPath := entry.Name()
			if relDir != "" {
				relPath = path.Join(relDir, entry.Name())
			}
			targetPath := filepath.Join(viewRoot, filepath.FromSlash(relPath))
			hostPath := filepath.Join(hostDir, entry.Name())

			if spec, ok := copyAsByRel[relPath]; ok {
				if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
					return err
				}
				data, err := os.ReadFile(spec.src)
				if err != nil {
					return err
				}
				if err := os.WriteFile(targetPath, data, 0644); err != nil {
					return err
				}
				continue
			}

			if excluded, ok := excludedByRel[relPath]; ok {
				if excluded.isDir {
					if err := os.MkdirAll(targetPath, 0755); err != nil {
						return err
					}
				} else {
					if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
						return err
					}
					file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
					if err != nil {
						return err
					}
					if err := file.Close(); err != nil {
						return err
					}
				}
				continue
			}

			if entry.IsDir() {
				if err := os.MkdirAll(targetPath, 0755); err != nil {
					return err
				}
				if specialAncestorDirs[relPath] {
					if err := buildDir(hostPath, relPath); err != nil {
						return err
					}
				}
				continue
			}

			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
		return nil
	}

	if err := buildDir(projectDir, ""); err != nil {
		return "", cleanupPaths, err
	}

	return viewRoot, cleanupPaths, nil
}
