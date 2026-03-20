package prefix

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gumu/proton"

	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
)

type CreatePrefixOptions struct {
	Path   string
	Name   string
	Proton proton.ProtonRunner
}

func CreatePrefix(options *CreatePrefixOptions) error {
	var err error

	path := filepath.Join(options.Path, options.Name)
	if (options.Proton == proton.ProtonRunner{}) {
		options.Proton, err = proton.NewProtonRunner()
	}
	if err != nil {
		return err
	}

	wineCmd := exec.Command("umu-run", "reg", "/?")

	envs := os.Environ()
	envs = append(envs, fmt.Sprintf("WINEPREFIX=%v", path))
	envs = append(envs, fmt.Sprintf("PROTONPATH=%v", options.Proton.Path))
	wineCmd.Env = append(wineCmd.Env, envs...)
	log.Debug().Strs("ENV", wineCmd.Env).Send()

	run := func() {
		err = wineCmd.Run()
		if err != nil {
			log.Error().Err(err).Send()
		}
	}
	err = spinner.New().
		Title("Creating prefix...").
		Action(run).
		Run()
	log.Debug().Str("command", wineCmd.String()).Send()
	return err
}
