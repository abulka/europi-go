// Mock display for host/testing
//go:build !tinygo

package display

type MockOledDevice struct {
	Lines     []string
	Cleared   bool
	Displayed bool
}

func (m *MockOledDevice) ClearDisplay()                     { m.Cleared = true; m.Lines = nil }
func (m *MockOledDevice) Display()                          { m.Displayed = true }
func (m *MockOledDevice) WriteLine(x, y int16, text string) { m.Lines = append(m.Lines, text) }
