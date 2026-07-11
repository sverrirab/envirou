package ini

import (
	"os"
	"testing"
)

// Encrypted value tokens use RawURL base64 (A-Za-z0-9-_) specifically so
// they can never contain the "+=" or "^=" operator sequences. This guards
// the token format against the whole-line operator scan in parseLine.
func TestParseLineEncryptedToken(t *testing.T) {
	file, err := os.CreateTemp("", "ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	token := "enc:v1:Ab-_yZ09Ab-_yZ09Ab-_yZ09"
	file.WriteString("[profile:x]\nSECRET=" + token + "\nPREPENDED^=" + token + "\n")
	file.Close()

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if got := ini.GetString("profile:x", "SECRET", ""); got != token {
		t.Errorf("token mangled by parser: %q", got)
	}
	if op := ini.GetOperator("profile:x", "SECRET"); op != OpReplace {
		t.Errorf("token line must parse as replace, got operator %d", op)
	}
	if op := ini.GetOperator("profile:x", "PREPENDED"); op != OpPrepend {
		t.Errorf("explicit prepend must still work, got operator %d", op)
	}
}
