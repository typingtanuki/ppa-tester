package ppa

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Ppa struct {
	URL       string
	Version   string
	Sub       string
	PpaLink   string
	Enabled   bool
	Src       bool
	Outdated  bool
	Updatable bool
	Flags     []string
	Origin    string
}

func (p *Ppa) Print() {
	if p.Src {
		return
	}

	if p.Enabled && !p.Outdated {
		return
	}

	if p.Enabled && p.Outdated && !p.Updatable {
		return
	}

	fmt.Println("Ppa{")
	fmt.Println("\tURL: " + p.URL)
	if len(p.PpaLink) > 0 {
		fmt.Println("\tPpaLink: " + p.PpaLink)
	}
	fmt.Println("\tVersion: " + p.Version)
	if len(p.Sub) > 0 {
		fmt.Println("\tSub: " + p.Sub)
	}
	fmt.Println("\tEnabled: " + strconv.FormatBool(p.Enabled))
	if p.Src {
		fmt.Println("\tSrc: " + strconv.FormatBool(p.Src))
	}
	if p.Outdated {
		fmt.Println("\tOutdated: " + strconv.FormatBool(p.Outdated))
		fmt.Println("\tUpdatable: " + strconv.FormatBool(p.Updatable))
	}
	fmt.Println("\tFlags: " + strings.Join(p.Flags, ", "))
	fmt.Println("\tOrigin: " + p.Origin)
	fmt.Println("}")
}

func (p *Ppa) Consolidate(codename string) {
	if strings.Contains(p.URL, "ppa.launchpad") {
		urlParts := strings.Split(p.URL, "/")
		p.PpaLink = "ppa:" + urlParts[3] + "/" + urlParts[4]
	}

	if !p.Enabled && p.Outdated {
		checkUpdate(p, codename)
	}
}

func checkUpdate(p *Ppa, codename string) {
	updateURL := p.URL + "dists/" + codename + "/InRelease"

	_, err := http.Get(updateURL)

	p.Updatable = err == nil
}
