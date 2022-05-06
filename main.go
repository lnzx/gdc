package main

import (
	"fmt"
	"github.com/lnzx/gdc/internal/admin"
	"github.com/lnzx/gdc/internal/drive"
	"github.com/lnzx/gdc/internal/oauth"
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
			ts := oauth.InitTokenSource(sa, subject)
			if ts != nil {
				drive.InitService(ts)
				admin.InitService(ts)
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
		Commands: []*cli.Command{
			{
				Name:  "mb",
				Usage: "Create a drive",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return fmt.Errorf("enter a drive name")
					}
					group := c.String("group")
					user := c.String("user")
					drive.CreateDrive(c.Args().Slice(), c.Int("count"), group, user)
					return nil
				},
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:    "count",
						Aliases: []string{"c"},
						Value:   1,
					},
					&cli.StringFlag{
						Name:    "group",
						Aliases: []string{"g"},
						Usage:   "Share driver group",
					},
					&cli.StringFlag{
						Name:    "user",
						Aliases: []string{"u"},
						Usage:   "Share driver user",
					},
				},
			},
			{
				Name:  "rb",
				Usage: "Delete an empty drive",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return fmt.Errorf("enter a drive id")
					}
					force := c.Bool("force")
					drive.DeleteDrive(c.Args().Slice(), force)
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Deletes all objects in the drive including the drive itself",
					},
				},
			},
			{
				Name:  "ls",
				Usage: "List drives or drive's files",
				Action: func(c *cli.Context) error {
					if c.NArg() == 1 {
						drive.List(c.Args().Get(0))
					} else {
						drive.List("")
					}
					return nil
				},
			},
			{
				Name:  "cat",
				Usage: "Concatenate object content to stdout",
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						return fmt.Errorf("enter a file id")
					}
					ranges := c.String("range")
					quiet := c.Bool("quiet")
					count := c.Int("count")
					randx := c.Int64("rand")
					drive.Cat(c.Args().Get(0), ranges, count, quiet, randx)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "range",
						Aliases: []string{"r"},
						Usage: `Output byte range of the object. Ranges are can be of these forms:
								start-end (e.g., -r 256-5939)
								start-    (e.g., -r 256-)
								-numbytes (e.g., -r -5)`,
					},
					&cli.BoolFlag{
						Name:    "quiet",
						Aliases: []string{"q"},
						Usage:   "quiet (no output)",
					},
					&cli.IntFlag{
						Name:    "count",
						Aliases: []string{"c"},
						Usage:   "requests count",
						Value:   1,
					},
					&cli.IntFlag{
						Name:  "rand",
						Usage: "Randomly read specified bytes,cannot be larger than 128kib",
					},
				},
			},
			{
				Name:  "cp",
				Usage: "Copy files and objects",
				Action: func(c *cli.Context) error {
					if c.NArg() != 2 {
						return fmt.Errorf("parameter error: file,driveId")
					}
					if c.IsSet("remote") {
						drive.CopyRemote(c.Args().Get(0), c.Args().Get(1))
					} else {
						return drive.Copy(c.Args().Get(0), c.Args().Get(1))
					}
					return nil
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "remote",
						Aliases: []string{"r"},
						Usage:   "Copy GD remote files",
					},
				},
			},
			{
				Name:  "mv",
				Usage: "Moves a local file to drive",
				Action: func(c *cli.Context) error {
					if c.NArg() != 2 {
						return fmt.Errorf("parameter error: file,driveId")
					}
					return drive.Move(c.Args().Get(0), c.Args().Get(1))
				},
			},
			{
				Name:  "rm",
				Usage: "Remove objects",
				Action: func(c *cli.Context) error {
					drive.Remove(c.Args().Slice())
					return nil
				},
			},
			{
				Name:  "sync",
				Usage: "Synchronize content of two drives/directories",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "dir",
						Aliases:  []string{"d"},
						Value:    "/mnt/tmp",
						Usage:    "monitoring directory",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "suffix",
						Aliases: []string{"s"},
						Value:   ".plot",
						Usage:   "monitoring file suffix",
					},
					&cli.StringFlag{
						Name:    "replace",
						Aliases: []string{"r"},
						Value:   ".gz",
						Usage:   "replaced new suffix",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() != 1 {
						fmt.Println("Please input driveId")
						os.Exit(1)
					}
					dir := c.String("dir")
					suffix := c.String("suffix")
					replace := c.String("replace")
					driveId := c.Args().First()
					drive.Sync(dir, suffix, replace, driveId)
					return nil
				},
			},
			{
				Name:  "group",
				Usage: "group user add",
				Subcommands: []*cli.Command{
					{
						Name:  "adduser",
						Usage: "add user",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "file",
								Aliases: []string{"f"},
								Usage:   "User emails file",
							},
							&cli.StringFlag{
								Name:    "user",
								Aliases: []string{"u"},
								Usage:   "User email",
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
							admin.AddMember(group, user, filepath)
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
