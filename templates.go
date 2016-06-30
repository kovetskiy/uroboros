package main

import (
	"text/template"
)

var (
	TemplateMarkdownBuildPassing = template.Must(
		template.New(``).Parse(
			`# [![build passing](` +
				`https://img.shields.io/badge/build-passing-brightgreen.svg` +
				`)](http://uroboro.s/task/{{ .taskID }})

` + "```" + `
{{ .logs }}
` + "```" + `

`))

	TemplateMarkdownBuildFailure = template.Must(
		template.New(``).Parse(
			`# [![build failure](` +
				`https://img.shields.io/badge/build-failure-red.svg` +
				`)](http://uroboro.s/task/{{ .taskID }})

` + "```" + `
{{ .errors }}
` + "```" + `

`))
)
