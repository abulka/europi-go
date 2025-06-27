// tinygo flash -target=pico --monitor ./research/europi

package main

import (
	"image/color"
	"machine"
	"math/rand"
	"strconv"
	"time"

	"tinygo.org/x/drivers/ssd1306"

	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

const debugPrint = false
const version = "v0.01"
const production = false

// Control Voltage PWM Outputs 1..6 Init

// SetCV abstracts setting a CV output by index (1-6) and value (0..max)
func SetCV(cv int, value uint32) {
	switch cv {
	case 1:
		machine.PWM2.Set(1, value) // CV1 (GPIO21) - PWM2B
	case 2:
		machine.PWM2.Set(0, value) // CV2 (GPIO20) - PWM2A
	case 3:
		machine.PWM0.Set(0, value) // CV3 (GPIO16) - PWM0A
	case 4:
		machine.PWM0.Set(1, value) // CV4 (GPIO17) - PWM0B
	case 5:
		machine.PWM1.Set(0, value) // CV5 (GPIO18) - PWM1A
	case 6:
		machine.PWM1.Set(1, value) // CV6 (GPIO19) - PWM1B
	}
}

var MaxDuty uint32 = 9999 // Default max duty cycle for PWM

type ICV interface {
	Set(value uint32) // Set the CV value
	On()              // Set CV to max duty
	Off()             // Set CV to 0 duty
}

// CV represents a Control Voltage output with an index (1-6)
// It implements the ICV interface for setting values and turning on/off.
type CV struct {
	Index int
}

func (c *CV) Set(value uint32) {
	SetCV(c.Index, value)
}

func (c *CV) On() {
	SetCV(c.Index, MaxDuty) // Set to max duty
}

func (c *CV) Off() {
	SetCV(c.Index, 0) // Set to 0 duty
}

// Configure all PWM slices and pins, return the max duty cycle
func ConfigureCV() {
	const pwmFrequency = 20000           // Hz
	const pwmPeriod = 1e9 / pwmFrequency // Period in nanoseconds

	machine.PWM0.Configure(machine.PWMConfig{Period: pwmPeriod})
	machine.PWM1.Configure(machine.PWMConfig{Period: pwmPeriod})
	machine.PWM2.Configure(machine.PWMConfig{Period: pwmPeriod})

	machine.GPIO21.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO20.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO16.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO17.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO18.Configure(machine.PinConfig{Mode: machine.PinPWM})
	machine.GPIO19.Configure(machine.PinConfig{Mode: machine.PinPWM})

	// All slices have the same Top value after config, so we can use PWM0's Top as MaxDuty
	if MaxDuty != machine.PWM0.Top() {
		println("Warning: Resetting MaxDuty to PWM0 Top value:", machine.PWM0.Top())
		MaxDuty = machine.PWM0.Top()
	}
}

// OLED Init

// InitDisplay sets up the I2C and SSD1306 display and returns the display instance.
func InitDisplay() IOledDevice {
	i2c := machine.I2C0
	i2c.Configure(machine.I2CConfig{
		Frequency: 400000, // 400kHz CRITICAL!!
		SDA:       machine.GP0,
		SCL:       machine.GP1,
	})
	dev := ssd1306.NewI2C(i2c)
	dev.Configure(ssd1306.Config{
		Address: 0x3C, // CRITITCAL!
		Width:   128,  // Explicitly set width
		Height:  32,   // CRITITCAL!
	})
	return &SSD1306Adapter{dev: dev}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

const ainMin = 288
const ainMax = 22000

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// AnalogFilter with configurable size and range calibration - A "low-pass filter" perhaps
//   - A filter averages out rapid fluctuations that might otherwise trigger hysteresis
type AnalogFilter struct {
	history  []int // Slice for flexible window size
	index    int
	sum      int
	capacity int
}

func NewAnalogFilter(windowSize int) *AnalogFilter {
	return &AnalogFilter{
		history:  make([]int, windowSize),
		capacity: windowSize,
	}
}

func (f *AnalogFilter) Update(value int) int {
	f.sum -= f.history[f.index]
	f.history[f.index] = value
	f.sum += value
	f.index = (f.index + 1) % len(f.history)
	return f.sum / len(f.history)
}

// HysteresisFilter with edge-case handling
//   - HysteresisFilter provides final "gatekeeping" after filtering, applies a threshold to prevent rapid toggling
type HysteresisFilter struct {
	lastValue int
	threshold int
	min       int // Minimum expected value
	max       int // Maximum expected value
}

func (f *HysteresisFilter) Update(value int) int {
	// Always update if we're at min/max boundaries
	if value <= f.min || value >= f.max {
		f.lastValue = value
		return value
	}
	if abs(value-f.lastValue) >= f.threshold {
		f.lastValue = value
	}
	return f.lastValue
}

// VoltageReader handles ADC to voltage conversion with calibration
type VoltageReader struct {
	minRaw   int
	maxRaw   int
	scale    float64
	filter   *AnalogFilter
	hyst     *HysteresisFilter
	lastShow time.Time
}

func NewVoltageReader(minRaw, maxRaw int, filterSize, hystThreshold int) *VoltageReader {
	return &VoltageReader{
		minRaw: minRaw,
		maxRaw: maxRaw,
		scale:  5.0 / float64(maxRaw-minRaw),
		filter: NewAnalogFilter(filterSize),
		hyst: &HysteresisFilter{
			threshold: hystThreshold,
			min:       0,
			max:       500, // 5.00V * 100
		},
	}
}

func (v *VoltageReader) Read(raw int) float64 {
	clamped := clamp(raw, v.minRaw, v.maxRaw)
	scaled := float64(clamped-v.minRaw) * v.scale
	filtered := v.filter.Update(int(scaled * 100))
	stable := v.hyst.Update(filtered)
	return float64(stable) / 100.0
}

// KnobProcessor with smart locking and filtering
//   - SmartKnobProcessor processes raw knob values, applies filtering, and manages locking behavior
//
// A low-pass filter (AnalogFilter) to smooth ADC readings. A time-based lock
// that filters out small jitters after inactivity. A resumeThreshold to avoid
// resuming for tiny bumps.
//
// Adding HysteresisFilter would: Delay value changes even while the knob is
// moving, which directly conflicts with your goal of 1-step precision during
// activity. Possibly block valid inputs, especially near boundaries or fine movements.
type SmartKnobProcessor struct {
	filter           *AnalogFilter
	lastMapped       int
	lockedValue      int
	isLocked         bool
	lockAfter        time.Duration // idle time to lock
	resumeThreshold  int
	lastActivityTime time.Time
}

func NewSmartKnobProcessor() *SmartKnobProcessor {
	return &SmartKnobProcessor{
		filter:           NewAnalogFilter(8), // To smooth more, increase NewAnalogFilter(N) window size (e.g. 12)
		lastMapped:       -1,
		lockedValue:      -1,
		isLocked:         false,
		lockAfter:        500 * time.Millisecond, // To lock faster, reduce lockAfter (e.g. 300ms)
		resumeThreshold:  2,                      // To require more distance to resume, raise resumeThreshold (e.g. 3)
		lastActivityTime: time.Now(),
	}
}

func (k *SmartKnobProcessor) Process(rawValue int) int {
	filtered := k.filter.Update(rawValue)
	mapped := 99 - calibrateKnobValue(filtered, 0, 65535, 0, 99)
	now := time.Now()

	// First run
	if k.lastMapped == -1 {
		k.lastMapped = mapped
		k.lockedValue = mapped
		k.lastActivityTime = now
		return mapped
	}

	// If mapped value changed
	if mapped != k.lastMapped {
		k.lastActivityTime = now

		if k.isLocked {
			// Unlock only if jump is large enough
			if abs(mapped-k.lockedValue) >= k.resumeThreshold {
				k.isLocked = false
				k.lockedValue = mapped
			}
		} else {
			// Actively moving, update immediately
			k.lockedValue = mapped
		}

		k.lastMapped = mapped
	}

	// If no activity for a while, lock the value
	if !k.isLocked && now.Sub(k.lastActivityTime) > k.lockAfter {
		k.isLocked = true
	}

	return k.lockedValue
}

// calibrateKnobValue maps raw ADC values to a range with deadzones at extremes.
func calibrateKnobValue(raw, inMin, inMax, outMin, outMax int) int {
	// Expanded deadzones at extremes with smooth transitions
	deadzone := 16
	if raw <= inMin+deadzone {
		return outMin
	}
	if raw >= inMax-deadzone {
		return outMax
	}

	// Linear mapping with floating point for better precision
	scale := float64(outMax-outMin) / float64(inMax-inMin-2*deadzone)
	val := outMin + int(float64(raw-inMin-deadzone)*scale+0.5) // Round to nearest

	return clamp(val, outMin, outMax)
}

// --- Abstractions for Inputs ---
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

// Button abstraction (with optional debounce)
type Button struct {
	input *DigitalInput
}

var _ IButton = (*Button)(nil) // Ensure Button implements IButton

func NewButton(pin machine.Pin) *Button {
	return &Button{input: NewDigitalInput(pin, true)} // true: pressed = low
}

func (b *Button) Pressed() bool {
	return b.input.Get()
}

type IButton interface {
	Pressed() bool
}

type IKnob interface {
	Value() int // Returns the current value of the knob
}

// Knob abstraction
// Use machine.ADC struct with Pin field for Pico compatibility
type Knob struct {
	adc  machine.ADC
	proc *SmartKnobProcessor
}

func NewKnob(adcPin machine.Pin) *Knob {
	adc := machine.ADC{Pin: adcPin}
	adc.Configure(machine.ADCConfig{})
	return &Knob{
		adc:  adc,
		proc: NewSmartKnobProcessor(),
	}
}

func (k *Knob) Value() int {
	return k.proc.Process(int(k.adc.Get()))
}

// Analog input abstraction (e.g. for AIN)
type AnalogueInput struct {
	adc    machine.ADC
	reader *VoltageReader
}

var _ IDigitalInput = (*DigitalInput)(nil)   // Ensure DigitalInput implements IDigitalInput
var _ IAnalogueInput = (*AnalogueInput)(nil) // Ensure AnalogueInput implements IAnalogueInput

func NewAnalogueInput(adcPin machine.Pin) *AnalogueInput {
	adc := machine.ADC{Pin: adcPin}
	adc.Configure(machine.ADCConfig{})
	return &AnalogueInput{
		adc:    adc,
		reader: NewVoltageReader(ainMin, ainMax, 16, 3),
	}
}

func (a *AnalogueInput) Volts() float64 {
	return a.reader.Read(int(a.adc.Get()))
}

func (a *AnalogueInput) Value() int {
	return int(a.Volts() * 100)
}

// --- Interface definitions for IO abstractions ---
type IDigitalInput interface {
	Get() bool
}

type IAnalogueInput interface {
	Volts() float64
	Value() int
}

type IOledDevice interface {
	ClearDisplay()
	Display()
	WriteLine(x, y int16, text string)
}

// Adapter for ssd1306.Device to implement OledDevice
// Holds the ssd1306.Device and implements WriteLine using tinyfont

type SSD1306Adapter struct {
	dev ssd1306.Device
}

func (o *SSD1306Adapter) ClearDisplay() {
	o.dev.ClearDisplay()
}

func (o *SSD1306Adapter) Display() {
	o.dev.Display()
}

func (o *SSD1306Adapter) WriteLine(x, y int16, text string) {
	// Use default font and color for all lines
	tinyfont.WriteLine(&o.dev, &proggy.TinySZ8pt7b, x, y, text, color.RGBA{255, 255, 255, 255})
}

// --- Setup function ---

type IO struct {
	// Inputs
	K1, K2 IKnob
	B1, B2 IButton
	DIN    IDigitalInput
	AIN    IAnalogueInput
	// Outputs
	CV1, CV2, CV3, CV4, CV5, CV6 ICV
	Display                      IOledDevice
}

func SetupEuroPi() *IO {
	// CRITICAL: Initialize ADC hardware before any ADC use
	machine.InitADC()

	euroPiIO := &IO{
		K1:      NewKnob(machine.ADC1),
		K2:      NewKnob(machine.ADC2),
		B1:      NewButton(machine.GPIO4),
		B2:      NewButton(machine.GPIO5),
		DIN:     NewDigitalInput(machine.GPIO22, true), // true: active low
		AIN:     NewAnalogueInput(machine.ADC0),
		CV1:     &CV{Index: 1},
		CV2:     &CV{Index: 2},
		CV3:     &CV{Index: 3},
		CV4:     &CV{Index: 4},
		CV5:     &CV{Index: 5},
		CV6:     &CV{Index: 6},
		Display: InitDisplay(),
	}

	// Configure CV output pins
	ConfigureCV()

	return euroPiIO
}

// ----------- Application 1 Starts Here -----------

type Diagnostic struct{}

func (c Diagnostic) Name() string { return "Diagnostic Tester" }

// Holds the current state of the display - its an optional optimization
// to avoid unnecessary display updates if the state hasn't changed, to avoid flickering.
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

func (c Diagnostic) Run(io *IO) {
	displayState := &DiagnosticDisplayState{}
	const CV_STEP = 10     // Update CV's more slowly than main loop
	const PRINT_STEP = 100 // Print every 100 iterations (~1s at 10ms loop)
	var cvLoopCount int = CV_STEP

	for loopCount := 0; ; loopCount++ {
		knob1Value := io.K1.Value()
		knob2Value := io.K2.Value()
		ainVolts := io.AIN.Volts()
		ainValue := io.AIN.Value()
		btn1Pressed := io.B1.Pressed()
		btn2Pressed := io.B2.Pressed()
		dinState := io.DIN.Get()

		// if both buttons are pressed and held for 2s, exit the loop
		if ShouldExit(io) {
			break
		}

		if loopCount%PRINT_STEP == 0 { // Print every ~1s
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

		// Output PWM - update all CV outputs with random values
		cvLoopCount--
		if cvLoopCount <= 0 {
			cvLoopCount = CV_STEP // Reset counter
			for cv := 1; cv <= 6; cv++ {
				duty := uint32(rand.Intn(int(MaxDuty))) // 0–maxDuty
				SetCV(cv, duty)
			}
		}

		// OLED update
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

// DoubleButtonPressState holds state for non-blocking double button detection
var doubleButtonPressLastMs int64 = 0

// DoubleButtonPress returns true if both B1 and B2 are pressed and held for 2s (2000ms), non-blocking version.
// Call this repeatedly in your loop. It tracks state across calls.
func ShouldExit(io *IO) bool {
	now := time.Now().UnixMilli()
	if io.B1.Pressed() && io.B2.Pressed() {
		if doubleButtonPressLastMs == 0 {
			doubleButtonPressLastMs = now
		} else if now-doubleButtonPressLastMs >= 2000 {
			doubleButtonPressLastMs = 0 // reset for next time
			return true
		}
	} else {
		doubleButtonPressLastMs = 0
	}
	return false
}

// ----------- Application 2 Starts Here -----------

type HelloWorld struct{}

func (c HelloWorld) Name() string { return "Hello World" }

func (c HelloWorld) Run(io *IO) {
	println("Hello, World!")
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "Hello, World!")
	io.Display.Display()

	for {
		// if both buttons are pressed and held for 2s, exit the loop
		if ShouldExit(io) {
			break
		}
	}
	println("Exiting HelloWorld application.")
	io.Display.ClearDisplay()
	io.Display.Display() // Clear the display
	println("Display cleared. Goodbye!")
	time.Sleep(1 * time.Second) // Give time to see the goodbye message
}

// App interface for all plugins
// Name returns the display name, Run runs the app logic and must respect exit condition
// Run(io *IO) is required for hardware access

type App interface {
	Name() string
	Run(io *IO)
}

// App registry
var appRegistry []App

// RegisterApp adds an app to the registry
func RegisterApp(app App) {
	appRegistry = append(appRegistry, app)
}

// Register plugins on startup
func init() {
	RegisterApp(Diagnostic{})
	RegisterApp(HelloWorld{})
	RegisterApp(HelloWorld{})
	RegisterApp(Diagnostic{})
	RegisterApp(HelloWorld{})
}

// Holds the current state of the menu display to avoid unnecessary redraws and flicker.
type MenuDisplayState struct {
	Lines [3]string
}

// IsDirty returns true if any of the menu lines have changed.
func (m *MenuDisplayState) IsDirty(lines [3]string) bool {
	for i := 0; i < 3; i++ {
		if m.Lines[i] != lines[i] {
			return true
		}
	}
	return false
}

// Update sets the current menu lines to the new values.
func (m *MenuDisplayState) Update(lines [3]string) {
	copy(m.Lines[:], lines[:])
}

// MenuChooser displays a scrollable menu of registered apps, allows selection with K2, launch with B2
func MenuChooser(io *IO) int {
	numApps := len(appRegistry)
	if numApps == 0 {
		return -1
	}
	selected := 0
	lastK2 := io.K2.Value()
	const debounceMs = 150
	lastAction := time.Now()

	displayState := &MenuDisplayState{} // Menu display state

	for {
		// Read K2 knob for up/down
		k2 := io.K2.Value()
		if abs(k2-lastK2) > 2 && time.Since(lastAction) > debounceMs*time.Millisecond {
			if k2 > lastK2 {
				selected++
			} else if k2 < lastK2 {
				selected--
			}
			if selected < 0 {
				selected = 0
			}
			if selected >= numApps {
				selected = numApps - 1
			}
			lastK2 = k2
			lastAction = time.Now()
		}

		// Draw menu (3 lines)
		start := selected - 1
		if start < 0 {
			start = 0
		}
		if start > numApps-3 {
			start = numApps - 3
		}
		if start < 0 {
			start = 0
		}
		var lines [3]string // New lines to display
		for i := 0; i < 3 && (start+i) < numApps; i++ {
			idx := start + i
			name := appRegistry[idx].Name()
			line := name
			if idx == selected {
				line += " *"
			}
			lines[i] = line
		}

		// Only update display if lines have changed
		if displayState.IsDirty(lines) {
			io.Display.ClearDisplay()
			for i := 0; i < 3; i++ {
				if lines[i] != "" {
					io.Display.WriteLine(0, int16(10+10*i), lines[i])
				}
			}
			io.Display.Display()
			displayState.Update(lines)
		}

		// Launch app on B2 press
		if io.B2.Pressed() && !io.B1.Pressed() { // B1 must not be pressed
			// Debounce: wait for release
			for io.B2.Pressed() {
				time.Sleep(10 * time.Millisecond)
			}
			return selected
		}

		// Allow exit with double button press
		if ShouldExit(io) {
			return -1
		}

		time.Sleep(30 * time.Millisecond)
	}
}

func splashScreen(io *IO) {
	// Display a splash screen for 2 seconds
	io.Display.ClearDisplay()
	io.Display.WriteLine(0, 10, "EuroPi Simplified")
	io.Display.WriteLine(0, 20, "by TinyGo "+version)
	io.Display.Display()
	time.Sleep(2 * time.Second)
	io.Display.ClearDisplay()
}

// --- Mock implementations for testing ---

type MockKnob struct{ val int }

func (m *MockKnob) Value() int { return m.val }

// Set value for test
func (m *MockKnob) SetValue(v int) { m.val = v }

type MockButton struct{ pressed bool }

func (m *MockButton) Pressed() bool     { return m.pressed }
func (m *MockButton) SetPressed(p bool) { m.pressed = p }

type MockDigitalInput struct{ state bool }

func (m *MockDigitalInput) Get() bool       { return m.state }
func (m *MockDigitalInput) SetState(s bool) { m.state = s }

type MockAnalogueInput struct {
	volts float64
	value int
}

func (m *MockAnalogueInput) Volts() float64     { return m.volts }
func (m *MockAnalogueInput) Value() int         { return m.value }
func (m *MockAnalogueInput) SetVolts(v float64) { m.volts = v }
func (m *MockAnalogueInput) SetValue(val int)   { m.value = val }

type MockCV struct {
	val     uint32
	on, off bool
}

func (m *MockCV) Set(v uint32) { m.val = v }
func (m *MockCV) On()          { m.on = true }
func (m *MockCV) Off()         { m.off = true }

// Simple mock OLED device that just stores lines written

type MockOledDevice struct {
	lines     []string
	cleared   bool
	displayed bool
}

func (m *MockOledDevice) ClearDisplay()                     { m.cleared = true; m.lines = nil }
func (m *MockOledDevice) Display()                          { m.displayed = true }
func (m *MockOledDevice) WriteLine(x, y int16, text string) { m.lines = append(m.lines, text) }

// SetupEuroPiMock returns an IO struct with all fields set to mocks
func SetupEuroPiMock() *IO {
	return &IO{
		K1:      &MockKnob{},
		K2:      &MockKnob{},
		B1:      &MockButton{},
		B2:      &MockButton{},
		DIN:     &MockDigitalInput{},
		AIN:     &MockAnalogueInput{},
		CV1:     &MockCV{},
		CV2:     &MockCV{},
		CV3:     &MockCV{},
		CV4:     &MockCV{},
		CV5:     &MockCV{},
		CV6:     &MockCV{},
		Display: &MockOledDevice{},
	}
}

func main() {
	time.Sleep(1 * time.Second) // Allow time for the debug monitor to get ready
	println("Starting...")

	var io *IO
	if production {
		io = SetupEuroPi()
		println("EuroPi configured (production mode).")
	} else {
		io = SetupEuroPiMock()
		println("EuroPi configured (mock mode).")
	}

	// Display Splash Screen
	splashScreen(io)
	println("Entering main menu loop. Press B2 to select an app, K2 to scroll.")

	// Main menu loop
	// This loop will run until the user selects an app or exits with double button press
	// It will display the splash screen at the start and after each app run.
	for {
		idx := MenuChooser(io)
		if idx < 0 || idx >= len(appRegistry) {
			println("Exiting main menu loop.")
			break
		}
		println("Launching app:", appRegistry[idx].Name())
		appRegistry[idx].Run(io)
		println(appRegistry[idx].Name(), "completed. Returning to menu...")

		// Display Splash Screen
		splashScreen(io)
	}
}
