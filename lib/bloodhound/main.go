package bloodhound

import (
	"bloodhound/lib/client"
	"bloodhound/lib/evaluator"
	"bloodhound/lib/rules"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/html"
)

type Context struct {
	Url   string
	Score int
}

func NewContext(url string) Context {
	return Context{
		Url:   url,
		Score: 0,
	}
}

func (context *Context) AddScore(score int) {
	context.Score += score
}

func Execute(targetUrls []string, ruleset rules.Ruleset, clientConfig client.ClientConfig) []Context {
	log.WithField("size", len(ruleset.Scores)).Info("Initializing process with loaded ruleset")

	client := client.NewClient(clientConfig)

	var contexts []Context

	for _, url := range targetUrls {
		context := NewContext(url)

		resourceScore := evaluator.EvaluateUrl(url, &ruleset)
		context.AddScore(resourceScore)

		log.WithFields(log.Fields{
			"url":   url,
			"score": resourceScore,
		}).Trace("Finished evaluating URL")

		request, _ := http.NewRequest("GET", url, nil)
		response, err := client.Do(request)

		if err != nil {
			log.WithFields(log.Fields{
				"url": url,
				"err": err,
			}).Error("Failed to retrieve page content. Skipping url")

			continue
		}

		if response.StatusCode == http.StatusOK {
			document, _ := html.Parse(response.Body)
			contentScore := evaluator.EvaluateHTML(document, &ruleset)

			log.WithFields(log.Fields{
				"url":   url,
				"score": contentScore,
			}).Trace("Finished evaluating content")

			context.AddScore(contentScore)
		}

		log.WithFields(log.Fields{
			"url":   url,
			"score": context.Score,
		}).Info("Finished scoring URL")

		response.Body.Close()
		contexts = append(contexts, context)
	}

	// Rank URLs by score
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Score > contexts[j].Score
	})

	return contexts
}
