package main

import "typinganuki.github.com/ppa-manager/ppa"

func main() {
	list := ppa.Lister{}
	ppas := list.List()
	for _, p := range ppas {
		p.Print()
	}
}
