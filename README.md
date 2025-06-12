# Bloodhound

Bloodhound is a web resource evaluation tool, used to score and rank a input list of URLs from most to less interesting (or likely to be vulnerable) on a bug bounty context.

For each URL, this tool will retrieve, interpret and score it according to a set of customizable parameters.

## Inner Working

### Resource name evaluation

Most of the time, some information can be inferred about what a URL does based on it's contents.

For example, if a URL has a path with words like `login`, `logout` or `auth`, there is a high chance that it is related to a authentication flow. Same way that a path with words like `search` or `query` are likely search endpoints.

### Content evaluation

Some given aspects of a web page says a lot about how interesting a page is, even if they are not immediately visible. For example, having a `form` tag might indicate opportunities for SQL injection, and if a input of type file is present on this form, it might indicate opportunity for upload based attacks.

Script tags are a little more interesting, JavaScript can be used with different level of complexity and targeted user interaction. The existence of a `fetch` or `axios` call can show that the content is dynamic, or that there is a API behind this page. Pages that include these kinds of details should get more attention than a simple "about" page.

### Single resource example flowchart

![Main program flowchart](/doc/flowchart/img/main_program.svg)

## Ideas for the future

### Custom word list generation

Since the tool will be parsing data from (essentially) every page related to that service, it would be ideal to take that opportunity to generate a custom word list that can be later used for subdomain and resource path discovery. The idea is to use a NLP tokenizer, apply a PoS tagging, ignore parts of speech that would never be used in that context, and use this data to generate a unique targeted word list.

From my experience, common word lists that you can find online usually covers installed tools or common names given to products. But in real life, some companies name their products using strategic words or even the name they gave to that feature.

These can be misspelled or plays of other words, and these products are usually talked about and references on these pages.

### Integrated vertical correlation

One of my desires is to be able to use the [generated word list](#custom-word-list-generation) together with other known word lists to enumerate subdomain and resource names. This can be done by directly integrating with the [ffuf](https://github.com/ffuf/ffuf) backend.

## Technology

### Language and Runtime

The chosen technology for development is Golang, since most of the tools that are planned to be integrated with (ffuf, nuclei, ...) use Golang, making this integration faster and seamless.

### Tokenizing and interpreting HTML and JS

The core feature of this tool is to interpret HTML, JS, JSON, ... content and score this content based on pre-existing rules. The idea is to use the [native HTML parser](https://pkg.go.dev/golang.org/x/net@v0.40.0/html) implemented by Golang's networking module, and [goja's](https://github.com/dop251/goja/tree/master) JS runtime internal tokenizer to interpret the different content types (don't even need to mention JSON).

## TODO's

- [ ] Add metadata output