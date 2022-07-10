package config

import (
	"strings"

	"git.tcp.direct/kayos/common/squish"
	"github.com/muesli/termenv"
)

const Banner = "H4sIAAAAAAACA+OSjrY0szYxyH00rUE62hjK7AAjoIBBLggrKEBUmWGoIkYJkERWZWwOkgKxSNVJhBMMcrngNqB6B0kYyS2kKHg0rQXNNmxewdDRQZKFFHgAaBtBqzBCHaSiBaG4hQjHwoWJdAjQSwoq6ZklRallKgoKKimJJakqXABftjxAeAIAAA=="

func PrintBanner() {
	if noColorForce {
		println("tcp.ac\n")
		return
	}
	gitr := ""
	brn := ""
	if gitrev, ok := binInfo["vcs.revision"]; ok {
		gitr = gitrev[:7]
	}
	if vt, ok := binInfo["vcs.time"]; ok {
		brn = vt
	}

	p := termenv.ColorProfile()
	bnr, _ := squish.UnpackStr(Banner)
	gr := termenv.String(gitr).Foreground(termenv.ANSIBrightGreen).String()
	born := termenv.String(brn).Foreground(p.Color("#1e9575")).String()
	out := strings.Replace(bnr, "$gitrev$", gr, 1)
	out = strings.Replace(out, "$date$", born, 1)
	cout := termenv.String(out)
	print(cout.Foreground(p.Color("#948DB8")).String())

}
