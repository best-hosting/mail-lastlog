//go:build debug

package common

import (
    "fmt"
)

func LogDfn(format string, vs ...interface{}) {
    f := func(pref string) {
        Logger.Print(pref + fmt.Sprintf(format + "\n", vs...))
    }
    // Skip one frame in call stack: this function.
    logPrefix(1, f)
}
