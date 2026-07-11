package config

import (
	"log"
	"os"
	"testing"
)

const testConfig = `

[settings]
quiet=1
sort_keys=0
password=FOO*, BAR, *MATCH

[format]
group=SMURF
env_name=underline

[groups]
one=ONE
two= FIRST, SECOND, 

[custom]
tree= UNO, *DOS*, TRES

[ Profile:foo]
ONE=one
TWO=first second
THREE
; TODO: FOUR
FIVE= magic 

`

func removeFile(name string) {
	err := os.Remove(name)
	if err != nil {
		log.Fatal(err)
	}
}

func readTestConfig(t *testing.T, stringConfig string) *Configuration {
	file, err := os.CreateTemp("", "config")
	if err != nil {
		log.Fatal(err)
	}
	name := file.Name()
	defer removeFile(name)
	_, err = file.WriteString(stringConfig)
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	config, err := ReadConfiguration(name, false)
	if err != nil {
		t.Error("Failed to read configuration")
	}
	return config
}

func TestReadConfig(t *testing.T) {
	config := readTestConfig(t, testConfig)
	if config.SettingsQuiet != true {
		t.Error("Quiet should be true")
	}
	if config.SettingsSortKeys != false {
		t.Error("SortKeys should be false")
	}
	if config.SettingsPathTilde != true {
		t.Error("PathTilde should be false")
	}
	if config.FormatGroup != "magenta" {
		t.Errorf("expected magenta")
	}
	if config.FormatProfile != "green" {
		t.Errorf("expected green")
	}
	if config.FormatEnvName != "underline" {
		t.Errorf("expected underline")
	}
	// 3 configured + builtin ENVIROU_KEY
	if len(config.SettingsPassword) != 4 {
		t.Errorf("Unexpected password: %s", config.SettingsPassword)
	}
	if len(config.SettingsPath) != 0 {
		t.Errorf("Unexpected path: %s", config.SettingsPath)
	}
	if len(config.Groups) != 3 {
		t.Errorf("Unexpeced number of groups: %d", len(config.Groups))
	}
}

func TestReadDefault(t *testing.T) {
	file, err := os.CreateTemp("", "config")
	if err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = os.Remove(file.Name())
	if err != nil {
		t.Error("Failed to delete file")
	}

	// Deleted temp file - this should create the file:
	config, err := ReadConfiguration(file.Name(), false)
	if err != nil {
		t.Error("Failed to read configuration")
	}
	if config.SettingsQuiet != false {
		t.Error("Quiet should be false")
	}
	if config.SettingsSortKeys != true {
		t.Error("SortKeys should be true")
	}
	if config.SettingsPathTilde != true {
		t.Error("PathTilde should be false")
	}
	if config.FormatGroup != "magenta" {
		t.Error("expected magenta")
	}
	if config.FormatProfile != "green" {
		t.Error("expected green")
	}
	if config.FormatEnvName != "cyan" {
		t.Error("expected cyan")
	}
	// 6 configured patterns + hardcoded ENVIROU_KEY.
	if len(config.SettingsPassword) != 7 {
		t.Errorf("Unexpected password: %s", config.SettingsPassword)
	}
	if len(config.SettingsPath) != 6 {
		t.Errorf("Unexpected path: %s", config.SettingsPath)
	}
	if len(config.Groups) != 15 {
		t.Errorf("Unexpected number of groups: %d", len(config.Groups))
	}
	removeFile(file.Name())
}

func TestEncryptedProfileValueIsAutomaticallySensitive(t *testing.T) {
	config := readTestConfig(t, testConfig+"\n[profile:encrypted]\nORDINARY_NAME=enc:v1:test-token\n")
	want := map[string]bool{"ENVIROU_KEY": false, "ORDINARY_NAME": false}
	for _, pattern := range config.SettingsPassword {
		if _, ok := want[string(pattern)]; ok {
			want[string(pattern)] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("expected %s to be automatically sensitive; patterns: %v", name, config.SettingsPassword)
		}
	}
}

func TestEncryptionKeyVariableCaseHandling(t *testing.T) {
	if !IsEncryptionKeyVariable("ENVIROU_KEY", false) {
		t.Error("expected exact ENVIROU_KEY match")
	}
	if IsEncryptionKeyVariable("envirou_key", false) {
		t.Error("unexpected case-sensitive match")
	}
	if !IsEncryptionKeyVariable("envirou_key", true) {
		t.Error("expected case-insensitive Windows match")
	}
}

func TestReadDefaultPath(t *testing.T) {
	if len(GetDefaultConfigFilePath()) == 0 {
		t.Error("Failed to read default config path")
	}
}

func validateProfileValue(t *testing.T, config *Configuration, profile string, entry string, expectedValue string) {
	p := config.Profiles[profile]
	value, ok := p.Get(entry)
	if !ok {
		t.Errorf("Missing entry %s in profile %s", entry, profile)
	}
	if value != expectedValue {
		t.Errorf("Entry %s in profile %s - wrong value %s != %s", entry, profile, expectedValue, value)
	}
}

func validateProfileNil(t *testing.T, config *Configuration, profile string, entry string, expectedNil bool) {
	p := config.Profiles[profile]
	isNil := p.GetNil(entry)
	if expectedNil != isNil {
		t.Errorf("Entry %s in profile %s - wrong value %v != %v", entry, profile, expectedNil, isNil)
	}
}

func TestProfile(t *testing.T) {
	config := readTestConfig(t, testConfig)

	validateProfileValue(t, config, "foo", "ONE", "one")
	validateProfileValue(t, config, "foo", "TWO", "first second")
	validateProfileNil(t, config, "foo", "THREE", true)
	validateProfileNil(t, config, "foo", "NOT-THREE", false)

}
