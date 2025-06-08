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

type RuleContent struct {
	Element string
	Attr    map[string]string
	Matches []string
}

type Rule struct {
	Name    string
	Value   int
	Level   Level
	Content RuleContent
}

type Ruleset struct {
	Name   string
	Scores []Rule
}

func NewRuleset(filepath string) (Ruleset, error) {
	data, err := os.ReadFile(filepath)

	if err != nil {
		return Ruleset{}, fmt.Errorf("unable to open ruleset file. Reason: %s", err.Error())
	}

	ruleset := Ruleset{}
	err = yaml.Unmarshal(data, &ruleset)

	if err != nil {
		return Ruleset{}, fmt.Errorf("unable to parse ruleset file. Reason: %s", err.Error())
	}

	for _, rule := range ruleset.Scores {
		if !rule.isValid() {
			return Ruleset{}, fmt.Errorf("unable to parse rule named '%s'. Reason: Invalid rule configurations", rule.Name)
		}
	}

	return ruleset, nil
}

func (rule *Rule) isValid() bool {
	switch rule.Level {
	case ResourceLevel:
		return rule.isResourceRuleValid()
	case ContentLevel:
		return rule.isContentRuleValid()
	}

	return false
}

func (rule *Rule) isResourceRuleValid() bool {
	return len(rule.Content.Matches) != 0
}

func (rule *Rule) isContentRuleValid() bool {
	return len(rule.Content.Matches) != 0 || rule.Content.Element != ""
}
