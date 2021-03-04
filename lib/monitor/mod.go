package monitor

import (
	"log"
	"syscall"
)

// Helpers for measurement of CPU cost of operations
type Monitor struct {
	cpuTime float64
}

func NewMonitor() *Monitor {
	var m Monitor
	m.cpuTime = getCPUTime()
	return &m
}

func (m *Monitor) Reset() {
	m.cpuTime = getCPUTime()
}

func (m *Monitor) Record() float64 {
	return getCPUTime() - m.cpuTime
}

func (m *Monitor) RecordAndReset() float64 {
	old := m.cpuTime
	m.cpuTime = getCPUTime()
	return m.cpuTime - old
}

func (m *Monitor) GetCpuTime() float64 {
	return m.cpuTime
}

// Returns the sum of the system and the user CPU time used by the current process so far.
func getCPUTime() float64 {
	rusage := &syscall.Rusage{}
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, rusage); err != nil {
		log.Fatalln("Couldn't get rusage time:", err)
		return -1
	}
	s, u := rusage.Stime, rusage.Utime // system and user time
	return iiToMS(int64(s.Sec), int64(s.Usec)) + iiToMS(int64(u.Sec), int64(u.Usec))
}

// sec is in seconds, usec in microseconds
// Converts to milliseconds
func iiToMS(sec int64, usec int64) float64 {
	return float64(sec)*1000.0 + float64(usec)/1000.0
}

// Converts to seconds
func iiToS(sec int64, usec int64) float64 {
	return float64(sec) + float64(usec)/1000000.0
}
