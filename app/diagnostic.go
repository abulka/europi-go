// Diagnostic app
package app

import (
	hw "europi/controls"
	"math/rand"
	"strconv"
	"time"
)

type Diagnostic struct{}

func (c Diagnostic) Name() string { return "Diagnostic Tester" }

type DiagnosticDisplayState struct {
	Knob1, Knob2, Ain int
	Btn1, Btn2        bool
	Din               bool
}

func (d *DiagnosticDisplayState) IsDirty(knob1, knob2, ain int, btn1, btn2 bool, din bool) bool {
	return d.Knob1 != knob1 ||
		d.Knob2 != knob2 ||
		d.Ain != ain ||
		d.Btn1 != btn1 || d.Btn2 != btn2 || d.Din != din
}

func (d *DiagnosticDisplayState) Update(knob1, knob2, ain int, btn1, btn2 bool, din bool) {
	d.Knob1, d.Knob2, d.Ain = knob1, knob2, ain
	d.Btn1, d.Btn2 = btn1, btn2
	d.Din = din
}

func (c Diagnostic) Run(io *hw.Controls) {
	displayState := &DiagnosticDisplayState{}
	const CV_STEP = 10
	const PRINT_STEP = 100
	var cvLoopCount int = CV_STEP

	for loopCount := 0; ; loopCount++ {
		knob1Value := io.K1.Value()
		knob2Value := io.K2.Value()
		ainVolts := io.AIN.Volts()
		ainValue := io.AIN.Value()
		btn1Pressed := io.B1.Pressed()
		btn2Pressed := io.B2.Pressed()
		dinState := io.DIN.Get()

		if ShouldExit(io) {
			break
		}

		if loopCount%PRINT_STEP == 0 {
			println("K1:", knob1Value, "K2:", knob2Value, "AIN:", ainValue, "AINv:", ainVolts, "B1:", btn1Pressed, "B2:", btn2Pressed, "DIN:", dinState)
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
					io.CV1.Set(duty)
				case 2:
					io.CV2.Set(duty)
				case 3:
					io.CV3.Set(duty)
				case 4:
					io.CV4.Set(duty)
				case 5:
					io.CV5.Set(duty)
				case 6:
					io.CV6.Set(duty)
				}
			}
		}

		if displayState.IsDirty(knob1Value, knob2Value, ainValue, btn1Pressed, btn2Pressed, dinState) {
			dinDisp := " "
			if dinState {
				dinDisp = "1"
			}
			ainDisp := strconv.FormatFloat(ainVolts, 'f', 2, 64) + "v"
			io.Display.ClearDisplay()
			io.Display.WriteLine(0, 10, "Knob1: "+strconv.Itoa(knob1Value))
			io.Display.WriteLine(0, 20, "Knob2: "+strconv.Itoa(knob2Value))
			io.Display.WriteLine(0, 30, "B1:"+btn1Msg+" B2:"+btn2Msg)
			io.Display.WriteLine(75, 10, "DIN:"+dinDisp)
			io.Display.WriteLine(75, 20, "AIN:"+ainDisp)
			io.Display.Display()
			displayState.Update(knob1Value, knob2Value, ainValue, btn1Pressed, btn2Pressed, dinState)
		}

		time.Sleep(10 * time.Millisecond)
	}
}
