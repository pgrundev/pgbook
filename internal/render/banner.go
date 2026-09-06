package render

import "strings"

// logo is the pgbook wordmark. Kept under 80 columns so it never wraps
// in a default terminal.
var logo = []string{
	`             _                 _    `,
	` _ __   __ _| |__   ___   ___ | | __`,
	`| '_ \ / _` + "`" + ` | '_ \ / _ \ / _ \| |/ /`,
	`| |_) | (_| | |_) | (_) | (_) |   < `,
	`| .__/ \__, |_.__/ \___/ \___/|_|\_\`,
	`|_|    |___/                        `,
}

// Banner returns the greeting logo printed by a bare `pgbook`.
func Banner(color bool) string {
	var b strings.Builder
	b.WriteString("\n")
	for _, line := range logo {
		line = strings.TrimRight(line, " ")
		if color {
			line = ansiBold + ansiCyan + line + ansiReset
		}
		b.WriteString("  " + line + "\n")
	}
	tagline := "the Postgres Book in your terminal · pgbook.dev"
	if color {
		tagline = ansiDim + tagline + ansiReset
	}
	b.WriteString("\n  " + tagline + "\n")
	return b.String()
}
