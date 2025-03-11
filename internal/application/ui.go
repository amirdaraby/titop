package application

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/amirdaraby/titop/internal/usage"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

const OVERALL_STATS_BAR_CHARACTER = "▌"
const OVERALL_STATS_BAR_SPACE = " "
const USAGE_MAX_LEN = 8

type UI struct {
	screen tcell.Screen
}

func Init(cancelCtx context.CancelFunc) (UI, error) {

	s, err := tcell.NewScreen()

	ui := UI{
		screen: s,
	}

	if err != nil {
		return ui, err
	}

	err = s.Init()

	if err != nil {
		return ui, err
	}

	ui.setTerminalStyle()

	return ui, err
}

func (ui *UI) setTerminalStyle() {
	style := tcell.StyleDefault

	ui.screen.SetStyle(style)
}

func (ui *UI) display(cpu usage.CPU, mem usage.Memory) {
	ui.screen.Clear()

	barLen := getBarLen(ui.screen)

	startWidth := 2
	startHeight := 2

	barStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(250, 242, 161)).Background(tcell.NewRGBColor(69, 63, 120))
	usageStyle := tcell.StyleDefault.Foreground(tcell.ColorWhiteSmoke.TrueColor()).Background(tcell.ColorBlack)

	coreStartHeight := startHeight

	for i := 0; i < len(cpu.Cores)-1; i += 2 {
		coreStartWidth := startWidth

		leftCoreTitle := fmt.Sprintf("%d:", i)
		rightCoreTitle := fmt.Sprintf("%d:", i+1)
		leftCoreUsage := fmt.Sprintf("(%.2f%%)", cpu.Cores[i].Usage)
		rightCoreUsage := fmt.Sprintf("(%.2f%%)", cpu.Cores[i+1].Usage)

		leftCoreBar := generateBar(ui.screen, cpu.Cores[i].Usage)
		rightCoreBar := generateBar(ui.screen, cpu.Cores[i+1].Usage)

		// left core
		emitStr(ui.screen, coreStartWidth, coreStartHeight, usageStyle, leftCoreTitle)
		coreStartWidth += len(leftCoreTitle)

		emitStr(ui.screen, coreStartWidth, coreStartHeight, barStyle, leftCoreBar)
		coreStartWidth += barLen

		emitStr(ui.screen, coreStartWidth, coreStartHeight, usageStyle, leftCoreUsage)
		coreStartWidth += USAGE_MAX_LEN + 2

		// right core
		emitStr(ui.screen, coreStartWidth, coreStartHeight, usageStyle, rightCoreTitle)
		coreStartWidth += len(rightCoreTitle)

		emitStr(ui.screen, coreStartWidth, coreStartHeight, barStyle, rightCoreBar)
		coreStartWidth += barLen

		emitStr(ui.screen, coreStartWidth, coreStartHeight, usageStyle, rightCoreUsage)

		coreStartHeight += 2
	}

	startHeight += coreStartHeight

	memoryTitle := "MEM:"
	memoryBar := generateBar(ui.screen, mem.Usage)
	memoryUsage := fmt.Sprintf("(%.2fGB/%.2fGB)", float32(mem.Allocated)/float32(1000000), float32(mem.Total)/float32(1000000))

	memoryBarWidth := startWidth + len(memoryTitle)
	memoryUsageWidth := memoryBarWidth + barLen

	emitStr(ui.screen, startWidth, startHeight, usageStyle, memoryTitle)
	emitStr(ui.screen, memoryBarWidth, startHeight, barStyle, memoryBar)
	emitStr(ui.screen, memoryUsageWidth, startHeight, usageStyle, memoryUsage)

	startHeight += 2
	if mem.Swap != nil {
		swapTitle := "SWP:"
		swapBar := generateBar(ui.screen, mem.Swap.Usage)
		swapUsage := fmt.Sprintf("(%.2fGB/%.2fGB)", float32(mem.Swap.Allocated)/float32(1000000), float32(mem.Swap.Total)/float32(1000000))

		emitStr(ui.screen, startWidth, startHeight, usageStyle, swapTitle)
		emitStr(ui.screen, memoryBarWidth, startHeight, barStyle, swapBar)
		emitStr(ui.screen, memoryUsageWidth, startHeight, usageStyle, swapUsage)
	}

	ui.screen.Show()
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func (ui *UI) pollAndListenToEvents(cancelCtx context.CancelFunc) {
	for {
		ev := ui.screen.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventResize:
			ui.screen.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				cancelCtx()
				ui.screen.Fini()
				os.Exit(0)
			}
		}
	}
}

func generateBar(s tcell.Screen, usage float32) (bar string) {
	len := getBarLen(s)
	filled := int((usage / 100) * float32(len))

	return strings.Repeat(OVERALL_STATS_BAR_CHARACTER, filled) + strings.Repeat(OVERALL_STATS_BAR_SPACE, len-filled)
}

func getBarLen(s tcell.Screen) int {
	return 20
}
