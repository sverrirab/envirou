package ini

import (
	"bytes"
	"io/ioutil"
	"sort"
	"strings"
)

const (
	typeNil = iota
	typeEmpty
	typeString
)

// Operator indicates how a variable value should be applied.
const (
	OpReplace = iota // name=value (default)
	OpPrepend        // name^=value
	OpAppend         // name+=value
)


type Variable struct {
	varType  int
	value    string
	Operator int // OpReplace, OpPrepend, or OpAppend
}

type Section struct {
	variables map[string]Variable
}

type IniFile struct {
	sections map[string]Section
}

func NewIni(path string) (*IniFile, error) {
	ini := IniFile{sections: make(map[string]Section)}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(b, []byte{'\n'})
	sectionName := "_" // Default section name
	for _, line := range lines {
		line := bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		first_char := line[0]

		if first_char == ';' || first_char == '#' {
			// Comment - ignore for now.
			continue
		}
		if first_char == '[' && line[len(line)-1] == ']' {
			// Section header
			sectionName = string(bytes.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		varName, varValue, varType, operator := parseLine(line)
		section, ok := ini.sections[sectionName]
		if !ok {
			section = Section{variables: make(map[string]Variable)}
			ini.sections[sectionName] = section
		}
		section.variables[varName] = Variable{varType: varType, value: varValue, Operator: operator}
	}
	return &ini, nil
}

// parseLine extracts variable name, value, type, and operator from an INI line.
// Recognizes ^= (prepend), += (append), and = (replace).
func parseLine(line []byte) (string, string, int, int) {
	operator := OpReplace
	// Check for ^= and += before falling back to =
	if idx := bytes.Index(line, []byte("^=")); idx >= 0 {
		operator = OpPrepend
		name := string(bytes.TrimSpace(line[:idx]))
		value := string(bytes.TrimSpace(line[idx+2:]))
		if len(value) == 0 {
			return name, "", typeEmpty, operator
		}
		return name, value, typeString, operator
	}
	if idx := bytes.Index(line, []byte("+=")); idx >= 0 {
		operator = OpAppend
		name := string(bytes.TrimSpace(line[:idx]))
		value := string(bytes.TrimSpace(line[idx+2:]))
		if len(value) == 0 {
			return name, "", typeEmpty, operator
		}
		return name, value, typeString, operator
	}
	split := bytes.SplitN(line, []byte{'='}, 2)
	varName := string(bytes.TrimSpace(split[0]))
	if len(split) == 2 {
		varValue := string(bytes.TrimSpace(split[1]))
		if len(varValue) == 0 {
			return varName, "", typeEmpty, OpReplace
		}
		return varName, varValue, typeString, OpReplace
	}
	return varName, "", typeNil, OpReplace
}

// GetOperator returns the operator for a variable (OpReplace, OpPrepend, or OpAppend).
func (iniFile *IniFile) GetOperator(section string, name string) int {
	v, ok := iniFile.getVariable(section, name)
	if !ok {
		return OpReplace
	}
	return v.Operator
}

// GetString Read string variable from a section
func (iniFile *IniFile) GetString(section string, name string, defaultValue string) string {
	v, ok := iniFile.getVariable(section, name)
	if ! ok {
		return defaultValue
	}
	if v.varType == typeNil {
		return defaultValue
	}
	return v.value
}

// GetString Read string variable from a section
func (iniFile *IniFile) GetBool(section string, name string, defaultValue bool) bool {
	v, ok := iniFile.getVariable(section, name)
	if ! ok {
		return defaultValue
	}
	if v.varType == typeNil {
		return defaultValue
	}
	switch strings.ToLower(v.value) {
	case "0", "false", "no":
		return false
	}
	return true
}

// Exists Check if variable exists (could be nil)
func (iniFile *IniFile) Exists(section string, name string) bool {
	_, ok := iniFile.getVariable(section, name)
	return ok
}

// IsNil Check if variable is nil only returns true if the variable exists and is set to nil
func (iniFile *IniFile) IsNil(section string, name string) bool {
	v, ok := iniFile.getVariable(section, name)
	if ! ok {
		return false
	}
	return v.varType == typeNil
}

func (iniFile *IniFile) getVariable(section string, name string) (*Variable, bool) {
	s, ok := iniFile.sections[section]
	if ! ok {
		return nil, false
	}
	v, ok := s.variables[name]
	if ! ok {
		return nil, false
	}
	return &v, true
}

// GetAllSections Get sorted list of all sections
func (iniFile *IniFile) GetAllSections() []string {
	keys := make([]string, 0, len(iniFile.sections))
	for key := range iniFile.sections {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// GetAllVariables Get sorted list of all variables
func (iniFile *IniFile) GetAllVariables(section string) []string {
	s, ok := iniFile.sections[section]
	if ! ok {
		return []string{}
	}
	variables := make([]string, 0, len(s.variables))
	for name := range s.variables {
		variables = append(variables, name)
	}
	sort.Strings(variables)
	return variables
}
