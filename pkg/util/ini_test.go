package util

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"testing"
)

const testConfig = `
default_section=silly
[settings]
quiet=1
sort_keys=0
negative=FALSE

[groups]
one=ONE
two= FIRST, SECOND, 

[Groups]
Capital=yes

[groups]
added=true

# another comment
[custom]


 [ tricky=is#my;nömé ] 
tree= UNO, *DOS*, TRES

; comment
[profile:foo]
ONE=one
TWO = first second
THREE=
FOUR
 FIVE= magic = makes = the = world = go = around

`
func TestMissingConfig(t *testing.T) {
	_, err := NewIni("m i s s i n g - f i l e")	
	if err == nil {
		t.Error("This should fail")
	}
}

func checkString(t *testing.T, iniFile *IniFile, sectionName string, name string, value string) {
	if ! iniFile.Exists(sectionName, name) {
		t.Errorf("String variable not found \"%s\" [%s]", sectionName, name)
	}
	val := iniFile.GetString(sectionName, name, "")
	if val != value {
		t.Errorf("String variable \"%s\" [%s] value mismatch: \"%s\" != \"%s\"", sectionName, name, value, val)
	}
}

func checkBool(t *testing.T, iniFile *IniFile, sectionName string, name string, value bool) {
	if ! iniFile.Exists(sectionName, name) {
		t.Errorf("Bool variable not found \"%s\" [%s]", sectionName, name)
	}
	val := iniFile.GetBool(sectionName, name, false)
	if val != value {
		t.Errorf("Bool variable \"%s\" [%s] value mismatch: \"%v\" != \"%v\"", sectionName, name, value, val)
	}
}

func checkNil(t *testing.T, iniFile *IniFile, sectionName string, name string) {
	if ! iniFile.Exists(sectionName, name) {
		t.Errorf("Variable not found \"%s\" [%s]", sectionName, name)
	}
	if isNil := iniFile.IsNil(sectionName, name); ! isNil {
		t.Errorf("Variable \"%s\" [%s] not nil!", sectionName, name)
	}
}

func checkType(t *testing.T, iniFile *IniFile, sectionName string, name string, varType int) {
	v, ok := iniFile.getVariable(sectionName, name)
	if ! ok {
		t.Errorf("Variable not found \"%s\" [%s]", sectionName, name)
		
	} else {
		if varType != v.varType {
			t.Errorf("String variable \"%s\" [%s] type mismatch: \"%d\" != \"%d\"", sectionName, name, varType, v.varType)
		}
	}
}

func TestReadConfig(t *testing.T) {
	file, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())
	_, err = file.WriteString(testConfig)
	if err != nil {
		log.Fatal(err)
	}

	ini, err := NewIni(file.Name())
	if err != nil {
		t.Error("Failed to read configuration")
	}

	checkString(t, ini, "_", "default_section", "silly")
	checkString(t, ini, "settings", "quiet", "1")
	checkString(t, ini, "groups", "added", "true")
	checkString(t, ini, "Groups", "Capital", "yes")
	checkString(t, ini, "tricky=is#my;nömé", "tree", "UNO, *DOS*, TRES")
	checkString(t, ini, "profile:foo", "ONE", "one")
	checkString(t, ini, "profile:foo", "TWO", "first second")
	checkString(t, ini, "profile:foo", "THREE", "")
	checkString(t, ini, "profile:foo", "FIVE", "magic = makes = the = world = go = around")

	checkBool(t, ini, "settings", "quiet", true)
	checkBool(t, ini, "settings", "sort_keys", false)
	checkBool(t, ini, "settings", "negative", false)

	checkType(t, ini, "profile:foo", "THREE", typeEmpty)
	checkType(t, ini, "profile:foo", "FOUR", typeNil)
	checkType(t, ini, "profile:foo", "FIVE", typeString)
	checkNil(t, ini, "profile:foo", "FOUR")

	sections := ini.GetAllSections()
	if len(sections) != 6 {
		t.Errorf("Section length mismatch %d", len(sections))
	}
	if sort.SearchStrings(sections, "groups") != 2 {
		t.Errorf("Section order mismatch %d", sort.SearchStrings(sections, "groups"))
	}

	for _, section := range sections {
		variables := ini.GetAllVariables(section)
		if len(variables) == 0 {
			t.Errorf("Missing variables from section %s?", section)
		}
	}
}
