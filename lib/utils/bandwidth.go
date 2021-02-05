package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strconv"
	"time"
)

var ifaceName string
var txBytesPath string
var rxBytesPath string

func init() {
	// TODO: change when on real system, locally using lo
	ifaceName = "lo"
	txBytesPath = fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", ifaceName)
	rxBytesPath = fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", ifaceName)
}

type Measurement struct {
	Time time.Time
	Tx   int64
	Rx   int64
}

func CalcBW(m1, m2 Measurement) (string, string) {
	dur := m2.Time.Sub(m1.Time)
	sent := float64(m2.Tx - m1.Tx)
	recv := float64(m2.Rx - m1.Rx)
	tx := (sent / 1e6) / dur.Seconds()
	rx := (recv / 1e6) / dur.Seconds()
	return fmt.Sprintf("%.1f", tx), fmt.Sprintf("%.1f", rx)
}

func Measure() Measurement {
	// TODO: solve for osx
	if runtime.GOOS != "linux" {
		return Measurement{}
	}

	t := time.Now()
	tx, err := ioutil.ReadFile(txBytesPath)
	if err != nil {
		log.Fatalf("impossible to read txBytes file")
	}
	rx, err := ioutil.ReadFile(rxBytesPath)
	if err != nil {
		log.Fatalf("impossible to read rxBytes file")
	}
	tx = bytes.TrimSpace(tx)
	rx = bytes.TrimSpace(rx)
	itx, err := strconv.ParseInt(string(tx), 10, 64)
	if err != nil {
		log.Fatalf("impossible to parse tx value")
	}
	irx, err := strconv.ParseInt(string(rx), 10, 64)
	if err != nil {
		log.Fatalf("impossible to parse rx value")
	}

	return Measurement{
		Time: t,
		Tx:   itx,
		Rx:   irx,
	}
}

func LogBWLoop() {
	last := Measure()
	for {
		m := Measure()
		tx, rx := CalcBW(last, m)
		log.Printf("TxRx: net=%s  t=%d  tx=%d  rx=%d -- since last: (%s Mbit/s, %s Mbit/s)", ifaceName, m.Time.Unix(), m.Tx, m.Rx, tx, rx)
		last = m
		time.Sleep(10 * time.Second)
	}
}
