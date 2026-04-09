package cmd

import (
	"fmt"

	"github.com/h3y6e/skills/internal/git"
	"github.com/h3y6e/skills/internal/lock"
	"github.com/h3y6e/skills/internal/skill"
	"github.com/spf13/cobra"
)

func runUpdate(cmd *cobra.Command, source string, skillFilter []string, destDir string, yesFlag bool) error {
	layout := lock.NewLayout(destDir)
	lf, err := loadLockFile(layout)
	if err != nil {
		return err
	}
	entriesForDest := lock.FilterEntriesByDest(lf.Skills, layout.DestDir)

	entries, err := filterEntriesBySource(entriesForDest, source)
	if err != nil {
		return err
	}

	cloneFn, cleanup := skill.NewCloneFunc(cmd.Context(), "skills-update-*")
	defer cleanup()

	candidates, skipped, err := skill.AggregateUpdateCandidates(entries, cloneFn)
	if err != nil {
		return err
	}

	candidates = skill.FilterCandidates(candidates, skillFilter)

	if len(candidates) == 0 && len(skipped) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no installed skills to update")
		return nil
	}

	w := cmd.OutOrStdout()
	printSkippedEntries(w, skipped)

	if len(candidates) == 0 {
		fmt.Fprintln(w, "no matching skills to update")
		return nil
	}

	mode := ResolveApprovalMode(yesFlag, IsTTY())

	// Check if any updates are available.
	hasUpdates := false
	for _, c := range candidates {
		if c.Status == skill.StatusUpdateAvailable {
			hasUpdates = true
			break
		}
	}

	// Non-interactive: show preview and return an error.
	if hasUpdates && mode == ApprovalModePreview {
		for _, c := range candidates {
			fmt.Fprintf(w, "%s: %s\n", c.SkillName, c.Status)
		}
		return fmt.Errorf("updates available; re-run with -y or in a TTY to apply")
	}

	applicable := make([]skill.UpdateCandidate, 0, len(candidates))

	for _, c := range candidates {
		switch c.Status {
		case skill.StatusUpToDate:
			fmt.Fprintf(w, "%s: %s\n", c.SkillName, c.Status)

		case skill.StatusUpdateAvailable:
			diffOutput, hasDiff, diffErr := git.DiffNoIndex(cmd.Context(), layout.SkillDir(c.SkillName), c.StagedDir)
			if diffErr != nil {
				fmt.Fprintf(w, "%s: diff error: %v\n", c.SkillName, diffErr)
				continue
			}
			if hasDiff {
				fmt.Fprintf(w, "--- %s ---\n%s\n", c.SkillName, diffOutput)
			}

			applicable = append(applicable, c)

		case skill.StatusCheckFailed:
			fmt.Fprintf(w, "%s: %s (%s)\n", c.SkillName, c.Status, c.Reason)
		}
	}

	_, appliedNames, err := skill.ApplyCandidateUpdates(lf, applicable, layout)
	if err != nil {
		return err
	}

	for _, name := range appliedNames {
		fmt.Fprintf(w, "%s: updated\n", name)
	}

	return nil
}

func filterEntriesBySource(entries map[string]lock.Entry, source string) (map[string]lock.Entry, error) {
	if source == "" {
		return entries, nil
	}

	ref, err := skill.ParseSource(source)
	if err != nil {
		return nil, err
	}

	filtered := make(map[string]lock.Entry)
	for name, entry := range entries {
		if entry.Source != ref.CanonicalSource {
			continue
		}
		if ref.Ref != "" && entry.Ref != ref.Ref {
			continue
		}
		if ref.Ref == "" || entry.Ref == ref.Ref {
			filtered[name] = entry
		}
	}

	return filtered, nil
}
