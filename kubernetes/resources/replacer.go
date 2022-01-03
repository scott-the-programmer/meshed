package resources

import (
	"bytes"
	"io/ioutil"
	"strings"
)

type Replacer struct {
	Replacements map[string]string
	cwd          string
}

func NewReplacer(cwd string) *Replacer {
	return &Replacer{
		Replacements: make(map[string]string),
		cwd:          cwd,
	}
}

func (replacer *Replacer) Add(key string, value string) {
	replacer.Replacements[key] = value
}

func (replacer *Replacer) Replace(template string) (string, error) {
	input, err := ioutil.ReadFile(replacer.cwd + template)
	if err != nil {
		return "", err
	}

	var result []byte = input

	for key, value := range replacer.Replacements {
		result = bytes.Replace(result, []byte("$"+key), []byte(value), -1)
	}

	newFile := strings.Replace(replacer.cwd+template, ".template", "", 1)

	err = ioutil.WriteFile(newFile, result, 0666)

	return newFile, err
}
