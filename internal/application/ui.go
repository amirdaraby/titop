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
	INTERNAL_PADDING            = 2
	BOX_BORDER_WIDTH            = 2
	GAP_BETWEEN_BOXES           = 0
	CORES_PER_ROW               = 2

	// Usage thresholds
	LOW_USAGE_THRESHOLD  = 40.0
	HIGH_USAGE_THRESHOLD = 70.0
)

type UI struct {
	screen tcell.Screen
	styles uiStyles
}

type uiStyles struct {
	border        tcell.Style
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
			border:        tcell.StyleDefault.Foreground(tcell.NewRGBColor(98, 114, 164)),  // Muted blue-gray
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

func (ui *UI) display(cpu usage.CPU, mem usage.Memory) {
	ui.screen.Clear()
	width, _ := ui.screen.Size()

	dimensions := ui.calculateDimensions(width)

	lastPos := ui.renderCPUCores(cpu.Cores, dimensions)

	ui.renderMemorySection(mem, dimensions, lastPos)

	ui.screen.Show()
}

type displayDimensions struct {
	barLen, boxWidth, totalWidth, startWidth, maxUsageLen int
}

func (ui *UI) calculateDimensions(screenWidth int) displayDimensions {
	maxUsageLen := len(fmt.Sprintf("(%.2f%%)", 100.00))

	extraSpace := 4
	fixedWidth := len("00:") + maxUsageLen
	spaceForBars := screenWidth - (2 * (fixedWidth + extraSpace)) - 3
	barLen := spaceForBars / 2
	if barLen < MIN_BAR_LENGTH {
		barLen = MIN_BAR_LENGTH
	}

	boxWidth := len("00:") + barLen + maxUsageLen + INTERNAL_PADDING
	totalWidth := (boxWidth+BOX_BORDER_WIDTH)*2 + 1

	startWidth := (screenWidth - totalWidth) / 2
	if startWidth < 1 {
		startWidth = 1
	}
	startWidth += 2

	return displayDimensions{
		barLen:      barLen,
		boxWidth:    boxWidth,
		totalWidth:  totalWidth,
		startWidth:  startWidth,
		maxUsageLen: maxUsageLen,
	}
}

func (ui *UI) renderCPUCores(cores []usage.Core, dim displayDimensions) int {
	startHeight := 2
	coreStartHeight := startHeight
	coreStartWidth := dim.startWidth
	coreCount := 0

	for idx, core := range cores {
		if coreCount == 1 {
			coreStartWidth = dim.startWidth + dim.boxWidth + 3
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
			coreStartHeight += 3
			coreCount = 0
		}
	}

	return coreStartHeight + 1
}

func (ui *UI) renderCPUBox(x, y, boxWidth, coreIdx int, usage float32, barLen, maxUsageLen int) {
	ui.drawBoxBorders(x, y, boxWidth)

	currentX := x

	coreTitle := fmt.Sprintf("%d:", coreIdx)
	emitStr(ui.screen, currentX, y, ui.styles.text, coreTitle)
	currentX += len(coreTitle)

	// Draw background first
	emptyBar := strings.Repeat(" ", barLen)
	emitStr(ui.screen, currentX, y, ui.styles.barBackground, emptyBar)

	// Draw the bar on top
	bar := generateBarWithLen(barLen, usage)
	emitStr(ui.screen, currentX, y, getBarStyle(usage), bar)
	currentX += barLen

	usageText := fmt.Sprintf("(%.2f%%)", usage)
	usageText = fmt.Sprintf("%*s", maxUsageLen, usageText)
	emitStr(ui.screen, currentX, y, ui.styles.text, usageText)
}

func (ui *UI) renderMemorySection(mem usage.Memory, dim displayDimensions, startHeight int) {
	memoryTitle := "MEM:"
	memoryBar := generateBarWithLen(dim.barLen*2+3, mem.Usage)
	memoryUsage := fmt.Sprintf("(%.2fGB/%.2fGB)", float32(mem.Allocated)/float32(1000000), float32(mem.Total)/float32(1000000))

	memBoxWidth := len(memoryTitle) + (dim.barLen*2 + 3) + len(memoryUsage) + 2
	memStartWidth := (dim.totalWidth - (memBoxWidth + 2)) / 2
	if memStartWidth < 1 {
		memStartWidth = 1
	}
	memStartWidth += dim.startWidth - 1

	ui.drawBoxBorders(memStartWidth, startHeight, memBoxWidth)

	currentX := memStartWidth
	emitStr(ui.screen, currentX, startHeight, ui.styles.text, memoryTitle)
	currentX += len(memoryTitle)

	// Draw background first
	emptyBar := strings.Repeat(" ", dim.barLen*2+3)
	emitStr(ui.screen, currentX, startHeight, ui.styles.barBackground, emptyBar)

	// Draw the bar on top
	emitStr(ui.screen, currentX, startHeight, getBarStyle(mem.Usage), memoryBar)
	currentX += dim.barLen*2 + 3
	emitStr(ui.screen, currentX, startHeight, ui.styles.text, memoryUsage)

	if mem.Swap != nil {
		startHeight += 2
		swapTitle := "SWP:"
		swapBar := generateBarWithLen(dim.barLen*2+3, mem.Swap.Usage)
		swapUsage := fmt.Sprintf("(%.2fGB/%.2fGB)", float32(mem.Swap.Allocated)/float32(1000000), float32(mem.Swap.Total)/float32(1000000))

		ui.drawBoxBorders(memStartWidth, startHeight, memBoxWidth)

		currentX = memStartWidth
		emitStr(ui.screen, currentX, startHeight, ui.styles.text, swapTitle)
		currentX += len(swapTitle)

		// Draw background first
		emitStr(ui.screen, currentX, startHeight, ui.styles.barBackground, emptyBar)

		// Draw the bar on top
		emitStr(ui.screen, currentX, startHeight, getBarStyle(mem.Swap.Usage), swapBar)
		currentX += dim.barLen*2 + 3
		emitStr(ui.screen, currentX, startHeight, ui.styles.text, swapUsage)
	}
}

func (ui *UI) drawBoxBorders(x, y, width int) {
	emitStr(ui.screen, x-1, y-1, ui.styles.border, "+"+strings.Repeat("-", width)+"+")
	emitStr(ui.screen, x-1, y+1, ui.styles.border, "+"+strings.Repeat("-", width)+"+")
	emitStr(ui.screen, x-1, y, ui.styles.border, "|")
	emitStr(ui.screen, x+width, y, ui.styles.border, "|")
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

func generateBarWithLen(length int, usage float32) string {
	filled := int((usage / 100) * float32(length))
	if filled > length {
		filled = length
	}
	return strings.Repeat(OVERALL_STATS_BAR_CHARACTER, filled) + strings.Repeat(OVERALL_STATS_BAR_SPACE, length-filled)
}
