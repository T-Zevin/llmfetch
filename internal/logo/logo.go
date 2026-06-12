package logo

import (
	"fmt"
	"sort"
	"strings"
)

type Logo struct {
	Name  string
	Lines []string
}

var catalog = map[string]Logo{
	"almalinux": {
		Name: "AlmaLinux",
		Lines: []string{
			"      .:cccc:.",
			"   .:cdddddddc:.",
			"  :ddd:'  ':ddd:",
			" cddd:      :dddc",
			" cddd:      :dddc",
			"  :ddd:.  .:ddd:",
			"   ':cddddddc:'",
			"      '::::'",
		},
	},
	"alpine": {
		Name: "Alpine",
		Lines: []string{
			"       /\\",
			"      /  \\",
			"     / /\\ \\",
			"    / /  \\ \\",
			"   / / /\\ \\ \\",
			"  /_/ /  \\_\\_\\",
			"     /____\\",
		},
	},
	"arch": {
		Name: "Arch",
		Lines: []string{
			"        /\\",
			"       /  \\",
			"      / /\\ \\",
			"     / ____ \\",
			"    /_/    \\_\\",
			"   /   __    \\",
			"  /___/  \\____\\",
		},
	},
	"centos": {
		Name: "CentOS",
		Lines: []string{
			"    +--------+",
			"    |\\      /|",
			"    | \\____/ |",
			"    | /    \\ |",
			"    |/______\\|",
			"    +--------+",
		},
	},
	"debian": {
		Name: "Debian",
		Lines: []string{
			"      _____",
			"   .-'     '-.",
			"  /  .---.    \\",
			" |  (     )    |",
			"  \\  '---'   /",
			"   '-.____.-'",
			"       '--",
		},
	},
	"elementary": {
		Name: "elementary",
		Lines: []string{
			"     _______",
			"   .'       '.",
			"  /  .-----.  \\",
			" |  /       \\  |",
			" |  \\______.  |",
			"  \\         /",
			"   '.___..'",
		},
	},
	"endeavouros": {
		Name: "EndeavourOS",
		Lines: []string{
			"        /\\",
			"       /  \\",
			"      / /\\ \\",
			"     / /  \\ \\",
			"    / /____\\ \\",
			"   /_/      \\_\\",
			"      Endeavour",
		},
	},
	"fedora": {
		Name: "Fedora",
		Lines: []string{
			"      .====.",
			"    .'      '.",
			"   /  .--.    \\",
			"  |  |    |    |",
			"  |  '--' |    |",
			"   \\      /   /",
			"    '.__.'__.'",
		},
	},
	"garuda": {
		Name: "Garuda",
		Lines: []string{
			"       __",
			"   ___/  \\___",
			"  /  _    _  \\",
			" |  / \\__/ \\  |",
			"  \\ \\_    _/ /",
			"   '._\\__/_.",
			"      Garuda",
		},
	},
	"gentoo": {
		Name: "Gentoo",
		Lines: []string{
			"    .--.",
			"   /    \\",
			"  /  /\\  \\",
			" |  |  |  |",
			"  \\  \\/  /",
			"   \\    /",
			"    '--'",
		},
	},
	"kali": {
		Name: "Kali",
		Lines: []string{
			"       ____",
			"   ___/ __ \\___",
			"  /  _  __  _  \\",
			" |  / \\/  \\/ \\  |",
			"  \\ \\_/\\__/\\_/ /",
			"   '.__    __.'",
			"       '--'",
		},
	},
	"linux": {
		Name: "Linux",
		Lines: []string{
			"       .--.",
			"      |o_o |",
			"      |:_/ |",
			"     //   \\ \\",
			"    (|     | )",
			"   /'\\_   _/`\\",
			"   \\___)=(___/",
		},
	},
	"macos": {
		Name: "macOS",
		Lines: []string{
			"                    'c.",
			"                 ,xNMM.",
			"               .OMMMMo",
			"               lMM\"",
			"     .;loddo:.  .olloddol;.",
			"   cKMMMMMMMMMMNWMMMMMMMMMM0:",
			" .KMMMMMMMMMMMMMMMMMMMMMMMWd.",
			" XMMMMMMMMMMMMMMMMMMMMMMMX.",
			";MMMMMMMMMMMMMMMMMMMMMMMM:",
			":MMMMMMMMMMMMMMMMMMMMMMMM:",
			".MMMMMMMMMMMMMMMMMMMMMMMX.",
			" kMMMMMMMMMMMMMMMMMMMMMMMMWd.",
			" 'XMMMMMMMMMMMMMMMMMMMMMMMMMMk",
			"  'XMMMMMMMMMMMMMMMMMMMMMMMMK.",
			"    kMMMMMMMMMMMMMMMMMMMMMMd",
			"     ;KMMMMMMMWXXWMMMMMMMk.",
			"       \"cooc*\"    \"*coo'\"",
		},
	},
	"manjaro": {
		Name: "Manjaro",
		Lines: []string{
			"  ██████████",
			"  ██████████",
			"  ███",
			"  ███   ███",
			"  ███   ███",
			"  ███   ███",
			"  ███   ███",
		},
	},
	"mint": {
		Name: "Linux Mint",
		Lines: []string{
			"   ___________",
			"  |  _______  |",
			"  | |  ___  | |",
			"  | | |   | | |",
			"  | | |___| | |",
			"  | |_______| |",
			"  |___________|",
		},
	},
	"nixos": {
		Name: "NixOS",
		Lines: []string{
			"   \\  /\\  /",
			"    \\/  \\/",
			"    /\\  /\\",
			"   /  \\/  \\",
			"   \\  /\\  /",
			"    \\/  \\/",
		},
	},
	"opensuse": {
		Name: "openSUSE",
		Lines: []string{
			"       _____",
			"   .-''     ''.",
			"  /  .--.      \\",
			" |  ( () )      |",
			"  \\  '--'  __  /",
			"   '.___.-'  '-'",
		},
	},
	"pop": {
		Name: "Pop!_OS",
		Lines: []string{
			"   __________",
			"  |  ______  |",
			"  | | ____ | |",
			"  | ||_  _|| |",
			"  | |__||__| |",
			"  |__________|",
			"       !_OS",
		},
	},
	"rhel": {
		Name: "RHEL",
		Lines: []string{
			"     ________",
			"   .'  ____  '.",
			"  /   /____\\   \\",
			" |    ______    |",
			"  \\  /      \\  /",
			"   '.__RHEL__.'",
		},
	},
	"rocky": {
		Name: "Rocky Linux",
		Lines: []string{
			"        /\\",
			"       /  \\",
			"      / /\\ \\",
			"     / /  \\ \\",
			"    /_/____\\_\\",
			"      Rocky",
		},
	},
	"ubuntu": {
		Name: "Ubuntu",
		Lines: []string{
			"        .---.",
			"    .--'     '--.",
			"   /  o       o  \\",
			"  |      .-.      |",
			"   \\  o  '-'  o  /",
			"    '--._____.--'",
			"        '---'",
		},
	},
	"void": {
		Name: "Void",
		Lines: []string{
			"     ______",
			"   .'  __  '.",
			"  /   /  \\   \\",
			" |   | () |   |",
			"  \\   \\__/   /",
			"   '.__  __.'",
			"      '--'",
		},
	},
}

func Names() []string {
	names := make([]string, 0, len(catalog))
	for name := range catalog {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func RenderCatalog() string {
	var b strings.Builder
	for _, name := range Names() {
		item := catalog[name]
		b.WriteString(fmt.Sprintf("[%s] %s\n", name, item.Name))
		for _, line := range item.Lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}
