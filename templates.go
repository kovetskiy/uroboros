package main

import (
	"text/template"
)

var (
	TemplateBadgeBuildPassing = template.Must(template.New("").Parse(
		"# [![uroboros: build passing](" +
			"{{ .basic_url }}/badges/build-passing.svg" +
			")]({{ .basic_url }}/task/{{ .taskID }})",
	))

	TemplateBadgeBuildFailure = template.Must(template.New("").Parse(
		"# [![uroboros: build failure](" +
			"{{ .basic_url }}/badges/build-failure.svg" +
			")]({{ .basic_url }}/task/{{ .taskID }})",
	))
)

var (
	TemplateCommentBuildPassing = template.Must(template.New("").Parse(
		"# [![uroboros: build passing](" +
			"{{ .basic_url }}/badges/build-passing.svg" +
			")]({{ .basic_url }}/task/{{ .taskID }})" +
			"```{{ .logs }}```",
	))

	TemplateCommentBuildFailure = template.Must(template.New("").Parse(
		"# [![uroboros: build failure](" +
			"{{ .basic_url }}/badges/build-failure.svg" +
			")]({{ .basic_url }}/task/{{ .taskID }})" +
			"```{{ .errors }}```",
	))
)
