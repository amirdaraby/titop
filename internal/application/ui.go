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

const (
	OVERALL_STATS_BAR_CHARACTER = "▌"
	OVERALL_STATS_BAR_SPACE     = " "
	USAGE_MAX_LEN               = 8
	MIN_BAR_LENGTH              = 10
	INTERNAL_PADDING            = 1
	GAP_BETWEEN_BOXES           = 1 // Reduced from 2 to 1
	CORES_PER_ROW               = 2

	// Usage thresholds
	LOW_USAGE_THRESHOLD  = 30.0
	HIGH_USAGE_THRESHOLD = 70.0
)

type UI struct {
	screen    tcell.Screen
	styles    uiStyles
	cpu       usage.CPU
	mem       usage.Memory
	processes []usage.Process
}

type uiStyles struct {
	text          tcell.Style
	barBackground tcell.Style
}

func getBarStyle(usage float32) tcell.Style {
	// A deep navy blue background that's easy on the eyes
	baseStyle := tcell.StyleDefault.Background(tcell.NewRGBColor(28, 33, 48))

	// Softer, more modern colors for the bars
	switch {
	case usage < LOW_USAGE_THRESHOLD:
		return baseStyle.Foreground(tcell.NewRGBColor(80, 250, 123)) // Soft mint green
	case usage < HIGH_USAGE_THRESHOLD:
		return baseStyle.Foreground(tcell.NewRGBColor(255, 184, 108)) // Warm orange
	default:
		return baseStyle.Foreground(tcell.NewRGBColor(255, 85, 85)) // Soft red
	}
}

func Init(cancelCtx context.CancelFunc) (UI, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return UI{}, err
	}

	if err := s.Init(); err != nil {
		return UI{}, err
	}

	ui := UI{
		screen: s,
		styles: uiStyles{
			text:          tcell.StyleDefault.Foreground(tcell.NewRGBColor(248, 248, 242)), // Soft white
			barBackground: tcell.StyleDefault.Background(tcell.NewRGBColor(28, 33, 48)),    // Deep navy blue
		},
	}

	ui.setTerminalStyle()
	return ui, nil
}

func (ui *UI) setTerminalStyle() {
	ui.screen.SetStyle(tcell.StyleDefault)
}

func (ui *UI) update(cpu usage.CPU, mem usage.Memory, proccesses []usage.Process) {
	ui.cpu = cpu
	ui.mem = mem
	ui.processes = proccesses
	ui.draw()
}

func (ui *UI) draw() {
	ui.screen.Clear()
	width, _ := ui.screen.Size()

	dimensions := ui.calculateDimensions(width)

	lastPos := ui.renderCPUCores(dimensions)

	ui.renderMemorySection(dimensions, lastPos)

	ui.screen.Show()
}

type displayDimensions struct {
	barLen, boxWidth, totalWidth, startWidth, maxUsageLen int
}

func (ui *UI) calculateDimensions(screenWidth int) displayDimensions {
	maxUsageLen := len(fmt.Sprintf("(%.2f%%)", 100.00))

	// Use full screen width minus minimal spacing
	barLen := (screenWidth - 4 - GAP_BETWEEN_BOXES) / 2 // 4 for minimal total side margins
	if barLen < MIN_BAR_LENGTH {
		barLen = MIN_BAR_LENGTH
	}

	boxWidth := barLen
	totalWidth := boxWidth*2 + GAP_BETWEEN_BOXES

	return displayDimensions{
		barLen:      barLen,
		boxWidth:    boxWidth,
		totalWidth:  totalWidth,
		startWidth:  2, // Just a minimal left margin
		maxUsageLen: maxUsageLen,
	}
}

func (ui *UI) renderCPUCores(dim displayDimensions) int {
	startHeight := 1
	coreStartHeight := startHeight
	coreStartWidth := dim.startWidth
	coreCount := 0

	for idx, core := range ui.cpu.Cores {
		if coreCount == 1 {
			coreStartWidth = dim.startWidth + dim.boxWidth + GAP_BETWEEN_BOXES
		}

		ui.renderCPUBox(
			coreStartWidth,
			coreStartHeight,
			dim.boxWidth,
			idx,
			core.Usage,
			dim.barLen,
			dim.maxUsageLen,
		)

		coreCount++
		if coreCount >= CORES_PER_ROW {
			coreStartWidth = dim.startWidth
			coreStartHeight += 2
			coreCount = 0
		}
	}

	return coreStartHeight + 1
}

func (ui *UI) renderCPUBox(x, y, boxWidth, coreIdx int, usage float32, barLen, maxUsageLen int) {
	currentX := x

	coreTitle := fmt.Sprintf("CPU%d (%.1f%%)", coreIdx, usage)
	emitStr(ui.screen, currentX, y-1, ui.styles.text, coreTitle)

	ui.renderColoredBar(currentX, y, usage, barLen)
}

func (ui *UI) renderMemorySection(dim displayDimensions, startHeight int) {
	memoryTitle := fmt.Sprintf("MEM (%.1f/%.1fG)", float32(ui.mem.Allocated)/float32(1000000), float32(ui.mem.Total)/float32(1000000))
	currentX := dim.startWidth

	// Draw memory title and bar
	emitStr(ui.screen, currentX, startHeight-1, ui.styles.text, memoryTitle)
	ui.renderColoredBar(currentX, startHeight, ui.mem.Usage, dim.barLen)

	if ui.mem.Swap != nil {
		swapX := dim.startWidth + dim.boxWidth + GAP_BETWEEN_BOXES
		swapTitle := fmt.Sprintf("SWP (%.1f/%.1fG)", float32(ui.mem.Swap.Allocated)/float32(1000000), float32(ui.mem.Swap.Total)/float32(1000000))

		// Draw swap title and bar
		emitStr(ui.screen, swapX, startHeight-1, ui.styles.text, swapTitle)
		ui.renderColoredBar(swapX, startHeight, ui.mem.Swap.Usage, dim.barLen)
	}
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
			ui.draw()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				cancelCtx()
				ui.screen.Fini()
				os.Exit(0)
			}
		}
	}
}

func (ui *UI) renderColoredBar(x, y int, usage float32, barLen int) {
	filled := min(int((usage/100)*float32(barLen)), barLen)

	lowPos := int((LOW_USAGE_THRESHOLD / 100) * float32(barLen))
	highPos := int((HIGH_USAGE_THRESHOLD / 100) * float32(barLen))

	emptyBar := strings.Repeat(" ", barLen)
	emitStr(ui.screen, x, y, ui.styles.barBackground, emptyBar)

	greenLen := min(filled, lowPos)
	if greenLen > 0 {
		greenBar := strings.Repeat(OVERALL_STATS_BAR_CHARACTER, greenLen)
		emitStr(ui.screen, x, y, getBarStyle(LOW_USAGE_THRESHOLD-1), greenBar)
	}

	if filled > lowPos {
		yellowLen := min(filled-lowPos, highPos-lowPos)
		if yellowLen > 0 {
			yellowBar := strings.Repeat(OVERALL_STATS_BAR_CHARACTER, yellowLen)
			emitStr(ui.screen, x+lowPos, y, getBarStyle((LOW_USAGE_THRESHOLD+HIGH_USAGE_THRESHOLD)/2), yellowBar)
		}
	}

	if filled > highPos {
		redLen := filled - highPos
		if redLen > 0 {
			redBar := strings.Repeat(OVERALL_STATS_BAR_CHARACTER, redLen)
			emitStr(ui.screen, x+highPos, y, getBarStyle(HIGH_USAGE_THRESHOLD+1), redBar)
		}
	}
}
