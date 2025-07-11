//go:build tinygo

// Real hardware implementations (TinyGo only)
package controls

import (
	"europi/display"
	"europi/util"
	"machine"
)

// DigitalInput abstraction (with optional inversion)
type DigitalInput struct {
	pin      machine.Pin
	inverted bool
}

func NewDigitalInput(pin machine.Pin, inverted bool) *DigitalInput {
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	return &DigitalInput{pin: pin, inverted: inverted}
}

func (d *DigitalInput) Get() bool {
	v := d.pin.Get()
	if d.inverted {
		return !v
	}
	return v
}

// Button abstraction
type Button struct {
	input *DigitalInput
}

func NewButton(pin machine.Pin) *Button {
	return &Button{input: NewDigitalInput(pin, true)} // true: pressed = low
}

func (b *Button) Pressed() bool {
	return b.input.Get()
}

// Knob abstraction
type Knob struct {
	adc  machine.ADC
	proc *util.SmartKnobProcessor
}

func NewKnob(adcPin machine.Pin) *Knob {
	adc := machine.ADC{Pin: adcPin}
	adc.Configure(machine.ADCConfig{})
	return &Knob{
		adc:  adc,
		proc: util.NewSmartKnobProcessor(),
	}
}

func (k *Knob) Value() int {
	return k.proc.Process(int(k.adc.Get()))
}

// AnalogueInput abstraction
type AnalogueInput struct {
	adc    machine.ADC
	reader *util.VoltageReader
}

func NewAnalogueInput(adcPin machine.Pin) *AnalogueInput {
	adc := machine.ADC{Pin: adcPin}
	adc.Configure(machine.ADCConfig{})
	return &AnalogueInput{
		adc:    adc,
		reader: util.NewVoltageReader(288, 22000, 16, 3),
	}
}

func (a *AnalogueInput) Volts() float64 {
	return a.reader.Read(int(a.adc.Get()))
}

func (a *AnalogueInput) Value() int {
	return int(a.Volts() * 100)
}

// CV output abstraction
type CV struct {
	Index int
}

var MaxDuty uint32 = 9999

func SetCV(cv int, value uint32) {
	switch cv {
	case 1:
		machine.PWM2.Set(1, value)
	case 2:
		machine.PWM2.Set(0, value)
	case 3:
		machine.PWM0.Set(0, value)
	case 4:
		machine.PWM0.Set(1, value)
	case 5:
		machine.PWM1.Set(0, value)
	case 6:
		machine.PWM1.Set(1, value)
	}
}

func (c *CV) Set(value uint32) {
	SetCV(c.Index, value)
}

func (c *CV) On() {
	SetCV(c.Index, MaxDuty)
}

func (c *CV) Off() {
	SetCV(c.Index, 0)
}

func ConfigureCV() {
	const pwmFrequency = 20000
	const pwmPeriod = 1e9 / pwmFrequency

	machine.PWM0.Configure(machine.PWMConfig{Period: pwmPeriod})
	machine.PWM1.Configure(machine.PWMConfig{Period: pwmPeriod})
	machine.PWM2.Configure(machine.PWMConfig{Period: pwmPeriod})

	machine.GPIO21.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO20.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO16.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO17.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO18.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO19.Configure(machine.PinConfig{Mode: machine.PinPWM})

	if MaxDuty != machine.PWM0.Top() {
		MaxDuty = machine.PWM0.Top()
	}
}

// Initializes all real hardware IO
func SetupEuroPiWithDisplay(display display.IOledDevice) *Controls {
	machine.InitADC()
	ConfigureCV()
	return &Controls{
		K1:      NewKnob(machine.ADC1),
		K2:      NewKnob(machine.ADC2),
		B1:      NewButton(machine.GPIO4),
		B2:      NewButton(machine.GPIO5),
		DIN:     NewDigitalInput(machine.GPIO22, true),
		AIN:     NewAnalogueInput(machine.ADC0),
		CV1:     &CV{Index: 1},
		CV2:     &CV{Index: 2},
		CV3:     &CV{Index: 3},
		CV4:     &CV{Index: 4},
		CV5:     &CV{Index: 5},
		CV6:     &CV{Index: 6},
		Display: display,
	}
}
