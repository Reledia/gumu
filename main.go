package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gumu/prefix"
	"gumu/proton"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	kc "github.com/jotaen/kong-completion"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var styleOutputBox = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 1).
	BorderForeground(lipgloss.Color("#9573ff"))

var CLI struct {
	Debug      bool          `short:"D" help:"Show debug log"`
	Completion kc.Completion `cmd:"" hidden:""`
	Prefix     PrefixCmd     `cmd:"prefix" help:"Manage your prefixs"`
	Proton     ProtonCmd     `cmd:"proton" help:"Manage your proton installations"`
}

type ProtonCmd struct {
	List ProtonListCmd   `cmd:"list" help:"List your proton installations" completion-predictor:"protonList"`
	Pick ProtonPickerCmd `cmd:"pick" help:"Temp"`
}

type ProtonListCmd struct{}

func (p *ProtonListCmd) Run() error {
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

type ProtonPickerCmd struct{}

func (p *ProtonPickerCmd) Run() error {
	protonRunner, err := proton.NewProtonRunner()
	log.Debug().Any("protonRunner", protonRunner.Path).Send()
	return err
}

type PrefixCmd struct {
	List   PrefixListCmd   `cmd:""`
	Create PrefixCreateCmd `cmd:""`
}

type PrefixListCmd struct{}

type PrefixCreateCmd struct {
	Path   string `help:"Indicate where to create the new prefix"`
	Proton string `help:"Indicate which proton to use for the new prefix"`
	Name   string `help:"Name of the new prefix"`
}

func (p *PrefixCreateCmd) Run() error {
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

	if CLI.Prefix.Create.Path == "" {
		formPath.Run()
	} else {
		path, _ = filepath.Abs(CLI.Prefix.Create.Path)
	}

	if CLI.Prefix.Create.Name == "" {
		formName.Run()
	} else {
		name = strings.TrimSpace(CLI.Prefix.Create.Name)
	}

	if CLI.Prefix.Create.Proton == "" {
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

	parser, _ := kong.New(&CLI, kong.UsageOnError())
	kc.Register(parser)
	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	if CLI.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	err = ctx.Run(&CLI)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
