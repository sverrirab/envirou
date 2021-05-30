package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/sverrirab/envirou/pkg/util"
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
FOUR
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
	util.Printlnf("%v", config)
	if err != nil {
		t.Error("Failed to read configuration")
	}
	if config.Quiet != true {
		t.Error("Quiet should be true")
	}
	if config.SortKeys != false {
		t.Error("SortKeys should be false")
	}
	if config.PathTilde != false {
		t.Error("PathTilde should be false")
	}
	if len(config.Groups) != 3 {
		t.Errorf("Unexpeced number of groups: %d", len(config.Groups))
	}
}
