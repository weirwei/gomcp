package gomcp

import (
	"log"
	"runtime/debug"
)

// Safe mode, recover the panic, prevent crash the server.
func Safe(f func()) func() {
	return func() {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				log.Printf("panic: %v, stack: %s", rec, stack)
			}
		}()
		f()
	}
}
