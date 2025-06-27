# TinyGo and EuroPi Experiments

This repository contains various experiments and examples using TinyGo, particularly focused on the EuroPi platform. The code is structured to demonstrate different functionalities, including OLED display initialization and usage.

# Run

To run the examples, you can use the following command:

```bash
tinygo flash -target=pico ./research/blinky 
tinygo flash -target=pico --monitor ./research/dualcore
```

# OLED Display Examples
The OLED display examples demonstrate how to initialize and use an SSD1306 OLED display with TinyGo. The `oledinit.go` file contains the initialization logic, and the `oled.go` files in the `research` directory show how to use this initialization in different contexts.

The EuroPi `SSD1306` OLED has a 128x32 OLED display using I2C on pins GPIO 0 and GPIO 1, so its using I2C0. It uses address `0x3C`. Frequency is set to 400kHz.

```bash
tinygo flash -target=pico ./research/claude
tinygo flash -target=pico --monitor ./research/oled
tinygo flash -target=pico --monitor ./research/oled2
tinygo flash -target=pico --monitor ./research/oled3
```
