package main

import (
	"fmt"
	"github.com/lnzx/gdc/internal"
	"github.com/lnzx/gdc/internal/admin"
	"github.com/lnzx/gdc/internal/commands"
	"github.com/lnzx/gdc/internal/drive"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:    "gdc",
		Usage:   "google drive cli",
		Version: "0.0.1",
		Before: func(c *cli.Context) error {
			sa := c.String("sa")
			subject := c.String("subject")
			ts := internal.InitTokenSource(sa, subject)
			if ts != nil {
				drive.InitService(ts)
				if subject != "" {
					admin.InitService(ts)
				}
			}

			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "sa",
				Value: "sa.json",
			},
			&cli.StringFlag{
				Name:    "subject",
				Aliases: []string{"s"},
				Usage:   "user email to impersonate",
			},
		},
		Commands: append(commands.Drive, commands.Group...),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
