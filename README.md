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

> If flashing fails, you may need to reset the EuroPi. Unplug the USB cable then press and hold the reset button on the device whilst re-connecting the USB cable. Then run the flash command again

To build the firmware without flashing:

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

# Developing

Set `.vscode/settings.json` to use the `tinygo` build tag for development:

```
{
  "go.buildTags": "tinygo"
}
```

When actively developing the firmware, so that the TinyGo build tags are set correctly. Tinygo compiler will ALWAYS use the `tinygo` build tag anyway, so its not really necessary - but it makes vscode behave better re intellisense and code navigation.

You will get squiggly lines in vscode for some lines in the other mode (tinygo or !tinygo), but they are not errors, just warnings that the code is not compatible with the standard Go compiler. You can ignore them.

When you are in mock mode including the running of tests, you should disable the `tinygo` build tag in `.vscode/settings.json`, something like this:

```json
{
  "go.buildTagsOFFLINE": ""
}

Or just remove the line with `go.buildTags` completely, as it is not needed for mock mode.

# Testing

To run tests, you can use the following command:

```bash
go test ./...
```

or for just the display package:

```bash
go test ./display
```

You can also use the Test Explorer in VSCode to run tests interactively.  Remember to disable the `tinygo` build tag in `.vscode/settings.json` if you are running tests (see discussion above).

