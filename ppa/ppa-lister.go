package ppa

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Lister struct{}

func (lister *Lister) List() (ppas []*Ppa) {
	out := []*Ppa{}

	aptSourceDir := "/etc/apt/sources.list.d"
	aptSourceFile := "/etc/apt/sources.list"
	releaseFile := "/etc/lsb-release"

	currentCodeName := findCodeName(releaseFile)

	files, err := ioutil.ReadDir(aptSourceDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		name := file.Name()
		out = append(out, readPpaFile(filepath.Join(aptSourceDir, name), currentCodeName)...)
	}

	out = append(out, readPpaFile(aptSourceFile, currentCodeName)...)

	return out
}

func findCodeName(file string) string {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DISTRIB_CODENAME") {
			return strings.TrimSpace(strings.SplitN(line, "=", 2)[1])
		}
	}
	panic("No codename")
}

func readPpaFile(file string, codename string) []*Ppa {
	var out []*Ppa

	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "# deb") {
			continue
		}
		out = append(out, buildPpa(file, line, codename))
	}

	return buildDB(out, codename)
}

func buildPpa(file string, line string, codename string) *Ppa {
	enabled := true
	if strings.HasPrefix(line, "#") {
		enabled = false
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[1])
	}

	parts := strings.Split(line, " ")
	if len(parts) < 4 {
		panic("Wrong number of parts: " + line)
	}

	cursor := 0

	repoType := parts[cursor]
	cursor++

	repoURL := parts[cursor]
	cursor++

	if strings.Contains(repoURL, "[arch=") {
		repoURL = parts[cursor]
		cursor++
	}

	version := parts[cursor]
	cursor++

	sub := ""
	if strings.Contains(version, "-") {
		sub = strings.SplitN(version, "-", 2)[1]
		version = strings.SplitN(version, "-", 2)[0]
	}

	return &Ppa{
		URL:      repoURL,
		Version:  version,
		Sub:      sub,
		Enabled:  enabled,
		Outdated: version != codename && version != "stable",
		Src:      strings.Contains(repoType, "-src"),
		Flags:    parts[cursor:],
		Origin:   file,
	}
}

var (
	resolveCache []string
	mu           = sync.Mutex{}
)

func buildDB(ppas []*Ppa, codename string) []*Ppa {
	m := make(map[string]*Ppa)

	for _, p := range ppas {
		if p.Src {
			continue
		}

		saved := m[p.URL]

		if saved == nil {
			m[p.URL] = p

			continue
		}

		if saved.Enabled && !p.Enabled {
			continue
		}

		if !saved.Enabled && p.Enabled {
			m[p.URL] = p

			continue
		}

		if isBetterVersion(saved.Version, p.Version) {
			m[p.URL] = p

			continue
		}
	}

	out := []*Ppa{}

	for _, found := range m {
		found.Consolidate(codename)
		out = append(out, found)
	}

	return out
}

func isBetterVersion(version1 string, version2 string) bool {
	return version1 < version2
}

func Reset() {
	mu.Lock()
	defer mu.Unlock()
	resolveCache = resolveCache[:0]
}
