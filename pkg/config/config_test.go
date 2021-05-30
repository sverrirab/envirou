package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const testConfig = `

[settings]
quiet=1
sort_keys=0

[groups]
one=ONE
two= FIRST, SECOND, 

[custom]
tree= UNO, *DOS*, TRES

[profile:foo]
ONE=one
TWO=first second
THREE=
; TODO: FOUR
FIVE= magic 

`

func TestReadConfig(t *testing.T) {
	file, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("file.Name(): %v", file)
	defer os.Remove(file.Name())
	_, err = file.WriteString(testConfig)
	if err != nil {
		log.Fatal(err)
	}

	config, err := ReadConfiguration(file.Name())
	if err != nil {
		t.Error("Failed to read configuration")
	}
	if config.Quiet != true {
		t.Error("Quiet should be true")
	}
	if config.SortKeys != false {
		t.Error("SortKeys should be false")
	}
	if config.PathTilde != true {
		t.Error("PathTilde should be false")
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
	if config.Quiet != false {
		t.Error("Quiet should be false")
	}
	if config.SortKeys != true {
		t.Error("SortKeys should be true")
	}
	if config.PathTilde != true {
		t.Error("PathTilde should be false")
	}
	if len(config.Groups) != 10 {
		t.Errorf("Unexpeced number of groups: %d", len(config.Groups))
	}
	os.Remove(file.Name())
}
