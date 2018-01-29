package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// list all releases for nerd
	releases, _, err := client.Repositories.ListReleases(ctx, "nerdalize", "nerd", nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, release := range releases {
		fmt.Printf("Downloads for release [%s]:\n", *release.Name)
		for _, asset := range release.Assets {
			fmt.Printf("\t%s: %d times\n", *asset.Name, *asset.DownloadCount)
		}
	}

}
