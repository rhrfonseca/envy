package cli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "envy-test-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create temp dir:", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(tmp, "envy")
	root, _ := filepath.Abs("../")
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/envy/")
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to build envy:", err)
		os.RemoveAll(tmp)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func run(t *testing.T, stdin string, env []string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		code = exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "envy-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.WriteString(content)
	f.Close()
	return f.Name()
}

func TestCLI_stdout(t *testing.T) {
	f := writeTempFile(t, "host: ${HOST}\nport: ${PORT:-5432}\n")
	stdout, _, code := run(t, "", []string{"HOST=localhost"}, f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "host: localhost") {
		t.Errorf("expected host substitution, got %q", stdout)
	}
	if !strings.Contains(stdout, "port: 5432") {
		t.Errorf("expected default port, got %q", stdout)
	}
}

func TestCLI_missingVars(t *testing.T) {
	f := writeTempFile(t, "a: ${FOO}\nb: ${BAR}\n")
	_, stderr, code := run(t, "", nil, f)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "BAR") || !strings.Contains(stderr, "FOO") {
		t.Errorf("expected both missing vars in stderr, got %q", stderr)
	}
}

func TestCLI_tmp(t *testing.T) {
	f := writeTempFile(t, "key: ${VAL}\n")
	stdout, _, code := run(t, "", []string{"VAL=hello"}, "--tmp", f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	tmpPath := strings.TrimSpace(stdout)
	if !strings.HasPrefix(tmpPath, "/tmp/envy-") {
		t.Errorf("expected /tmp/envy-* path, got %q", tmpPath)
	}
	t.Cleanup(func() { os.Remove(tmpPath) })
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("tmp file not found: %v", err)
	}
	if !strings.Contains(string(content), "key: hello") {
		t.Errorf("unexpected tmp file content: %q", content)
	}
}

func TestCLI_tmpFilenameContainsBaseName(t *testing.T) {
	f := writeTempFile(t, "key: ${VAL}\n")
	base := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))
	stdout, _, code := run(t, "", []string{"VAL=hello"}, "--tmp", f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	tmpPath := strings.TrimSpace(stdout)
	t.Cleanup(func() { os.Remove(tmpPath) })
	if !strings.Contains(tmpPath, base) {
		t.Errorf("expected base name %q in path %q", base, tmpPath)
	}
}

func TestCLI_out(t *testing.T) {
	f := writeTempFile(t, "key: ${VAL}\n")
	outFile := f + ".out"
	t.Cleanup(func() { os.Remove(outFile) })
	stdout, _, code := run(t, "", []string{"VAL=world"}, "--out", outFile, f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if stdout != "" {
		t.Errorf("expected silent stdout, got %q", stdout)
	}
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("out file not found: %v", err)
	}
	if !strings.Contains(string(content), "key: world") {
		t.Errorf("unexpected out file content: %q", content)
	}
}

func TestCLI_tmpAndOutMutuallyExclusive(t *testing.T) {
	f := writeTempFile(t, "key: value\n")
	_, stderr, code := run(t, "", nil, "--tmp", "--out", "/tmp/x.yaml", f)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected mutually exclusive error, got %q", stderr)
	}
}

func TestCLI_stdin(t *testing.T) {
	stdout, _, code := run(t, "host: ${HOST}\n", []string{"HOST=from-stdin"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "host: from-stdin") {
		t.Errorf("got %q", stdout)
	}
}

func TestCLI_stdinTmpFilename(t *testing.T) {
	stdout, _, code := run(t, "key: ${VAL}\n", []string{"VAL=x"}, "--tmp")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	tmpPath := strings.TrimSpace(stdout)
	t.Cleanup(func() { os.Remove(tmpPath) })
	if !strings.Contains(tmpPath, "envy-stdin-") {
		t.Errorf("expected envy-stdin-* path, got %q", tmpPath)
	}
}

func TestCLI_dotenv(t *testing.T) {
	envFile := writeTempFile(t, "DB_HOST=from-dotenv\nDB_PORT=3306\n")
	f := writeTempFile(t, "host: ${DB_HOST}\nport: ${DB_PORT}\n")
	stdout, _, code := run(t, "", nil, "--dotenv", envFile, f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "host: from-dotenv") || !strings.Contains(stdout, "port: 3306") {
		t.Errorf("got %q", stdout)
	}
}

func TestCLI_dotenvFirstWins(t *testing.T) {
	first := writeTempFile(t, "KEY=from-first\n")
	second := writeTempFile(t, "KEY=from-second\n")
	f := writeTempFile(t, "value: ${KEY}\n")
	stdout, _, code := run(t, "", nil, "--dotenv", first, "--dotenv", second, f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "value: from-first") {
		t.Errorf("expected first dotenv to win, got %q", stdout)
	}
}

func TestCLI_dotenvReplacesStandardLookup(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY=standard\n"), 0644)
	explicitEnv := writeTempFile(t, "KEY=explicit\n")
	f := writeTempFile(t, "value: ${KEY}\n")

	cmd := exec.Command(binaryPath, "--dotenv", explicitEnv, f)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("exit error: %v", err)
	}
	if !strings.Contains(string(out), "value: explicit") {
		t.Errorf("expected explicit dotenv to replace standard lookup, got %q", out)
	}
}

func TestCLI_envFlag(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY=base\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".env.production"), []byte("KEY=production\n"), 0644)
	f := writeTempFile(t, "value: ${KEY}\n")

	cmd := exec.Command(binaryPath, "--env", "production", f)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("exit error: %v", err)
	}
	if !strings.Contains(string(out), "value: production") {
		t.Errorf("expected production env value, got %q", out)
	}
}

func TestCLI_envFlagLocalOverrides(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env.production"), []byte("KEY=production\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".env.production.local"), []byte("KEY=production-local\n"), 0644)
	f := writeTempFile(t, "value: ${KEY}\n")

	cmd := exec.Command(binaryPath, "--env", "production", f)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("exit error: %v", err)
	}
	if !strings.Contains(string(out), "value: production-local") {
		t.Errorf("expected production.local to win, got %q", out)
	}
}

func TestCLI_shellEnvWinsOverDotenv(t *testing.T) {
	envFile := writeTempFile(t, "KEY=from-file\n")
	f := writeTempFile(t, "value: ${KEY}\n")
	stdout, _, code := run(t, "", []string{"KEY=from-shell"}, "--dotenv", envFile, f)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stdout, "value: from-shell") {
		t.Errorf("expected shell env to win, got %q", stdout)
	}
}

func TestCLI_invalidYAMLBefore(t *testing.T) {
	f := writeTempFile(t, "key: [unclosed\n")
	_, stderr, code := run(t, "", nil, f)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid YAML") {
		t.Errorf("expected invalid YAML error, got %q", stderr)
	}
}

func TestCLI_invalidYAMLAfter(t *testing.T) {
	f := writeTempFile(t, "key: ${VAL}\n")
	_, stderr, code := run(t, "", []string{"VAL={broken: [unclosed"}, f)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "invalid YAML after substitution") {
		t.Errorf("expected post-substitution YAML error, got %q", stderr)
	}
}

func TestCLI_help(t *testing.T) {
	stdout, _, code := run(t, "", nil, "--help")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, expected := range []string{"Usage", "--env", "--dotenv", "--tmp", "--out", "${VAR}", "Dotenv precedence"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("help missing %q", expected)
		}
	}
}
