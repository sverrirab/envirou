package ini

import (
	"os"
	"testing"
)

func TestGetStringMissingSection(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[existing]\nfoo=bar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Missing section should return default
	if got := ini.GetString("missing", "foo", "default"); got != "default" {
		t.Errorf("expected default for missing section, got %q", got)
	}
}

func TestGetStringMissingVariable(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfoo=bar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	if got := ini.GetString("section", "missing", "fallback"); got != "fallback" {
		t.Errorf("expected fallback for missing variable, got %q", got)
	}
}

func TestGetStringNilVariable(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nnilvar\nrealvar=value\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Nil variable should return default
	if got := ini.GetString("section", "nilvar", "default"); got != "default" {
		t.Errorf("expected default for nil variable, got %q", got)
	}
	if got := ini.GetString("section", "realvar", "default"); got != "value" {
		t.Errorf("expected 'value', got %q", got)
	}
}

func TestGetBoolMissingSection(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfoo=1\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	if got := ini.GetBool("missing", "foo", true); got != true {
		t.Error("expected default true for missing section")
	}
	if got := ini.GetBool("missing", "foo", false); got != false {
		t.Error("expected default false for missing section")
	}
}

func TestGetBoolNilVariable(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nnilvar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	if got := ini.GetBool("section", "nilvar", true); got != true {
		t.Error("expected default for nil bool variable")
	}
}

func TestGetBoolValues(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfalsy1=0\nfalsy2=false\nfalsy3=no\nfalsy4=FALSE\ntruthy1=1\ntruthy2=yes\ntruthy3=true\ntruthy4=anything\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"falsy1", "falsy2", "falsy3", "falsy4"} {
		if ini.GetBool("section", name, true) {
			t.Errorf("%s should be false", name)
		}
	}
	for _, name := range []string{"truthy1", "truthy2", "truthy3", "truthy4"} {
		if !ini.GetBool("section", name, false) {
			t.Errorf("%s should be true", name)
		}
	}
}

func TestGetOperatorMissing(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfoo=bar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Missing section
	if got := ini.GetOperator("missing", "foo"); got != OpReplace {
		t.Errorf("expected OpReplace for missing section, got %d", got)
	}
	// Missing variable
	if got := ini.GetOperator("section", "missing"); got != OpReplace {
		t.Errorf("expected OpReplace for missing variable, got %d", got)
	}
}

func TestIsNilMissing(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfoo=bar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Missing section
	if ini.IsNil("missing", "foo") {
		t.Error("missing section should not be nil")
	}
	// Missing variable
	if ini.IsNil("section", "missing") {
		t.Error("missing variable should not be nil")
	}
}

func TestGetAllVariablesMissingSection(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfoo=bar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	vars := ini.GetAllVariables("missing")
	if len(vars) != 0 {
		t.Errorf("expected empty list for missing section, got %v", vars)
	}
}

func TestParseLineEmptyPrependAppend(t *testing.T) {
	// ^= with empty value
	name, value, varType, op := parseLine([]byte("FOO^="))
	if name != "FOO" || value != "" || varType != typeEmpty || op != OpPrepend {
		t.Errorf("empty prepend: got name=%q value=%q type=%d op=%d", name, value, varType, op)
	}

	// += with empty value
	name, value, varType, op = parseLine([]byte("BAR+="))
	if name != "BAR" || value != "" || varType != typeEmpty || op != OpAppend {
		t.Errorf("empty append: got name=%q value=%q type=%d op=%d", name, value, varType, op)
	}
}

func TestExistsFunction(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	file.WriteString("[section]\nfoo=bar\nnilvar\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !ini.Exists("section", "foo") {
		t.Error("foo should exist")
	}
	if !ini.Exists("section", "nilvar") {
		t.Error("nilvar should exist (even though nil)")
	}
	if ini.Exists("section", "missing") {
		t.Error("missing should not exist")
	}
	if ini.Exists("missing", "foo") {
		t.Error("missing section should not have foo")
	}
}
