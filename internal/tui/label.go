package tui

import (
	"fmt"
	"math"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// foregroundForBackground returns "0" (black) or "255" (white) — whichever has
// better contrast against the given hex background, using W3C relative luminance.
func foregroundForBackground(hex string) lipgloss.Color {
	l, err := hexLuminance(hex)
	if err != nil || l > 0.179 {
		if err != nil {
			return "255"
		}

		return "0"
	}

	return "255"
}

func hexLuminance(hex string) (float64, error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return 0, err
	}

	return 0.2126*linearize(r) + 0.7152*linearize(g) + 0.0722*linearize(b), nil
}

func linearize(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}

	return math.Pow((c+0.055)/1.055, 2.4)
}

func parseHex(hex string) (r, g, b float64, err error) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %q", hex)
	}

	ri, err := strconv.ParseInt(hex[0:2], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	gi, err := strconv.ParseInt(hex[2:4], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	bi, err := strconv.ParseInt(hex[4:6], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255, nil
}

func renderLabelPill(name, hexColor string) string {
	fg := foregroundForBackground(hexColor)
	style := lipgloss.NewStyle().
		Background(lipgloss.Color(hexColor)).
		Foreground(fg).
		Padding(0, 1)

	return style.Render(name)
}
