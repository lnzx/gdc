package commands

import (
	"fmt"
	"github.com/lnzx/gdc/internal/admin"
	"github.com/urfave/cli/v2"
)

var Group []*cli.Command

func init() {
	Group = []*cli.Command{
		{
			Name:  "group",
			Usage: "group manager",
			Subcommands: []*cli.Command{
				{
					Name:  "user",
					Usage: "user manager",
					Subcommands: []*cli.Command{
						{
							Name:  "add",
							Usage: "add user",
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:    "file",
									Aliases: []string{"f"},
									Usage:   "user emails file",
								},
								&cli.StringFlag{
									Name:    "user",
									Aliases: []string{"u"},
									Usage:   "user email",
								},
							},
							Action: func(c *cli.Context) error {
								if c.NArg() != 1 {
									return fmt.Errorf("please enter group email")
								}
								user := c.String("user")
								filepath := c.String("file")
								if user == "" && filepath == "" {
									return fmt.Errorf("please enter user or user emails file")
								}
								group := c.Args().Get(0)
								admin.AddGroupMember(group, user, filepath)
								return nil
							},
						},
					},
				},
			},
		},
	}
}
