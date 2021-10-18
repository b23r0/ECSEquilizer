package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Protocol ...
type Protocol int

const (
	// TCP is tcp protocol
	TCP Protocol = iota
	// HTTP is http protocol
	HTTP
	// HTTPS is https protocol
	HTTPS
)

func (protocol Protocol) String() string {
	switch protocol {
	case TCP:
		return "tcp"
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	}
	return "unkown"
}

// Result ...
type Result struct {
	Counter        int
	SuccessCounter int
	Target         *Target

	MinDuration   time.Duration
	MaxDuration   time.Duration
	TotalDuration time.Duration
}

// Target is a ping
type Target struct {
	Protocol Protocol
	Host     string
	Port     int
	Proxy    string

	Counter  int
	Interval time.Duration
	Timeout  time.Duration
}

// Pinger is a ping interface
type Pinger interface {
	Start() <-chan struct{}
	Stop()
	Result() *Result
	SetTarget(target *Target)
}

// TCPing ...
type TCPing struct {
	target *Target
	done   chan struct{}
	result *Result
}

var _ Pinger = (*TCPing)(nil)

// NewTCPing return a new TCPing
func NewTCPing() *TCPing {
	tcping := TCPing{
		done: make(chan struct{}),
	}
	return &tcping
}

// SetTarget set target for TCPing
func (tcping *TCPing) SetTarget(target *Target) {
	tcping.target = target
	if tcping.result == nil {
		tcping.result = &Result{Target: target}
	}
}

// Result return the result
func (tcping *TCPing) Result() *Result {
	return tcping.result
}

// Start a tcping
func (tcping *TCPing) Start() <-chan struct{} {
	go func() {
		t := time.NewTicker(tcping.target.Interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if tcping.result.Counter >= tcping.target.Counter && tcping.target.Counter != 0 {
					tcping.Stop()
					return
				}
				duration, remoteAddr, err := tcping.ping()
				tcping.result.Counter++

				if err != nil {
					log.Printf("Ping - failed: %s\n", err)
					log.Panicln(tcping.target)
				} else {
					log.Printf("Ping (%s:%d) - Connected - time=%s\n", remoteAddr, tcping.target.Port, duration)

					if tcping.result.MinDuration == 0 {
						tcping.result.MinDuration = duration
					}
					if tcping.result.MaxDuration == 0 {
						tcping.result.MaxDuration = duration
					}
					tcping.result.SuccessCounter++
					if duration > tcping.result.MaxDuration {
						tcping.result.MaxDuration = duration
					} else if duration < tcping.result.MinDuration {
						tcping.result.MinDuration = duration
					}
					tcping.result.TotalDuration += duration
				}
			case <-tcping.done:
				return
			}
		}
	}()
	return tcping.done
}

// Stop the tcping
func (tcping *TCPing) Stop() {
	tcping.done <- struct{}{}
}

func timeIt(f func() interface{}) (int64, interface{}) {
	startAt := time.Now()
	res := f()
	endAt := time.Now()
	return endAt.UnixNano() - startAt.UnixNano(), res
}

func (tcping *TCPing) ping() (time.Duration, net.Addr, error) {
	var remoteAddr net.Addr
	duration, errIfce := timeIt(func() interface{} {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", tcping.target.Host, tcping.target.Port), tcping.target.Timeout)
		if err != nil {
			return err
		}
		remoteAddr = conn.RemoteAddr()
		conn.Close()
		return nil
	})
	if errIfce != nil {
		err := errIfce.(error)
		return 0, remoteAddr, err
	}
	return time.Duration(duration), remoteAddr, nil
}
