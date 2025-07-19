# Pixel Animation App

An app that draw sine waves, square waves, random lines and rectangles on the display.

This app is designed to demonstrate the capabilities of the Europi hardware in rendering simple animations using pixel manipulation. It also demonstrates the use of the Europi-Go controls and display libraries, as well as the integration of scheduled tasks for animation updates. It also shows how to use the SSD1306 display driver for rendering graphics. And the use of Go concurrency features to handle animations efficiently.

There are three variations of the pixel animation app:

1. **Pixels Simple**: A basic implementation that draws simple shapes and animations on the display.
2. **Pixels Scheduled**: An implementation that uses a scheduled task to update the display at regular intervals, allowing for more complex animations.
3. **Pixels Scheduled FastIO**: An optimized version that uses fast I/O operations to improve the performance of the animations, making it suitable for more demanding applications.

# Usage

Just have the `cmd/pico/main.go` file import the `pixel_animation` package and register the desired app(s) with the firmware. For example:

```go
import (
    "europi/apps/pixel_animation"
)

func main() {
    // Other initialization code...

    firmware.RegisterApp(pixel_animation.Pixels{})          // For Pixels Simple
    firmware.RegisterApp(pixel_animation.Pixels2{})         // For Pixels Scheduled
    firmware.RegisterApp(pixel_animation.Pixels3{})         // For Pixels Scheduled FastIO

    // Start the firmware...
}
```

This will allow you to run the pixel animation apps on your Europi device. Each app can be selected from the main menu, and they will render different animations on the display.
