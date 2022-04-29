package util

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

// Retry runs with attempts
func Retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep)
			sleep *= 2
		}
		err = f()
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

// Until runs until signal received. Handle panics
func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	select {
	case <-stopCh:
		return
	default:
	}
	for {
		func() {
			defer handleCrash()
			f()
		}()
		select {
		case <-stopCh:
			return
		case <-time.After(period):
		}
	}
}

// HandleCrash simply catches a crash and logs an error. Meant to be called via defer.
func handleCrash() {
	if r := recover(); r != nil {
		callers := string(debug.Stack())
		log.Printf("recovered from panic: %#v (%v)\n%v", r, r, callers)
	}
}
