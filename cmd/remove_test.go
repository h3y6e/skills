package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/h3y6e/skills/cmd"
	"github.com/h3y6e/skills/internal/lock"
)

func TestRemoveFlags(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("test")
	removeCmd, _, _ := root.Find([]string{"remove"})

	t.Run("--dest flag is registered", func(t *testing.T) {
		t.Parallel()
		flag := removeCmd.Flags().Lookup("dest")
		if flag == nil {
			t.Fatal("expected --dest flag on remove command")
		}
	})
}

func TestRemoveAlias(t *testing.T) {
	t.Parallel()

	bareURL := initBareRepo(t, map[string]string{
		"skills/alpha/SKILL.md": "# Alpha\n",
	})

	destDir := filepath.Join(t.TempDir(), ".agents", "skills")

	// Install.
	root1 := cmd.NewRootCmd("test")
	root1.SetArgs([]string{"add", "-d", destDir, bareURL})
	if err := root1.Execute(); err != nil {
		t.Fatalf("add error: %v", err)
	}

	// Remove via rm alias.
	var out bytes.Buffer
	root2 := cmd.NewRootCmd("test")
	root2.SetOut(&out)
	root2.SetArgs([]string{"rm", "-d", destDir, "alpha"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("rm error: %v", err)
	}

	// Skill directory should be gone.
	if _, err := os.Stat(filepath.Join(destDir, "alpha")); !os.IsNotExist(err) {
		t.Error("expected alpha directory to be removed via rm alias")
	}

	// Lockfile should not contain alpha.
	lf, err := lock.ReadFile(lock.FilePath(destDir))
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	if _, ok := lf.Skills["alpha"]; ok {
		t.Error("expected alpha to be removed from lockfile via rm alias")
	}
}

func TestRemove(t *testing.T) {
	t.Parallel()

	t.Run("removes skill directory and lockfile entry", func(t *testing.T) {
		t.Parallel()

		bareURL := initBareRepo(t, map[string]string{
			"skills/alpha/SKILL.md": "# Alpha\n",
			"skills/beta/SKILL.md":  "# Beta\n",
		})

		destDir := filepath.Join(t.TempDir(), ".agents", "skills")

		// Install both skills.
		root1 := cmd.NewRootCmd("test")
		root1.SetArgs([]string{"add", "-d", destDir, bareURL})
		if err := root1.Execute(); err != nil {
			t.Fatalf("add error: %v", err)
		}

		// Remove only alpha.
		root2 := cmd.NewRootCmd("test")
		root2.SetArgs([]string{"remove", "-d", destDir, "alpha"})
		if err := root2.Execute(); err != nil {
			t.Fatalf("remove error: %v", err)
		}

		// Alpha directory should be gone.
		if _, err := os.Stat(filepath.Join(destDir, "alpha")); !os.IsNotExist(err) {
			t.Error("expected alpha directory to be removed")
		}

		// Beta directory should still exist.
		if _, err := os.Stat(filepath.Join(destDir, "beta", "SKILL.md")); err != nil {
			t.Error("expected beta to remain after removing alpha")
		}

		// Lockfile should only have beta.
		lf, err := lock.ReadFile(lock.FilePath(destDir))
		if err != nil {
			t.Fatalf("read lockfile: %v", err)
		}
		if _, ok := lf.Skills["alpha"]; ok {
			t.Error("alpha should not be in lockfile after removal")
		}
		if _, ok := lf.Skills["beta"]; !ok {
			t.Error("beta should still be in lockfile after removing alpha")
		}
	})

	t.Run("errors when skill exists in a different dest", func(t *testing.T) {
		t.Parallel()

		rootDir := t.TempDir()
		destA := filepath.Join(rootDir, ".agents", "skills")
		destB := filepath.Join(rootDir, ".config", "opencode", "skills")

		lockData := []byte(`{
		  "version": 1,
		  "skills": {
		    "beta": {
		      "source": "h3y6e/spec-skills",
		      "sourceType": "github",
		      "computedHash": "abc123",
		      "dest": ` + strconv.Quote(destB) + `
		    }
		  }
		}`)
		if err := os.MkdirAll(filepath.Dir(lock.FilePath(destA)), 0o755); err != nil {
			t.Fatalf("mkdir lockfile parent: %v", err)
		}
		if err := os.WriteFile(lock.FilePath(destA), lockData, 0o644); err != nil {
			t.Fatalf("write lockfile: %v", err)
		}

		root := cmd.NewRootCmd("test")
		root.SetArgs([]string{"remove", "-d", destA, "beta"})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error when removing skill installed in a different dest")
		}
	})

	t.Run("errors when skill not in lockfile", func(t *testing.T) {
		t.Parallel()

		destDir := filepath.Join(t.TempDir(), ".agents", "skills")

		// Create a lockfile with one skill.
		lf := lock.File{
			Version: 1,
			Skills: map[string]lock.Entry{
				"alpha": {Source: "h3y6e/spec-skills", SourceType: "github", ComputedHash: "abc123"},
			},
		}
		lock.WriteFile(lock.FilePath(destDir), lf)

		// Try to remove a skill that doesn't exist.
		root := cmd.NewRootCmd("test")
		root.SetArgs([]string{"remove", "-d", destDir, "nonexistent"})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error when removing nonexistent skill")
		}
	})

	t.Run("errors when lockfile missing", func(t *testing.T) {
		t.Parallel()

		destDir := filepath.Join(t.TempDir(), ".agents", "skills")
		root := cmd.NewRootCmd("test")
		root.SetArgs([]string{"remove", "-d", destDir, "alpha"})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error when lockfile missing")
		}
	})
}
