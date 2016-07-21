package main

const homeTemplate = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>
	<body>
		<h1>Genomes in Chado</h1>
		<ul>
		{{range .Items}}
			<li>
				<a href="{{ .JBrowseURL }}?data={{ $.FakeDirURL }}/{{ . }}">
				{{ . }}
				</a>
			</li>
		{{end}}
		</ul>
	</body>
</html>
`
