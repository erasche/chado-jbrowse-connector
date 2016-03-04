package main

import (
	"github.com/codegangsta/cli"
	"os"
)

var addr string

func main() {

	app := cli.NewApp()
	app.Name = "chado-jb-connector"
	app.Usage = "serve Chado as JBrowse REST compatible API"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "db",
			Value: "postgres://postgres:postgres@localhost/postgres?sslmode=disable",
			Usage: "database address",
		},
		cli.StringFlag{
			Name:  "listenAddr",
			Value: "0.0.0.0:5000",
		},
		cli.StringFlag{
			Name:  "sitePath",
			Value: "http://shed.hx42.org:5000",
			Usage: "set externally accessible URL. I'll fix this eventually.",
		},
	}

	app.Action = func(c *cli.Context) {

		addr = c.String("sitePath")
		Db(
			c.String("db"),
			c.String("listenAddr"),
		)
	}
	app.Run(os.Args)
}
