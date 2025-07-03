# EuroPi-Go

An alternative implementation of a EuroPi firmware in Go, using TinyGo for embedded systems.

# Installation
To get started with EuroPi-Go, you need to install [TinyGo](https://tinygo.org/getting-started/).

Download this repository and navigate to the project directory and run `go mod tidy` to ensure all dependencies are installed.

# Run

To flash the firmware and apps to the EuroPi, you can use the following commands.

> Note for tinyfont mode the `-tags tinyfont` build flag is required to build the firmware with the smaller font. This is not a runtime flag, but a build-time flag to include the TinyFont mode.

```bash
tinygo flash -target=pico --monitor ./cmd/pico
# or for TinyFont mode
tinygo flash -tags tinyfont -target=pico --monitor ./cmd/pico
```

Or just build the firmware without flashing:

```bash
tinygo build -target=pico ./cmd/pico
# or for TinyFont mode
tinygo build -tags tinyfont -target=pico ./cmd/pico
```

## Mock Version

To run the mock version, which simulates the EuroPi hardware without needing the actual device, use the following commands.

> Note that for mock mode, using the `-tinyfont` flag is not a build flag but a runtime command line flag to switch to TinyFont mode, which uses a smaller font for display. There is no build-time flag for mock mode, as it is designed to run on any system that supports Go.

```bash
go run ./cmd/mock
go run ./cmd/mock -tinyfont
```

For fancy bubbletea mock UI, you can run:

```bash
go run ./cmd/mock -tea
go run ./cmd/mock -tea -tinyfont
```

Note all logging will be logged to file `mock.log` in the project root directory.

Example mock output:

```
┌─────────────────────────┐
│Knob1: 0    DIN:         │
│Knob2: 0    AIN:0.00v    │
│B1:Down B2:Down          │
└─────────────────────────┘
```

