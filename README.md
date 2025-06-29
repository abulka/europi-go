# EuroPi-Go

An alternative implementation of a EuroPi firmware in Go, using TinyGo for embedded systems.

# Installation
To get started with EuroPi-Go, you need to install [TinyGo](https://tinygo.org/getting-started/).

Download this repository and navigate to the project directory and run `go mod tidy` to ensure all dependencies are installed.

# Run

To flash the firmware and apps to the EuroPi, you can use the following command:

```bash
`tinygo flash -target=pico --monitor ./cmd/pico`
```

Or just build the firmware without flashing:

```bash
tinygo build -target=pico ./cmd/pico
```

## Mock Version

To run the mock version, which simulates the EuroPi hardware without needing the actual device, use:

```bash
go run ./cmd/mock
```

For fancy bubbletea mock UI, you can run:

```bash
go run -tags tea ./cmd/mock -tea
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

