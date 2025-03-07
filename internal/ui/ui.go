package ui

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

const OVERALL_STATS_BAR_CHARACTER = "▌"
const OVERALL_STATS_BAR_SPACE = " "
const USAGE_MAX_LEN = 8

func GenerateBar(usage float32, len int) (bar string) {
	filled := int((usage / 100) * float32(len))

	return strings.Repeat(OVERALL_STATS_BAR_CHARACTER, filled) + strings.Repeat(OVERALL_STATS_BAR_SPACE, len-filled)
}

func GetBarLen(s tcell.Screen) int {
	return 20
}
