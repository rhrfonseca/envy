package envy

import "testing"

func TestSubstitute_midString(t *testing.T) {
	out, missing := Substitute("url: https://${HOST}/api", map[string]string{"HOST": "example.com"})
	if out != "url: https://example.com/api" {
		t.Errorf("got %q", out)
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}

func TestSubstitute_defaultUsed(t *testing.T) {
	out, missing := Substitute("port: ${PORT:-5432}", map[string]string{})
	if out != "port: 5432" {
		t.Errorf("got %q", out)
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}

func TestSubstitute_defaultIgnoredWhenVarSet(t *testing.T) {
	out, missing := Substitute("port: ${PORT:-5432}", map[string]string{"PORT": "9999"})
	if out != "port: 9999" {
		t.Errorf("got %q", out)
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}

func TestSubstitute_escape(t *testing.T) {
	out, missing := Substitute("key: $${NOT_A_VAR}", map[string]string{})
	if out != "key: ${NOT_A_VAR}" {
		t.Errorf("got %q", out)
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}

func TestSubstitute_inComment(t *testing.T) {
	out, _ := Substitute("# connects to ${HOST}", map[string]string{"HOST": "db.local"})
	if out != "# connects to db.local" {
		t.Errorf("got %q", out)
	}
}

func TestSubstitute_missingVarsCollectedAndDeduplicated(t *testing.T) {
	_, missing := Substitute("a: ${FOO}\nb: ${BAR}\nc: ${FOO}", map[string]string{})
	if len(missing) != 2 {
		t.Errorf("expected 2 missing vars, got %v", missing)
	}
	seen := map[string]bool{}
	for _, v := range missing {
		if seen[v] {
			t.Errorf("duplicate missing var: %s", v)
		}
		seen[v] = true
	}
}

func TestSubstitute_noPatterns(t *testing.T) {
	content := "key: value\nother: 123"
	out, missing := Substitute(content, map[string]string{})
	if out != content {
		t.Errorf("got %q", out)
	}
	if len(missing) != 0 {
		t.Errorf("unexpected missing: %v", missing)
	}
}
