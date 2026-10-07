// Spell: gpio-control (wild school) — Argos controls hardware peripherals.
// Uses sysfs, libgpiod, and direct device access for GPIO, SPI, I2C, PWM, ADC.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Input struct {
	Path string            `json:"path"`
	Args map[string]string `json:"args"`
}

type Output struct {
	Summary string `json:"summary"`
	Error   string `json:"error,omitempty"`
}

func main() {
	var in Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("reading input: %v", err)})
		return
	}
	action := in.Args["action"]
	pin := in.Args["pin"]
	value := in.Args["value"]

	switch action {
	case "gpio":
		mode := in.Args["mode"]
		if mode == "" { mode = "out" }
		// Use /sys/class/gpio if available, else gpioset
		gpioPath := fmt.Sprintf("/sys/class/gpio/gpio%s", pin)
		if _, err := os.Stat(gpioPath); os.IsNotExist(err) {
			os.WriteFile("/sys/class/gpio/export", []byte(pin), 0644)
		}
		os.WriteFile(gpioPath+"/direction", []byte(mode+"\n"), 0644)
		if value != "" {
			os.WriteFile(gpioPath+"/value", []byte(value+"\n"), 0644)
		}
		val, _ := os.ReadFile(gpioPath + "/value")
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("GPIO%s mode=%s value=%s", pin, mode, strings.TrimSpace(string(val)))})

	case "spi":
		device := in.Args["device"]
		if device == "" { device = "/dev/spidev0.0" }
		data := in.Args["data"]
		cmd := exec.Command("python3", "-c", fmt.Sprintf(`
import spidev
spi = spidev.SpiDev()
spi.open(0, 0)
spi.max_speed_hz = %s
resp = spi.xfer2([%s])
spi.close()
print(" ".join(hex(b) for b in resp))
`, valOr(in.Args, "speed", "1000000"), data))
		spiOut, spiErr := cmd.CombinedOutput()
		if spiErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", spiErr, string(spiOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("SPI %s: %s", device, strings.TrimSpace(string(spiOut)))})

	case "i2c":
		bus := in.Args["bus"]
		if bus == "" { bus = "1" }
		addr := in.Args["addr"]
		reg := in.Args["reg"]
		i2cCmd := exec.Command("i2cget", "-y", bus, addr, reg)
		i2cOut, i2cErr := i2cCmd.CombinedOutput()
		if i2cErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", i2cErr, string(i2cOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("I2C bus=%s addr=%s reg=%s: %s", bus, addr, reg, strings.TrimSpace(string(i2cOut)))})

	case "pwm":
		chip := in.Args["chip"]
		if chip == "" { chip = "0" }
		channel := in.Args["channel"]
		if channel == "" { channel = "0" }
		period := in.Args["period"]
		duty := in.Args["duty"]
		pwmPath := fmt.Sprintf("/sys/class/pwm/pwmchip%s/pwm%s", chip, channel)
		if _, err := os.Stat(pwmPath); os.IsNotExist(err) {
			os.WriteFile(fmt.Sprintf("/sys/class/pwm/pwmchip%s/export", chip), []byte(channel+"\n"), 0644)
		}
		if period != "" { os.WriteFile(pwmPath+"/period", []byte(period+"\n"), 0644) }
		if duty != "" { os.WriteFile(pwmPath+"/duty_cycle", []byte(duty+"\n"), 0644) }
		os.WriteFile(pwmPath+"/enable", []byte("1\n"), 0644)
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("PWM chip=%s ch=%s period=%s duty=%s", chip, channel, period, duty)})

	case "adc":
		channel := in.Args["channel"]
		if channel == "" { channel = "0" }
		adcPath := fmt.Sprintf("/sys/bus/iio/devices/iio:device0/in_voltage%s_raw", channel)
		val, err := os.ReadFile(adcPath)
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: err.Error()})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("ADC ch%s: %s", channel, strings.TrimSpace(string(val)))})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown action: %s", action)})
	}
}

func valOr(args map[string]string, key, def string) string {
	if v, ok := args[key]; ok && v != "" { return v }
	return def
}
