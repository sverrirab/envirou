package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sverrirab/envirou/pkg/data"
	"github.com/sverrirab/envirou/pkg/ini"
)

func SaveSnapshot(profile *data.Profile, groups *data.Groups, caseInsensitive bool) error {
	var b strings.Builder
	b.WriteString("[meta]\n")
	b.WriteString(fmt.Sprintf("timestamp=%s\n", time.Now().Format(time.RFC3339)))
	b.WriteString("\n[snapshot]\n")
	for _, name := range profile.SortedNames(false) {
		if groups.IsIgnored(name, caseInsensitive) {
			continue
		}
		value, _ := profile.Get(name)
		b.WriteString(fmt.Sprintf("%s=%s\n", name, value))
	}
	err := os.MkdirAll(GetDefaultConfigFileFolder(), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(GetSnapshotFilePath(), []byte(b.String()), 0644)
}

func LoadSnapshot(caseInsensitive bool) (*data.Profile, error) {
	path := GetSnapshotFilePath()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	iniFile, err := ini.NewIni(path)
	if err != nil {
		return nil, err
	}
	profile := data.NewProfile(caseInsensitive)
	for _, name := range iniFile.GetAllVariables("snapshot") {
		if iniFile.IsNil("snapshot", name) {
			profile.SetNil(name)
		} else {
			profile.Set(name, iniFile.GetString("snapshot", name, ""))
		}
	}
	return profile, nil
}

func RemoveSnapshot() error {
	err := os.Remove(GetSnapshotFilePath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
