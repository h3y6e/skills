package skill

import (
	"cmp"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
)

// DiscoverSkills finds skills in a cloned repository.
// It prefers the skills/ subdirectory; if found, the root level is ignored.
// A valid skill directory must contain SKILL.md. Only one level is scanned.
func DiscoverSkills(repoRoot string) ([]DiscoveredSkill, error) {
	skillsDir := filepath.Join(repoRoot, "skills")
	if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
		skills, err := scanDir(skillsDir)
		if err != nil {
			return nil, err
		}
		if len(skills) > 0 {
			return skills, nil
		}
	}

	return scanDir(repoRoot)
}

func scanDir(dir string) ([]DiscoveredSkill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var skills []DiscoveredSkill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		skillDir := filepath.Join(dir, name)

		if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("stat SKILL.md in %q: %w", skillDir, err)
		}

		hash, err := ComputeHash(skillDir)
		if err != nil {
			return nil, err
		}

		skills = append(skills, DiscoveredSkill{
			Name:         name,
			Dir:          skillDir,
			ComputedHash: hash,
		})
	}

	slices.SortFunc(skills, func(a, b DiscoveredSkill) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return skills, nil
}

// ComputeHash returns a SHA256 hash of the directory contents.
// Files are sorted by path before hashing.
func ComputeHash(dir string) (string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("compute hash for %q: %w", dir, err)
	}

	slices.Sort(files)

	h := sha256.New()
	for _, rel := range files {
		if _, err := io.WriteString(h, rel); err != nil {
			return "", fmt.Errorf("write path %q to hash: %w", rel, err)
		}

		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", fmt.Errorf("read file %q: %w", rel, err)
		}

		if _, err := h.Write(data); err != nil {
			return "", fmt.Errorf("write file %q to hash: %w", rel, err)
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
