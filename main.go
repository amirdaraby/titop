package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/amirdaraby/titop/internal/ui"
	"github.com/amirdaraby/titop/internal/usage"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func main() {

	s, err := tcell.NewScreen()

	if err != nil {
		panic(err)
	}

	err = s.Init()

	if err != nil {
		panic(err)
	}

	tStyle := tcell.StyleDefault.Background(tcell.ColorBlack)

	s.SetStyle(tStyle)

	memRes := make(chan usage.Memory, 1)
	cpuRes := make(chan usage.CPU, 1)

	mainCtx := context.Background()
	ctx, cancelCtx := context.WithCancel(mainCtx)

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				break loop
			default:
				usage.Overall(cpuRes, memRes)
				time.Sleep(time.Second * 1)
			}
		}
	}()

	go func() {
		for {
			ev := s.PollEvent()

			switch ev := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					cancelCtx()
					s.Fini()
					os.Exit(0)
				}
			}
		}
	}()

	for {
		cpu := <-cpuRes
		mem := <-memRes

		Display(s, cpu, mem)
		s.Show()
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

func Display(s tcell.Screen, cpu usage.CPU, mem usage.Memory) {
	s.Clear()

	barLen := ui.GetBarLen(s)

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

		leftCoreBar := ui.GenerateBar(cpu.Cores[i].Usage, barLen)
		rightCoreBar := ui.GenerateBar(cpu.Cores[i+1].Usage, barLen)

		// left core
		emitStr(s, coreStartWidth, coreStartHeight, usageStyle, leftCoreTitle)
		coreStartWidth += len(leftCoreTitle)

		emitStr(s, coreStartWidth, coreStartHeight, barStyle, leftCoreBar)
		coreStartWidth += barLen

		emitStr(s, coreStartWidth, coreStartHeight, usageStyle, leftCoreUsage)
		coreStartWidth += ui.USAGE_MAX_LEN + 2

		// right core
		emitStr(s, coreStartWidth, coreStartHeight, usageStyle, rightCoreTitle)
		coreStartWidth += len(rightCoreTitle)

		emitStr(s, coreStartWidth, coreStartHeight, barStyle, rightCoreBar)
		coreStartWidth += barLen

		emitStr(s, coreStartWidth, coreStartHeight, usageStyle, rightCoreUsage)

		coreStartHeight += 2
	}

	startHeight += coreStartHeight

	memoryTitle := "MEM:"
	memoryBar := ui.GenerateBar(mem.Usage, barLen)
	memoryUsage := fmt.Sprintf("(%.2fGB/%.2fGB)", float32(mem.Allocated)/float32(1000000), float32(mem.Total)/float32(1000000))

	memoryBarWidth := startWidth + len(memoryTitle)
	memoryUsageWidth := memoryBarWidth + barLen

	emitStr(s, startWidth, startHeight, usageStyle, memoryTitle)
	emitStr(s, memoryBarWidth, startHeight, barStyle, memoryBar)
	emitStr(s, memoryUsageWidth, startHeight, usageStyle, memoryUsage)

	startHeight += 2
	if mem.Swap != nil {
		swapTitle := "SWP:"
		swapBar := ui.GenerateBar(mem.Swap.Usage, barLen)
		swapUsage := fmt.Sprintf("(%.2fGB/%.2fGB)", float32(mem.Swap.Allocated)/float32(1000000), float32(mem.Swap.Total)/float32(1000000))

		emitStr(s, startWidth, startHeight, usageStyle, swapTitle)
		emitStr(s, memoryBarWidth, startHeight, barStyle, swapBar)
		emitStr(s, memoryUsageWidth, startHeight, usageStyle, swapUsage)
	}
}
