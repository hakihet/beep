package main

import (
	"flag"
	"fmt"
	"time"
)

type Cha struct {
	Q, D chan struct{}
}

func NewCha() Cha {
	return Cha{make(chan struct{}), make(chan struct{})}
}

func closed(c chan struct{}) bool {
	select {
	case <-c:
		return true
	default:
	}
	return false
}

func main() {

	var boops = flag.Int("boops", 2000, "boop times")
	var beeps = flag.Int("beeps", 3000, "beep times")
	var booping = flag.Int("booping", 0, "boop throttle")
	var beeping = flag.Int("beeping", 0, "beep throttle")
	var bliping = flag.Int("bliping", 0, "blip throttle")
	var throttle = flag.Int("throttle", 0, "ledger throttle")
	var timeout = flag.Int("timeout", 30, "timeout")
	flag.Parse()

	fmt.Println("Timeout ", *timeout)

	boop := NewCha()
	beep := NewCha()
	blip := NewCha()
	qs, cs := []chan struct{}{boop.Q, beep.Q, blip.Q}, []chan struct{}{boop.D, beep.D, blip.D}
	timer := make(chan struct{})
	ledger := make(chan string, 10000)

	//boop......boop
	go func(q, d chan struct{}, l chan string, reps int) {
		defer close(d)
		for {
			if closed(q) {
				return
			}
			if reps <= 0 {
				break
			}
			l <- "boop..."
			time.Sleep(time.Duration(*booping) * time.Second)
			l <- "...boop"
			reps--
		}
	}(boop.Q, boop.D, ledger, *boops)

	//beep......beep
	go func(q, d chan struct{}, l chan string, reps int, t int) {
		for !closed(q) {
			if reps > 0 {
				l <- "beep..."
				time.Sleep(time.Duration(*beeping) * time.Second)
				l <- "...beep"
				reps--
				continue
			}
			break
		}
		close(d)
	}(beep.Q, beep.D, ledger, *beeps, *beeping)

	//...blip...
	go func(q, d chan struct{}, l chan string, t int) {
		for !closed(q) {
			time.Sleep(time.Duration(t) * time.Second)
			l <- "...blip..."
		}
		close(d)
	}(blip.Q, blip.D, ledger, *bliping)

	//timer
	go func(c chan struct{}, t int) {
		<-time.After(time.Duration(t) * time.Second)
		c <- struct{}{}
	}(timer, *timeout)

	//main
	{
	out:
		for {
			time.Sleep(time.Duration(*throttle) * time.Second)
			select {
			case v := <-ledger:
				fmt.Println(v)
			case <-timer:
				fmt.Println("TIME UP!")
				for i := range qs {
					close(qs[i])
				}
			default:
				for i := range cs {
					if !closed(cs[i]) {
						continue out
					}
				}
				if len(ledger) == 0 {
					break out
				}
			}
		}
		fmt.Println(len(ledger))
	}
}
