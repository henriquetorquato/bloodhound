package pipeline

import "golang.org/x/net/html"

type Context struct {
	Url     string
	Content *html.Node
	Score   int
}

func NewContext(targetUrl string) Context {
	return Context{
		Url:     targetUrl,
		Content: nil,
	}
}

func (context *Context) AddScore(score int) {
	context.Score += score
}
