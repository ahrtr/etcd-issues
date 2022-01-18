package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
)

func allGoroutines(excludeSelf bool) (gs []string) {
	buf := make([]byte, 2<<20)

	pc, fileName, line, ok := runtime.Caller(0)
	funcName := runtime.FuncForPC(pc).Name()
	fmt.Printf("name: %s, file: %s, line: %d, ok: %t\n\n", funcName, fileName, line, ok)

	buf = buf[:runtime.Stack(buf, true)]
	for _, g := range strings.Split(string(buf), "\n\n") {
		/*sl := strings.SplitN(g, "\n", 2)
		if len(sl) != 2 {
			continue
		}
		stack := strings.TrimSpace(sl[1])
		gs = append(gs, stack) */

		if excludeSelf && ok {
			if strings.Contains(g, funcName) && strings.Contains(g, fileName) {
				continue
			}
		}
		gs = append(gs, g)
	}
	sort.Strings(gs)
	return gs
}

func routine1() {
	select {
	case <-time.After(10 * time.Second):
	}
}

func routine2() {
	select {
	case <-time.After(10 * time.Second):
	}
}

func Demo1() {
	go routine1()
}

func main() {
	Demo1()
	go routine1()
	go routine2()

	gs := allGoroutines(false)
	fmt.Printf("Goroutine count: %d\n\n", len(gs))
	for i, g := range gs {
		fmt.Printf("### %d:\n", i)
		fmt.Println(g)
		fmt.Println()
	}
}
