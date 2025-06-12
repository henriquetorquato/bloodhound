package rules

type RuleContent struct {
	Element string
	Attr    map[string]string
	Matches []string
}

type Rule struct {
	Name    string
	Value   int
	Level   Level
	Remove  bool
	Content RuleContent
}

func NewMatchRuleContent(matches []string) RuleContent {
	return RuleContent{
		Matches: matches,
	}
}

func NewElementRuleContent(element string, attr map[string]string) RuleContent {
	return RuleContent{
		Element: element,
		Attr:    attr,
	}
}

func NewResourceRule(name string, value int, remove bool, content RuleContent) Rule {
	return NewRule(name, ResourceLevel, value, remove, content)
}

func NewContentRule(name string, value int, remove bool, content RuleContent) Rule {
	return NewRule(name, ContentLevel, value, remove, content)
}

func NewRule(name string, level Level, value int, remove bool, content RuleContent) Rule {
	return Rule{
		Name:    name,
		Level:   level,
		Value:   value,
		Remove:  remove,
		Content: content,
	}
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
