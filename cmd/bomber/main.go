package main

import (
	"log"
	"os"

	"github.com/qerdcv/bomber"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Description: "Bomber to flood ip addresses via packets",
		Commands: []*cli.Command{
			{
				Description: "flood server via ping packages",
				Name:        "ping",
				Action:      bomber.Ping,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    bomber.FlagWorkers,
						Aliases: []string{"w"},
						Value:   1000,
					},
					&cli.StringFlag{
						Name:     bomber.FlagAddress,
						Aliases:  []string{"a"},
						Required: true,
					},
					&cli.StringFlag{
						Name:    bomber.FlagPort,
						Aliases: []string{"p"},
						Value:   "80",
					},
					&cli.StringFlag{
						Name:    bomber.FlagProtocol,
						Aliases: []string{"pt"},
						Value:   "udp4",
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
