package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
	kc "github.com/jotaen/kong-completion"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var styleOutputBox = lipgloss.NewStyle().
	Border(lipgloss.ThickBorder(), false, false, false, true).
	Padding(0, 1).
	BorderForeground(lipgloss.Color("#9573ff"))

func findProtons() ([]string, error) {
	var protonVersions []string
	home, _ := os.LookupEnv("HOME")
	pathLower := filepath.Join(home, ".local/share/Steam/compatibilitytools.d", "*proton*")
	pathHigher := filepath.Join(home, ".local/share/Steam/compatibilitytools.d", "*Proton*")
	results, err := filepath.Glob(pathLower)
	if err != nil {
		return protonVersions, err
	}
	protonVersions = append(protonVersions, results...)

	results, err = filepath.Glob(pathHigher)
	if err != nil {
		return protonVersions, err
	}
	protonVersions = append(protonVersions, results...)
	slices.Sort(protonVersions)

	return protonVersions, nil
}

var CLI struct {
	Prefix     PrefixCmd     `cmd:"prefix" help:"Manage your prefixs"`
	Proton     ProtonCmd     `cmd:"proton" help:"Manage your proton installations"`
	Completion kc.Completion `cmd:"" hidden:""`
}

type ProtonCmd struct {
	List ProtonListCmd `cmd:"list" help:"List your proton installations" completion-predictor:"protonList"`
}

type ProtonListCmd struct{}

func (p *ProtonListCmd) Run() error {
	protonVersions, err := findProtons()
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

type PrefixCmd struct {
	prefixList PrefixListCmd
}

type PrefixListCmd struct{}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	parser, _ := kong.New(&CLI, kong.UsageOnError())
	kc.Register(parser)
	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)
	err = ctx.Run(&CLI)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
