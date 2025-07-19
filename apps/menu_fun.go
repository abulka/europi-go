// MenuFun app: a menu of questions, select to see answers, K2 returns to menu
package apps

import (
	"europi/controls"
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
	{"Cats are best!"},
	{"3.141592653589"},
	{"TinyGo is Go", "for", "microcontrollers!"},
	{"Both are great!"},
	{"16 per line!"},
	{"For fun!"},
	{"Modular magic!"},
	{"Both!"},
	{"Hi there!", "How are you?", "Enjoy!"},
}

func (MenuFun) Run(hw *controls.Controls) {

	// Try switching to 4 lines if possible
	if hw.Display.NumLines() == 3 {
		logutil.Println("Switching to 4 lines for MenuFun app.")
		hw.Display.SetNumLines(4)
	}
	
	for {
		choice := firmware.ScrollingMenu(menuFunQuestions, hw, hw.Display.NumLines()) // was 3
		if choice < 0 || choice >= len(menuFunQuestions) {
			logutil.Println("Exiting MenuFun app.")
			hw.Display.ClearDisplay()
			hw.Display.Display()
			logutil.Println("Display cleared. Goodbye!")
			time.Sleep(1 * time.Second)
			return
		}
		// Show answer
		hw.Display.ClearBuffer()
		ans := menuFunAnswers[choice]
		for i := 0; i < 3; i++ {
			if i < len(ans) {
				hw.Display.WriteLine(i, ans[i])
			} else {
				hw.Display.WriteLine(i, "")
			}
		}
		hw.Display.Display()

		// Wait for B1 to return to menu
		for {
			if hw.B1.Pressed() && !hw.B2.Pressed() {
				break // Return to menufun menu
			}
			if firmware.ShouldExit(hw) {
				return
			}
			time.Sleep(2 * time.Millisecond)
		}

		// At this point we have pressed B1. Check if we should exit or continue in this app
		if firmware.ShouldExit(hw) {
			return
		}
	}
}
