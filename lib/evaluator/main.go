package evaluator

import (
	"bloodhound/lib/client"
	"bloodhound/lib/rules"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

type EvaluationResult struct {
	Score  int
	Remove bool
}

func NewEvaluationResult(score int, remove bool) EvaluationResult {
	return EvaluationResult{
		Score:  score,
		Remove: remove,
	}
}

func DefaultEvaluationResult() EvaluationResult {
	return NewEvaluationResult(0, false)
}

type Context struct {
	Url      string
	Response *http.Response
	Content  *html.Node
	Score    int
}

func NewContext(targetUrl string) Context {
	return Context{
		Url:      targetUrl,
		Response: nil,
		Content:  nil,
	}
}

func (context *Context) AddScore(score int) {
	context.Score += score
}

func Evaluate(targetUrls []string, ruleset *rules.Ruleset, clientConfig client.ClientConfig) []Context {
	log.WithFields(log.Fields{
		"targetsSize": len(targetUrls),
		"rulesetSize": len(ruleset.Rules),
	}).Trace("Initializing evaluation pipeline")

	maxChannelSize := len(targetUrls)

	// Put context into pipeline
	inputChannel := make(chan Context, maxChannelSize)
	go pipelineInput(targetUrls, inputChannel)

	// Apply resource name rules
	resourceLevelResultChannel := make(chan Context, maxChannelSize)
	go applyResourceNameRules(ruleset, inputChannel, resourceLevelResultChannel)

	// Retrieve resource
	requestResultChannel := make(chan Context, maxChannelSize)
	go retrieveResource(clientConfig, resourceLevelResultChannel, requestResultChannel)

	// Apply content level evaluation
	contentLevelResultChannel := make(chan Context, maxChannelSize)
	go applyContentRules(ruleset, requestResultChannel, contentLevelResultChannel)

	log.WithFields(log.Fields{
		"targetsSize": len(targetUrls),
		"rulesetSize": len(ruleset.Rules),
	}).Info("Initialized evaluation pipeline")

	return pipelineOutput(contentLevelResultChannel)
}

func pipelineInput(targetUrls []string, out chan<- Context) {
	defer close(out)

	for _, targetUrl := range targetUrls {
		log.WithFields(log.Fields{
			"target": targetUrl,
		}).Trace("Created new context")

		out <- NewContext(targetUrl)
	}
}

func pipelineOutput(in <-chan Context) []Context {
	var contexts []Context
	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
			"score":  context.Score,
		}).Info("Finished processing target")

		contexts = append(contexts, context)
	}

	// Rank URLs by score
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Score > contexts[j].Score
	})

	return contexts
}

func applyResourceNameRules(ruleset *rules.Ruleset, in <-chan Context, out chan<- Context) {
	defer close(out)
	resourceRules := ruleset.GetRules(rules.ResourceLevel)

	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
		}).Trace("Started resource level rule evaluation")

		evaluation := EvaluateUrl(&context.Url, &resourceRules)

		log.WithFields(log.Fields{
			"target":     context.Url,
			"evaluation": evaluation,
		}).Trace("Finished resource level rule evaluation")

		if !evaluation.Remove {
			context.AddScore(evaluation.Score)
			out <- context
		}
	}
}

func retrieveResource(clientConfig client.ClientConfig, in <-chan Context, out chan<- Context) {
	defer close(out)
	client := client.NewClient(clientConfig)

	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
		}).Trace("Requesting resource")

		request, _ := http.NewRequest("GET", context.Url, nil)
		response, err := client.Do(request)

		if err != nil {
			log.WithFields(log.Fields{
				"target": context.Url,
				"err":    err.Error(),
			}).Error("Unable to process and request URL: Skipping")

			continue
		}

		if response.StatusCode != http.StatusOK {
			log.WithFields(log.Fields{
				"target":     context.Url,
				"statusCode": response.StatusCode,
			}).Warn("Resource returned non-OK status: Content evaluation will not be available")
			// TODO: Have proper response handling
		} else {
			log.WithFields(log.Fields{
				"target": context.Url,
			}).Trace("Finished requesting resource")

			document, _ := html.Parse(response.Body)

			context.Content = document
			context.Response = response
			out <- context
		}

	}
}

func applyContentRules(ruleset *rules.Ruleset, in <-chan Context, out chan<- Context) {
	defer close(out)
	contentRules := ruleset.GetRules(rules.ContentLevel)

	for context := range in {
		log.WithFields(log.Fields{
			"target": context.Url,
		}).Trace("Started content level evaluation")

		// TODO: Add content type check and call popper content type evaluator
		evaluation := EvaluateHTML(context.Content, contentRules)

		log.WithFields(log.Fields{
			"target":     context.Url,
			"evaluation": evaluation,
		}).Trace("Finished content level evaluation")

		if !evaluation.Remove {
			context.AddScore(evaluation.Score)
			out <- context
		}
	}
}
