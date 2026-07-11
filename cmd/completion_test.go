package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// Completion scripts must arrive on stdout so that
// "envirou completion SHELL > file" works.
func TestCompletionZsh(t *testing.T) {
	out := executeCommand(t, "completion", "zsh")
	if !strings.HasPrefix(out, "#compdef envirou") {
		t.Errorf("Expected zsh completion script on stdout, got: %.80s", out)
	}
	if !strings.Contains(out, "_envirou()") {
		t.Error("Expected _envirou function in zsh completion script")
	}
	if !strings.Contains(out, "compdef _envirou ev") {
		t.Error("Expected zsh completion script to register the ev wrapper")
	}
}

func TestCompletionBash(t *testing.T) {
	out := executeCommand(t, "completion", "bash")
	if !strings.Contains(out, "__start_envirou") {
		t.Errorf("Expected bash completion script on stdout, got: %.80s", out)
	}
	if !strings.Contains(out, "complete -o default -F __start_envirou ev") {
		t.Error("Expected bash completion script to register the ev wrapper")
	}
}

func TestCompletionFish(t *testing.T) {
	out := executeCommand(t, "completion", "fish")
	if !strings.Contains(out, "complete -c envirou") {
		t.Errorf("Expected fish completion script on stdout, got: %.80s", out)
	}
}

func TestCompletionPowershell(t *testing.T) {
	out := executeCommand(t, "completion", "powershell")
	if !strings.Contains(out, "Register-ArgumentCompleter") {
		t.Errorf("Expected PowerShell completion script on stdout, got: %.80s", out)
	}
	if !strings.Contains(out, "Register-ArgumentCompleter -CommandName 'ev'") {
		t.Error("Expected PowerShell completion script to register the ev wrapper")
	}
}

func TestDynamicCompletionUsesStdout(t *testing.T) {
	resetState(t)

	oldStdout, oldStderr := os.Stdout, os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = stdoutW, stderrW
	t.Cleanup(func() {
		os.Stdout, os.Stderr = oldStdout, oldStderr
		setCommandOutput(rootCmd, os.Stderr)
	})

	type result struct {
		stream string
		text   string
	}
	results := make(chan result, 2)
	go func() {
		var b bytes.Buffer
		_, _ = io.Copy(&b, stdoutR)
		results <- result{stream: "stdout", text: b.String()}
	}()
	go func() {
		var b bytes.Buffer
		_, _ = io.Copy(&b, stderrR)
		results <- result{stream: "stderr", text: b.String()}
	}()

	rootCmd.SetArgs([]string{"__complete", "completion", ""})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	_ = stdoutW.Close()
	_ = stderrW.Close()

	output := map[string]string{}
	for range 2 {
		r := <-results
		output[r.stream] = r.text
	}

	for _, candidate := range []string{"bash", "zsh", "fish", "powershell", ":4"} {
		if !strings.Contains(output["stdout"], candidate+"\n") {
			t.Errorf("stdout missing completion %q; got %q", candidate, output["stdout"])
		}
	}
	if strings.Contains(output["stderr"], "bash\n") || strings.Contains(output["stderr"], "powershell\n") {
		t.Errorf("completion candidates leaked to stderr: %q", output["stderr"])
	}
}

func TestSetCompletionProfiles(t *testing.T) {
	t.Setenv("TEST_ENV", "not-active")

	out := executeCommand(t, "__complete", "set", "")
	for _, profile := range []string{"dev", "prod", "tools", "venv"} {
		if !strings.Contains(out, profile+"\t") {
			t.Errorf("set completion missing profile %q; got %q", profile, out)
		}
	}
	if !strings.Contains(out, ":4\n") {
		t.Errorf("set completion must disable file completion; got %q", out)
	}
}

func TestSetCompletionFiltersPrefixAndUsedProfiles(t *testing.T) {
	out := executeCommand(t, "__complete", "set", "dev", "pr")
	if !strings.Contains(out, "prod\t") {
		t.Errorf("set completion should include matching profile prod; got %q", out)
	}
	for _, unexpected := range []string{"dev\t", "tools\t", "venv\t"} {
		if strings.Contains(out, unexpected) {
			t.Errorf("set completion unexpectedly included %q; got %q", unexpected, out)
		}
	}

	out = executeCommand(t, "__complete", "set", "dev", "")
	if strings.Contains(out, "dev\t") {
		t.Errorf("set completion should not repeat an already selected profile; got %q", out)
	}
}

func TestSetCompletionLabelsActiveProfile(t *testing.T) {
	t.Setenv("TEST_ENV", "development")
	out := executeCommand(t, "__complete", "set", "d")
	if !strings.Contains(out, "dev\tactive profile\n") {
		t.Errorf("set completion should label dev active; got %q", out)
	}
}

func TestCompletionInvalidArg(t *testing.T) {
	resetState(t)
	rootCmd.SetArgs([]string{"completion", "fish-food"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("Expected error for invalid shell type")
	}
}

func TestCompletionNoArg(t *testing.T) {
	resetState(t)
	rootCmd.SetArgs([]string{"completion"})
	if err := rootCmd.Execute(); err == nil {
		t.Error("Expected error when no shell type provided")
	}
}
