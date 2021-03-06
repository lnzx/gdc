package commands

import (
	"fmt"
	"github.com/lnzx/gdc/internal/drive"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var Drive []*cli.Command

func init() {
	Drive = []*cli.Command{
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
				&cli.DurationFlag{
					Name:    "time",
					Aliases: []string{"t"},
					Usage:   "task time(m)",
					Value:   time.Minute,
				},
				&cli.StringFlag{
					Name:     "parentId",
					Aliases:  []string{"p"},
					Usage:    "head parent dir id",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					fmt.Println("Please input driveId")
					os.Exit(1)
				}
				dir := c.String("dir")
				driveId := c.Args().First()
				t := c.Duration("time")
				parentId := c.String("parentId")
				drive.Sync(dir, driveId, t, parentId)
				return nil
			},
		},
		{
			Name:  "drive",
			Usage: "drive manager",
			Subcommands: []*cli.Command{
				{
					Name:  "addgroup",
					Usage: "Share to a group",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "group",
							Aliases: []string{"g"},
							Usage:   "group email address",
						},
					},
					Action: func(c *cli.Context) error {
						group := c.String("group")
						if group == "" {
							fmt.Println("please input a group email address")
							return nil
						}
						driveId := c.Args().First()
						if driveId == "" {
							fmt.Println("please input a drive id arg")
							return nil
						}
						drive.AddDriveGroup(driveId, group)
						return nil
					},
				},
				{
					Name:  "adduser",
					Usage: "Share to a user",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "user",
							Aliases: []string{"u"},
							Usage:   "user email address",
						},
					},
					Action: func(c *cli.Context) error {
						user := c.String("user")
						if user == "" {
							fmt.Println("please input a user email address")
							return nil
						}
						driveId := c.Args().First()
						if driveId == "" {
							fmt.Println("please input a drive id arg")
							return nil
						}
						drive.AddDriveUser(driveId, user)
						return nil
					},
				},
			},
		},
	}
}
