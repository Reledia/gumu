package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gumu/config"
	"gumu/prefix"
	"gumu/proton"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

var (
	styleOutputBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			BorderForeground(lipgloss.Color("#9573ff"))

	conf, _ = config.LoadConf()
)

var cmd = &cli.Command{
	EnableShellCompletion:  true,
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "Print debug logs",
			Action: func(ctx context.Context, c *cli.Command, b bool) error {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
				return nil
			},
		},
	},
	Commands: []*cli.Command{
		{
			Name:  "prefix",
			Usage: "Manage your prefixs",
			Commands: []*cli.Command{
				{
					Name:   "create",
					Usage:  "Create a new prefix",
					Action: PrefixCreateCmd,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  "proton",
							Value: "",
							Usage: "Indicate proton to use for the new prefix",
						},
						&cli.StringFlag{
							Name:  "path",
							Value: "",
							Usage: "Indicate where to create the new prefix",
						},
						&cli.StringFlag{
							Name:  "name",
							Value: "",
							Usage: "Name of the new prefix",
						},
					},
				},
			},
		},
		{
			Name:  "proton",
			Usage: "Manage your proton installations",
			Commands: []*cli.Command{
				{
					Name:   "list",
					Usage:  "List your proton installations",
					Action: ProtonListCmd,
				},
				{
					Name:   "download",
					Usage:  "Download a new proton runner",
					Action: ProtonDownloadCmd,
				},
			},
		},
		{
			Name:        "test",
			Description: "debug commands",
			Commands: []*cli.Command{
				{
					Name: "configWrite",
					Action: func(ctx context.Context, c *cli.Command) error {
						conf, err := config.LoadConf()
						log.Debug().Strs("conf", conf.ConfigProton.Repos).Send()
						return err
					},
				},
			},
		},
	},
}

func ProtonListCmd(c context.Context, CLI *cli.Command) error {
	protonVersions, err := proton.FindProtons()
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}
	protonVersions = lo.Map(protonVersions, func(item string, _ int) string {
		return filepath.Base(item)
	})
	output := strings.Join(protonVersions, "\n")
	output = styleOutputBox.Render(output)
	fmt.Println(output)
	return nil
}

func ProtonDownloadCmd(c context.Context, CLI *cli.Command) error {
	return proton.DownloadNewProton(c, conf)
}

func PrefixCreateCmd(c context.Context, CLI *cli.Command) error {
	var path, name, base string
	var protonRunner proton.ProtonRunner
	var options prefix.CreatePrefixOptions
	var err error
	base, _ = filepath.Abs("./..")

	formPath := huh.NewForm(
		huh.NewGroup(
			huh.NewFilePicker().
				Title("Prefix location").
				CurrentDirectory(base).
				Picking(true).
				DirAllowed(true).
				FileAllowed(false).
				Height(20).
				Value(&path),
		))
	formName := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Prefix name").
				Value(&name),
		))

	if CLI.String("path") == "" {
		formPath.Run()
	} else {
		path, _ = filepath.Abs(CLI.String("path"))
	}

	if CLI.String("name") == "" {
		formName.Run()
	} else {
		name = strings.TrimSpace(CLI.String("name"))
	}

	if CLI.String("proton") == "" {
		protonRunner, err = proton.NewProtonRunner()
		if err != nil {
			return err
		}
	}

	options.Proton = protonRunner
	options.Path = path
	options.Name = name
	return prefix.CreatePrefix(&options)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
