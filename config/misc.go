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
	p := termenv.ColorProfile()
	bnr, _ := squish.UnpackStr(Banner)
	gr := termenv.String(binInfo["vcs.revision"][:7]).Foreground(termenv.ANSIBrightGreen).String()
	born := termenv.String(binInfo["vcs.time"]).Foreground(p.Color("#1e9575")).String()
	out := strings.Replace(bnr, "$gitrev$", gr, 1)
	out = strings.Replace(out, "$date$", born, 1)
	cout := termenv.String(out)
	print(cout.Foreground(p.Color("#948DB8")).String())

}
