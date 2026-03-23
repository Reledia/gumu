package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gumu/prefix"
	"gumu/proton"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
)

var styleOutputBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 1).
	BorderForeground(lipgloss.Color("#9573ff"))

var cmd = &cli.Command{
	EnableShellCompletion:  true,
	UseShortOptionHandling: true,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Aliases: []string{"d"},
			Value:   false,
			Usage:   "Print debug logs",
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

	if cmd.Bool("debug") {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
