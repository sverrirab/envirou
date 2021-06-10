package config

import (
	"io/ioutil"
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

func readTestConfig(t *testing.T, stringConfig string) *Configuration {
	file, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(stringConfig)
	if err != nil {
		log.Fatal(err)
	}

	config, err := ReadConfiguration(file.Name())
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
	if len(config.SettingsPassword) != 3 {
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
	file, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal(err)
	}
	os.Remove(file.Name())

	// Deleted temp file so it does not exist - this should create the file:
	config, err := ReadConfiguration(file.Name())
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
	if len(config.SettingsPassword) != 2 {
		t.Errorf("Unexpected password: %s", config.SettingsPassword)
	}
	if len(config.SettingsPath) != 6 {
		t.Errorf("Unexpected path: %s", config.SettingsPath)
	}
	if len(config.Groups) != 13 {
		t.Errorf("Unexpeced number of groups: %d", len(config.Groups))
	}
	os.Remove(file.Name())
}

func TestReadDefaultPath(t *testing.T) {
	if len(GetDefaultConfigFilePath()) == 0 {
		t.Error("Failed to read default config path")
	}
}

func validateProfileValue(t *testing.T, config *Configuration, profile string, entry string, expectedValue string) {
	p := config.Profiles[profile]
	value, ok := p.Get(entry)
	if ! ok {
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
