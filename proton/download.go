package proton

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gumu/config"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/schollz/progressbar/v3"
	"golift.io/xtractr"
)

func DownloadNewProton(c context.Context, conf *config.Config) error {
	repos := conf.ConfigProton.Repos
	confirm := false
	reposLinks := make(map[string][]apiRelease, 10)
	protonVersionsInstalled, err := FindProtonsVersions()
	log.Debug().Strs("installed", protonVersionsInstalled).Send()

	if err != nil {
		return err
	}

	for _, repo := range repos {
		owner, repoName, _ := strings.Cut(repo, "/")
		resp, err := listReleases(c, owner, repoName, "")
		if err != nil {
			return err
		}
		reposLinks[repo] = resp
	}

	var repoSelected, versionSelected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(huh.NewOptions(repos...)...).
				Title("Repo").
				Value(&repoSelected),

			huh.NewSelect[string]().
				TitleFunc(func() string { return repoSelected }, &repoSelected).
				OptionsFunc(func() []huh.Option[string] {
					outputStrings := lo.Map(reposLinks[repoSelected], func(v apiRelease, index int) string {
						if slices.Contains(protonVersionsInstalled, v.TagName) {
							return fmt.Sprintf("%s [Installed]", v.TagName)
						} else {
							return v.TagName
						}
					})
					return huh.NewOptions(outputStrings...)
				}, &repoSelected).
				Value(&versionSelected),

			huh.NewConfirm().Title("Confirm?").Value(&confirm),
		),
	)

	form.Run()

	if !confirm {
		return nil
	}

	urlAsset, found := lo.Find(reposLinks[repoSelected], func(v apiRelease) bool {
		return v.TagName == strings.Fields(versionSelected)[0]
	})
	url, found := lo.Find(urlAsset.Assets, func(v asset) bool {
		if strings.Contains(v.Name, "tar.xz") && strings.Contains(v.Name, "x86_64") {
			return true
		}
		if strings.Contains(v.Name, "tar.gz") {
			return true
		}
		if strings.Contains(v.Name, "tar.xz") {
			return true
		}

		return false
	})

	if found == false {
		err := errors.New("Couldn't find link to download this version")
		return err
	}

	home, _ := os.LookupEnv("HOME")
	pathOutput := filepath.Join(home, ".local/share/Steam/compatibilitytools.d")
	pathArchive := filepath.Join(pathOutput, "tmp")
	os.MkdirAll(pathArchive, 0o755)
	defer os.RemoveAll(pathArchive)
	err = downloadAsset(c, url.URL, filepath.Join(pathArchive, url.Name), "")
	log.Debug().Str("selected", versionSelected).Send()
	if err != nil {
		return err
	}

	finalFolderName := strings.TrimSuffix(url.Name, filepath.Ext(url.Name))
	finalFolderName = strings.TrimSuffix(finalFolderName, filepath.Ext(finalFolderName))
	pathOutput = filepath.Join(pathOutput, finalFolderName)
	err = spinner.New().ActionWithErr(func(ctx context.Context) error {
		return extractProton(filepath.Join(pathArchive, url.Name), pathOutput)
	}).Title("Extracting...").Run()

	return err
}

func extractProton(archive string, path string) error {
	log.Debug().Str("archive", archive).Str("output path", path).Send()
	response := make(chan *xtractr.Response)
	archiveQueue := &xtractr.Xtract{
		Name:      "archive",
		CBChannel: response,
		Filter: xtractr.Filter{
			Path: archive,
		},
	}
	q := xtractr.NewQueue(&xtractr.Config{
		Parallel: 1,
		FileMode: 0o644,
		DirMode:  0o755,
	})
	defer q.Stop()

	q.Extract(archiveQueue)
	resp := <-response
	log.Debug().Int("archives", resp.Queued).Msg("Started decompressing")
	resp = <-response
	log.Debug().Str("output", resp.Output).Strs("new files", resp.NewFiles).Msg("Finished")

	if resp.Error != nil {
		return resp.Error
	}
	if len(resp.NewFiles) < 1 {
		return errors.New("No files extracted")
	}

	newfolder := resp.NewFiles[0]
	err := os.Rename(newfolder, path)
	return err
}

type asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func (a asset) String() string {
	return fmt.Sprintf("%v, %v", a.Name, a.URL)
}

type apiRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

func (a apiRelease) String() string {
	var output strings.Builder

	output.WriteString(a.TagName + "   ")
	for _, v := range a.Assets {
		output.WriteString(v.String() + " ")
	}

	return output.String()
}

func listReleases(ctx context.Context, owner, repo, token string) ([]apiRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gumu")
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	var resp *http.Response
	downloadProton := func(c context.Context) error {
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		return nil
	}
	spinner.New().Title("Fetching...").ActionWithErr(downloadProton).Run()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github error: %s", body)
	}

	var releases []apiRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	return releases, err
}

func downloadAsset(ctx context.Context, url, filepath, token string) error {
	log.Debug().Str("URL", url).Send()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "gumu")
	req.Header.Set("Accept", "application/octet-stream")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: %s", body)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	foreground := lipgloss.Color("#32348F")
	style := lipgloss.NewStyle().Foreground(lipgloss.Lighten(foreground, 0.05))
	saucer := style.Foreground(lipgloss.Lighten(foreground, 0.10)).Render("█")

	bar := progressbar.NewOptions(
		int(resp.ContentLength),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription(style.Render("Downloading...")),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        saucer,
			SaucerPadding: " ",
			BarStart:      "|",
			BarEnd:        "|",
		}),
	)

	_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	fmt.Println() // newline after progress finishes
	return err
}
