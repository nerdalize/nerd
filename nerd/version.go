package nerd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

const (
	//GHReleasesURL is the GitHub API URL that lists all releases.
	GHReleasesURL = "https://api.github.com/repos/nerdalize/nerd/releases"
	//BuiltFromSourceVersion is the version when building form source (no real version).
	BuiltFromSourceVersion = "built.from.src"
)

//GHRelease is a `release` object from the GitHub API.
type GHRelease struct {
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

//GHError is an error object from the GitHub API.
type GHError struct {
	Message string `json:"message"`
	URL     string `json:"documentation_url"`
}

//VersionMessage shows a message to the user if a new CLI version is available.
func VersionMessage(current string) {
	if current == BuiltFromSourceVersion {
		return
	}
	var releases []GHRelease
	e := new(GHError)
	s := sling.New().Get(GHReleasesURL)
	_, err := s.Receive(&releases, e)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "failed to access GitHub releases page"))
		return
	}
	if e.Message != "" {
		fmt.Fprintf(os.Stderr, "Recieved GitHub error message: %v (%v)\n", e.Message, e.URL)
		return
	}
	if len(releases) > 0 {
		latest := releases[0]
		latestVersion, err := ParseSemVer(strings.Replace(latest.Name, "v", "", 1))
		if err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "failed to parse latest semantic version"))
			return
		}
		currentVersion, err := ParseSemVer(current)
		if err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "failed to parse current semantic version"))
			return
		}
		if latestVersion.GreaterThan(currentVersion) {
			fmt.Fprintf(os.Stderr, "A new version (%v) of the nerd CLI is available. Your current version is %v. Please visit %v to get the latest version.\n", latestVersion.ToString(), currentVersion.ToString(), latest.HTMLURL)
		}
	}
}

//SemVer is a semantic version.
type SemVer struct {
	Major int
	Minor int
	Patch int
}

//ParseSemVer parses a semantic version from string.
func ParseSemVer(ver string) (*SemVer, error) {
	parts := strings.Split(ver, ".")
	if len(parts) != 3 {
		return nil, errors.Errorf("failed to parse semantic version '%v', because does not consist of 3 parts", ver)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse semantic version '%v', because the major version is not an integer", ver)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse semantic version '%v', because the minor version is not an integer", ver)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse semantic version '%v', because the patch version is not an integer", ver)
	}
	return &SemVer{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

//GreaterThan checks if `s` is a greater semantic version than `other`
func (s *SemVer) GreaterThan(other *SemVer) bool {
	if s.Major > other.Major {
		return true
	}
	if s.Major < other.Major {
		return false
	}
	if s.Minor > other.Minor {
		return true
	}
	if s.Minor < other.Minor {
		return false
	}
	if s.Patch > other.Patch {
		return true
	}
	return false
}

//ToString converts a SemVer to a string
func (s *SemVer) ToString() string {
	return fmt.Sprintf("%v.%v.%v", s.Major, s.Minor, s.Patch)
}
