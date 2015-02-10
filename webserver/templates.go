package webserver

import (
	"html/template"
)

func parseTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template, 2)

	base := template.Must(template.New("base").Parse(baseSrc))

	index := template.Must(base.Clone())
	result["index"] = template.Must(index.Parse(indexSrc))

	schedule := template.Must(base.Clone())
	result["schedule"] = template.Must(schedule.Parse(scheduleSrc))

	return result
}

const baseSrc = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <title>Heating controller</title>
    <meta name="viewport" content="width=device-width,initial-scale=1.0" />
  </head>
  <body>
    {{ template "content" . }}
  </body>
</html>
`
