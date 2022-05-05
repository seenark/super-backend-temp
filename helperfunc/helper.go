package helperfunc

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// helpers
func ToMilliseconds(t time.Time) int {
	return int(t.UnixNano()) / 1e6
}

func PrintPretty(data interface{}) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
}

func Trace() {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}
