package main

import (
	"github.com/codegangsta/cli"
	"os"
)

var (
	addr    string
	jbrowse string
)

func main() {

	app := cli.NewApp()
	app.Name = "chado-jb-connector"
	app.Usage = "serve Chado as JBrowse REST compatible API"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "db",
			Value: "postgres://postgres:postgres@localhost/postgres?sslmode=disable",
			Usage: "database address",
			EnvVar: "CHADOJB_DBSTRING",
		},
		cli.StringFlag{
			Name:  "listenAddr",
			Value: "0.0.0.0:5000",
			EnvVar: "CHADOJB_LISTENADDR",
		},
		cli.StringFlag{
			Name:  "sitePath",
			Value: "http://localhost:5000",
			Usage: "set externally accessible URL.",
			EnvVar: "CHADOJB_SITEPATH",
		},
		cli.StringFlag{
			Name:  "jbrowse",
			Value: "https://jbrowse.org/code/JBrowse-1.12.0/",
			Usage: "JBrowse deployment to display REST tracks on",
			EnvVar: "CHADOJB_JBROWSE",
		},
	}

	app.Action = func(c *cli.Context) {

		addr = c.String("sitePath")
		jbrowse = c.String("jbrowse")
		connect(
			c.String("db"),
			c.String("listenAddr"),
		)
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
