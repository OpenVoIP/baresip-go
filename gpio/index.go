package gpio

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	gpiod "github.com/warthog618/gpiod"
)

//Blinker 灯信息
type Blinker struct {
	offset int //rpi.GPIO4
	chip   *gpiod.Chip
	line   *gpiod.Line
	value  chan int
}

//Close 释放资源
func (blinker *Blinker) Close() {
	blinker.chip.Close()
	defer func() {
		blinker.line.Reconfigure(gpiod.AsInput)
		blinker.line.Close()
	}()
}

//Create 点亮指定 gpio
func Create(offset int) (*Blinker, error) {
	var err error
	blinker := &Blinker{}

	blinker.chip, err = gpiod.NewChip("gpiochip0")
	if err != nil {
		log.Error("gpio new chip error", err)
		return nil, err
	}

	// init
	v := 0
	blinker.line, err = blinker.chip.RequestLine(offset, gpiod.AsOutput(v))
	if err != nil {
		log.Error("RequestLine chip error", err)
		return nil, err
	}
	log.Info("Set pin %d %s\n", offset, gpioStatus[v])
	go blinker.loop()
	return blinker, nil
}

func (blinker *Blinker) loop() {
	// capture exit signals to ensure pin is reverted to input on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	for {
		select {
		case value := <-blinker.value:
			blinker.line.SetValue(value)
			fmt.Printf("Set pin %d %s\n", blinker.offset, gpioStatus[value])
		case <-quit:
			blinker.Close()
		}
	}
}
