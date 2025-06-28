# EuroPi-Go

An alternative implementation of EuroPi in Go, using TinyGo for embedded systems.

# Run

To run the examples, you can use the following command:

```bash
`tinygo flash -target=pico --monitor ./cmd/pico`
```

To run the mock version, which simulates the EuroPi hardware without needing the actual device, use:

```bash
go run ./cmd/mock
```

For fancy bubbletea mock UI, you can run:

```bash
go run -tags tea ./cmd/mock -tea
```

Note all logging will be logged to file `mock.log` in the project root directory.

