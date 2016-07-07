package main

import (
	"text/template"
)

var (
	TemplateBadge = template.Must(template.New("").Parse(
		"# [![uroboros: build status](" +
			"{{ .basic_url }}/badge/{{ .slug }}" +
			")]({{ .basic_url }}/status/{{ .slug }})",
	))
)

var (
	TemplateCommentBuildPassing = template.Must(template.New("").Parse(
		"# [![uroboros: build passing](" +
			"{{ .basic_url }}" + pathStaticBadgeBuildPassing +
			")]({{ .basic_url }}/status/{{ .id }})" +
			"\n```{{ .logs }}```",
	))

	TemplateCommentBuildFailure = template.Must(template.New("").Parse(
		"# [![uroboros: build failure](" +
			"{{ .basic_url }}" + pathStaticBadgeBuildFailure +
			")]({{ .basic_url }}/status/{{ .id }})" +
			"\n```{{ .errors }}```",
	))
)
