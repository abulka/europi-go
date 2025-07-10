// MenuFun app: a menu of questions, select to see answers, K2 returns to menu
package apps

import (
	hw "europi/controls"
	"europi/firmware"
	"europi/logutil"
	"time"
)

type MenuFun struct{}

func (MenuFun) Name() string { return "Menu Fun" }

var menuFunQuestions = []string{
	"What is Go?",
	"Best animal?",
	"Pi digits?",
	"TinyGo?",
	"Tea or Coffee?",
	"Max chars?",
	"Why EuroPi?",
	"Best synth?",
	"Cats or Dogs?",
	"Hello?",
}

var menuFunAnswers = [][]string{
	{"A language!"},
	{"Capybara!"},
	{"3.141592653589"},
	{"Go for MCUs!"},
	{"Both are great!"},
	{"16 per line!"},
	{"For fun!"},
	{"Modular magic!"},
	{"Both!"},
	{"Hi there!", "How are you?", "Enjoy!"},
}

func (MenuFun) Run(io *hw.Controls) {
	for {
		choice := firmware.ScrollingMenu(menuFunQuestions, io, 3)
		if choice < 0 || choice >= len(menuFunQuestions) {
			logutil.Println("Exiting MenuFun app.")
			io.Display.ClearDisplay()
			io.Display.Display()
			logutil.Println("Display cleared. Goodbye!")
			time.Sleep(1 * time.Second)
			return
		}
		// Show answer
		for {
			io.Display.ClearDisplay()
			ans := menuFunAnswers[choice]
			for i := 0; i < 3; i++ {
				if i < len(ans) {
					io.Display.WriteLine(i, ans[i])
				} else {
					io.Display.WriteLine(i, "")
				}
			}
			io.Display.Display()
			if io.B1.Pressed() && !io.B2.Pressed() {
				for io.B1.Pressed() {
					time.Sleep(10 * time.Millisecond)
					if firmware.ShouldExit(io) {
						return
					}
				}
				break // Return to menu
			}
			if firmware.ShouldExit(io) {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		if firmware.ShouldExit(io) {
			return
		}
	}
}
