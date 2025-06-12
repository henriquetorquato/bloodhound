package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Level string

const (
	UnknownLevel  Level = ""
	ResourceLevel Level = "resource"
	ContentLevel  Level = "content"
)

type Ruleset struct {
	Name  string
	Rules []Rule
}

func NewRuleset(filepath string) (*Ruleset, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return &Ruleset{}, fmt.Errorf("unable to open ruleset file. Reason: %s", err.Error())
	}

	ruleset := Ruleset{}
	err = yaml.Unmarshal(data, &ruleset)

	if err != nil {
		return &Ruleset{}, fmt.Errorf("unable to parse ruleset file. Reason: %s", err.Error())
	}

	for _, rule := range ruleset.Rules {
		if !rule.isValid() {
			return &Ruleset{}, fmt.Errorf("unable to parse rule named '%s'. Reason: Invalid rule configurations", rule.Name)
		}
	}

	return &ruleset, nil
}

func (ruleset *Ruleset) GetRules(level Level) []Rule {
	var result []Rule
	for _, rule := range ruleset.Rules {
		if rule.Level != level {
			continue
		}

		result = append(result, rule)
	}

	return result
}
