package proton

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func FindProtons() ([]string, error) {
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

	log.Debug().Strs("Protons", protonVersions).Send()

	return protonVersions, nil
}

func FindProtonsVersions() ([]string, error) {
	var protonVersionsTag []string

	protonVersions, err := FindProtons()
	if err != nil {
		log.Error().Err(err).Send()
		return protonVersionsTag, err
	}

	for _, protonVersion := range protonVersions {
		file := filepath.Join(protonVersion, "version")
		data, _ := os.ReadFile(file)
		dataString := string(data)

		version := strings.Fields(dataString)[1]
		protonVersionsTag = append(protonVersionsTag, version)
	}

	return protonVersionsTag, nil
}

type ProtonRunner struct {
	Path string
	Wine string
}

func (p *ProtonRunner) SetPath(path string) {
	p.Path = path
	p.Wine = filepath.Join(path, "files/bin/wine")
}

func NewProtonRunnerFromForm() (ProtonRunner, error) {
	var newProtonRunner ProtonRunner
	choices, err := FindProtons()
	if err != nil {
		log.Error().Err(err).Send()
		return newProtonRunner, err
	}

	choicesDisplayed := lo.Map(choices, func(v string, _ int) string {
		return filepath.Base(v)
	})

	var protonChoice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Pick your proton runner").
				Options(
					huh.NewOptions(choicesDisplayed...)...,
				).
				Value(&protonChoice),
		))
	form.Run()

	log.Debug().Str("Picked", protonChoice).Send()
	newProtonRunner.SetPath(choices[slices.Index(choicesDisplayed, protonChoice)])
	return newProtonRunner, nil
}

func NewProtonRunnerFromPath(path string) (ProtonRunner, error) {
	var output ProtonRunner
	path, err := filepath.Abs(path)
	if err != nil {
		return output, err
	}

	if strings.Contains(strings.ToLower(path), "proton") {
		output.SetPath(path)
	}

	return output, nil
}
