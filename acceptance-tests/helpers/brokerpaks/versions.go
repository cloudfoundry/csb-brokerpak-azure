package brokerpaks

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/blang/semver/v4"
)

// LatestVersion will determine the latest released version of the brokerpak
// according to semantic versioning
func LatestVersion() string {
	versions := Versions()
	latest := versions[len(versions)-1].String()
	fmt.Printf("Latest brokerpak version: %s\n", latest)
	return latest
}

// Versions will get all the released Versions
func Versions() []semver.Version {
	body := newClient().get(fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=100", brokerpak), "application/json") // max per page
	defer body.Close()
	data := must(io.ReadAll(body))

	var receiver []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(data, &receiver); err != nil {
		panic(err)
	}

	var versions []semver.Version
	for _, r := range receiver {
		v := must(semver.ParseTolerant(r.TagName))
		if len(v.Pre) == 0 { // skip pre-release
			versions = append(versions, v)
		}
	}
	sort.SliceStable(versions, func(i, j int) bool { return versions[i].LT(versions[j]) })

	return versions
}
