package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	flag "github.com/spf13/pflag"
	rpio "github.com/stianeikeland/go-rpio/v4"
)

var version = "dev"

func usage() {
	fmt.Fprintf(os.Stderr, "usage: gpiotest {pin} [options]\n\noptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	versionFlag := flag.BoolP("version", "v", false, "print the version and exit")
	pinMode := flag.StringP("mode", "m", "out", "set mode to 'out' or 'pwm'")
	pinOn := flag.BoolP("high", "H", false, "if set and --out, will turn the pin on (high)")
	pinOff := flag.BoolP("low", "L", false, "if set and --out, will turn the pin off (low)")
	var sleepTime time.Duration
	var autoMode bool
	flag.DurationVar(&sleepTime, "sleep", time.Duration(500)*time.Millisecond, "set the sleep time")
	flag.BoolVarP(&autoMode, "auto", "a", false, "if set, pwm will increment every --sleep duration, if not, user interaction is required")
	rf := flag.Uint32("rf", 2000, "set the frequency, default to 2KHz")
	var duty uint32
	flag.Uint32Var(&duty, "duty", 50, "set the duty, default to 50")
	cycle := flag.Uint32("cycle", 100, "set the cycle, default to 100")
	flag.Parse()

	if *versionFlag {
		fmt.Fprintf(os.Stdout, "gpiotest %s\n", version)
		os.Exit(0)
	}

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "fatal: missing required pin argument\n")
		usage()
		os.Exit(1)
	}

	uid := os.Getuid()
	if uid != 0 {
		fmt.Fprintf(os.Stderr, "fatal: this program needs to run as root (sudo)\n")
		os.Exit(1)
	}

	pinStr := flag.Arg(0)
	pinNum, err := strconv.Atoi(pinStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: pin was not an integer, was: %s\n", pinStr)
		os.Exit(1)
	}

	err = rpio.Open()
	if err != nil {
		panic(fmt.Sprint("unable to open gpio", err.Error()))
	}

	pin := rpio.Pin(pinNum)
	if *pinMode == "out" {
		pin.Output()
		if *pinOn && !*pinOff {
			pin.High()
			fmt.Printf("Set pin %d HIGH.\n", pinNum)
		} else if *pinOff && !*pinOn {
			pin.Low()
			fmt.Printf("Set pin %d LOW.\n", pinNum)
		} else {
			fmt.Println("You must specify only --on or --off with -m out")
			os.Exit(1)
		}
		defer rpio.Close()
	} else if *pinMode == "pwm" {

		pin.Output()
		pin.Pwm()
		actualRf := (*rf) * (*cycle)
		pin.DutyCycle(duty, *cycle)
		pin.Freq(int(actualRf))
		rpio.StartPwm()
		fmt.Printf("Set pin %d to %d Hz (%d actual) with a Duty Cycle of %d / %d.\n", pinNum, *rf, actualRf, duty, *cycle)
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
		go func() {
			<-signalCh
			rpio.StopPwm()
			rpio.Close()
			os.Exit(0)
		}()
		reader := bufio.NewReader(os.Stdin)

		for {
			if autoMode {
				fmt.Printf("Waiting %s...\n", sleepTime)
				select {
				case <-time.After(sleepTime):
				}
			} else {
				fmt.Println("Press enter to continue...")
				reader.ReadString('\n')
			}
			duty++
			if duty > *cycle {
				duty = 0
			}
			pin.DutyCycle(duty, *cycle)
			fmt.Printf("Duty set to %d / %d\n", duty, *cycle)
		}

	} else {
		fmt.Printf("Invalid pin mode '%s', only 'pwm' or 'out' are valid.\n", *pinMode)
		os.Exit(1)
	}
}
