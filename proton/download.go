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

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

func DownloadNewProton(c context.Context) error {
	repos := []string{"cachyos/proton-cachyos", "GloriousEggroll/proton-ge-custom"}
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
			log.Error().Err(err).Send()
			continue
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
		),
	)

	form.Run()

	urlAsset, found := lo.Find(reposLinks[repoSelected], func(v apiRelease) bool {
		return v.TagName == versionSelected
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

	err = spinner.New().ActionWithErr(func(c context.Context) error {
		return downloadAsset(c, url.URL, filepath.Join("./", url.Name), "")
	}).
		Title("Downloading...").
		Run()
	log.Debug().Str("selected", versionSelected).Send()
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

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
