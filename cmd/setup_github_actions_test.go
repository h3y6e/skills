package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/h3y6e/skills/cmd"
)

func TestSetupGithubActionsCommand(t *testing.T) {
	t.Run("registers command and help", func(t *testing.T) {
		t.Parallel()

		root := cmd.NewRootCmd("test")
		setupCmd, _, err := root.Find([]string{"setup-github-actions"})
		if err != nil {
			t.Fatalf("Find(setup-github-actions) error = %v", err)
		}
		if setupCmd == nil || setupCmd.Name() != "setup-github-actions" {
			t.Fatalf("Find(setup-github-actions) returned %v", setupCmd)
		}

		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs([]string{"setup-github-actions", "--help"})
		if err := root.Execute(); err != nil {
			t.Fatalf("setup-github-actions --help error: %v", err)
		}

		got := out.String()
		if !strings.Contains(got, "Generate a GitHub Actions workflow") {
			t.Fatalf("help output missing command description: %q", got)
		}
	})

	t.Run("generates workflow in git repository", func(t *testing.T) {
		repoDir := t.TempDir()
		initGitRepo(t, repoDir)
		t.Chdir(repoDir)

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		root := cmd.NewRootCmd("test")
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"setup-github-actions"})
		if err := root.Execute(); err != nil {
			t.Fatalf("setup-github-actions error: %v", err)
		}

		workflowPath := filepath.Join(repoDir, ".github", "workflows", "skills-update.yml")
		data, err := os.ReadFile(workflowPath)
		if err != nil {
			t.Fatalf("read generated workflow: %v", err)
		}

		got := string(data)
		if !strings.Contains(got, "workflow_dispatch:") {
			t.Fatalf("generated workflow missing workflow_dispatch trigger: %q", got)
		}
		if !strings.Contains(got, "cron: '0 9 * * 1'") {
			t.Fatalf("generated workflow missing weekly schedule: %q", got)
		}
		if !strings.Contains(got, "skills update -y -d .agents/skills") {
			t.Fatalf("generated workflow missing default skills update command: %q", got)
		}
		if !strings.Contains(got, "peter-evans/create-pull-request@v7") {
			t.Fatalf("generated workflow missing PR action: %q", got)
		}
		if stderr.Len() != 0 {
			t.Fatalf("unexpected stderr output in git repo: %q", stderr.String())
		}
	})

	t.Run("refuses overwrite without force and overwrites with force", func(t *testing.T) {
		repoDir := t.TempDir()
		initGitRepo(t, repoDir)
		t.Chdir(repoDir)

		workflowPath := filepath.Join(repoDir, ".github", "workflows", "skills-update.yml")
		if err := os.MkdirAll(filepath.Dir(workflowPath), 0o755); err != nil {
			t.Fatalf("mkdir workflow dir: %v", err)
		}
		if err := os.WriteFile(workflowPath, []byte("original\n"), 0o644); err != nil {
			t.Fatalf("seed workflow: %v", err)
		}

		root := cmd.NewRootCmd("test")
		root.SetArgs([]string{"setup-github-actions"})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected overwrite protection error")
		}
		if !strings.Contains(err.Error(), "--force") {
			t.Fatalf("overwrite error should mention --force: %v", err)
		}

		root = cmd.NewRootCmd("test")
		root.SetArgs([]string{"setup-github-actions", "--force"})
		if err := root.Execute(); err != nil {
			t.Fatalf("setup-github-actions --force error: %v", err)
		}

		data, err := os.ReadFile(workflowPath)
		if err != nil {
			t.Fatalf("read overwritten workflow: %v", err)
		}
		if strings.Contains(string(data), "original") {
			t.Fatalf("workflow was not overwritten: %q", string(data))
		}
	})

	t.Run("writes custom dest dir into workflow", func(t *testing.T) {
		repoDir := t.TempDir()
		initGitRepo(t, repoDir)
		t.Chdir(repoDir)

		root := cmd.NewRootCmd("test")
		root.SetArgs([]string{"setup-github-actions", "-d", "dot_agents/exact_skills"})
		if err := root.Execute(); err != nil {
			t.Fatalf("setup-github-actions -d error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(repoDir, ".github", "workflows", "skills-update.yml"))
		if err != nil {
			t.Fatalf("read generated workflow: %v", err)
		}
		if !strings.Contains(string(data), "skills update -y -d dot_agents/exact_skills") {
			t.Fatalf("generated workflow missing custom dest dir: %q", string(data))
		}
	})

	t.Run("warns outside git repository and still writes workflow", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)

		var stderr bytes.Buffer

		root := cmd.NewRootCmd("test")
		root.SetErr(&stderr)
		root.SetArgs([]string{"setup-github-actions"})
		if err := root.Execute(); err != nil {
			t.Fatalf("setup-github-actions outside git repo error: %v", err)
		}

		if !strings.Contains(stderr.String(), "warning: current directory is not a git repository") {
			t.Fatalf("expected git repository warning, got: %q", stderr.String())
		}
		if _, err := os.Stat(filepath.Join(workDir, ".github", "workflows", "skills-update.yml")); err != nil {
			t.Fatalf("generated workflow missing outside git repo: %v", err)
		}
	})
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
}
