package main

import (
	"bloodhound/lib/evaluator"
	rules "bloodhound/lib/rule"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"golang.org/x/net/html"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)
}

func main() {
	// Prepare ruleset
	ruleset, err := rules.NewRuleset("rules.yml")

	if err != nil {
		log.WithField("reason", err).Fatal("Failed to process ruleset")
		return
	}

	log.WithField("size", len(ruleset.Scores)).Info("Initializing process with loaded ruleset")

	// Prepare client
	client := http.Client{
		// CheckRedirect: true,
	}

	url := "https://crawler-test.com/links/repeated_external_links"
	resourceScore := evaluator.EvaluateUrl(url, ruleset)

	log.WithFields(log.Fields{
		"url":   url,
		"score": resourceScore,
	}).Info("Finished evaluating URL")

	response, err := client.Get(url)

	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("Failed to retrieve page content")

		return
	}

	if response.StatusCode == http.StatusOK {
		document, _ := html.Parse(response.Body)
		evaluator.EvaluateHTML(document, *ruleset)
	}

	response.Body.Close()
}
