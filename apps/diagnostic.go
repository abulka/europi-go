// Diagnostic app
package apps

import (
	"europi/controls"
	"europi/firmware"
	"europi/logutil"
	"math/rand"
	"strconv"
	"time"
)

type Diagnostic struct{}

func (c Diagnostic) Name() string { return "Diagnostic Tester" }

func (c Diagnostic) Run(hw *controls.Controls) {
	hw.Display.SetNumLines(4)

	const CV_STEP = 10
	const PRINT_STEP = 100
	var cvLoopCount int = CV_STEP

	for loopCount := 0; ; loopCount++ {
		knob1Value := hw.K1.Value()
		knob2Value := hw.K2.Value()
		ainVolts := hw.AIN.Volts()
		ainValue := hw.AIN.Value()
		btn1Pressed := hw.B1.Pressed()
		btn2Pressed := hw.B2.Pressed()
		dinState := hw.DIN.Get()

		if firmware.ShouldExit(hw) {
			break
		}

		if loopCount%PRINT_STEP == 0 {
			logutil.Println("K1:", knob1Value, "K2:", knob2Value, "AIN:", ainValue, "AINv:", ainVolts, "B1:", btn1Pressed, "B2:", btn2Pressed, "DIN:", dinState)
		}

		btn1Msg := "Up"
		btn2Msg := "Up"
		if btn1Pressed {
			btn1Msg = "Down"
		}
		if btn2Pressed {
			btn2Msg = "Down"
		}

		cvLoopCount--
		if cvLoopCount <= 0 {
			cvLoopCount = CV_STEP
			for cv := 1; cv <= 6; cv++ {
				duty := uint32(rand.Intn(10000)) // Replace 10000 with io.CV1.MaxDuty if needed
				switch cv {
				case 1:
					hw.CV1.Set(duty)
				case 2:
					hw.CV2.Set(duty)
				case 3:
					hw.CV3.Set(duty)
				case 4:
					hw.CV4.Set(duty)
				case 5:
					hw.CV5.Set(duty)
				case 6:
					hw.CV6.Set(duty)
				}
			}
		}

		dinDisp := " "
		if dinState {
			dinDisp = "1"
		}
		ainDisp := strconv.FormatFloat(ainVolts, 'f', 2, 64) + "v"
		// io.Display.ClearDisplay()
		hw.Display.ClearBuffer()
		hw.Display.WriteLine(0, "Knob1: "+strconv.Itoa(knob1Value))
		hw.Display.WriteLine(1, "Knob2: "+strconv.Itoa(knob2Value))
		hw.Display.WriteLine(2, "B1:"+btn1Msg+" B2:"+btn2Msg)
		hw.Display.WriteLine(3, "DIN:"+dinDisp+" AIN:"+ainDisp)
		hw.Display.Display()

		time.Sleep(10 * time.Millisecond)
	}
}
