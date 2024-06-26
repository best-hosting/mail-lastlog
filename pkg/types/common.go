
package types

import (
    "log"
    "os"
    "runtime"
    "fmt"
    "strings"
    "path/filepath"
)

// caller is reimplementation of 'runtime.Caller()' returning caller's
// function name as well.
func caller(skip int) (file string, line int, fun string, ok bool) {
    // Only one element: just the name of caller function,
    pc := make([]uintptr, 1)
    // Skip 2 more frames: 'Callers()' itself and 'caller()'.
    n  := runtime.Callers(skip + 2, pc)
    if n == 0 {
        ok = false
        return
    }

    // This is essentially 'runtime.Caller()', but 'Caller' strangely does
    // not return caller's function name.
    fr, _ := runtime.CallersFrames(pc[:n]).Next()
    file = filepath.Base(fr.File)
    line = fr.Line

    // Remove filepath from function name.
    fun = strings.TrimPrefix(filepath.Base(fr.Function), strings.TrimSuffix(file, "go"))
    // Remove generics marker.
    if i := strings.Index(fun, "[...]"); i > 0 {
        fun = fun[:i] + fun[i + len(`[...]`):]
    }

    ok = true
    return
}

// FIXME: It seems, go runtime panic messages are not output correctly, when i
// use logger to stderr.
// FIXME: It seems, journalctl omits some output from go threads on error.
// Probably, this is due to buffering and may be avoided by using Sync() on
// (os.File*), but logger is not os.File. Thus, i may implement interface like
// WriterFlusher, which provides flush method on underlying object and it may
// be called from 'LogErr()' and 'LogFatal()' functions.
var Logger *log.Logger = log.New(os.Stderr, "[mail-lastlog] ", log.Ldate | log.Ltime)

func CallerInfo(skip int) string {
    f, l, fn, ok := caller(skip + 1)
    if !ok {
        return ""
    }
    return fmt.Sprintf("%v:%v@%v()", f, l, fn)
}

// logPrefix solves two problems of just using logger's methods directly
// ('Printf' and others):
// - obtain (proper) caller function name.
// - obtain proper line number (i.e. the equivalent of calling logger's Printf
// directly).
//
// It takes callback function, to which correct prefix will be passed to.
func logPrefix(skip int, cb func (string)) {
    f, l, fn, ok := caller(skip + 1)
    if !ok {
        cb("")
        return
    }
    cb(fmt.Sprintf("%v:%v | %v(): ", f, l, fn))
}

func Logf(format string, vs ...interface{}) {
    f := func(pref string) {
        Logger.Print(pref + fmt.Sprintf(format, vs...))
    }
    // Skip one frame in call stack: this function.
    logPrefix(1, f)
}

func Logfn(format string, vs ...interface{}) {
    f := func(pref string) {
        Logger.Print(pref + fmt.Sprintf(format + "\n", vs...))
    }
    logPrefix(1, f)
}

// logErr2 skips spcified number of frames and logs error. Error may be either
// provided as 'error' and in that case there should be no more arguments, or
// as format string with arguments for fmt.Errorf().
func logErr2(skip int, errFmt interface{}, vs ...interface{}) (err error) {
    var f func(string)
    switch v := errFmt.(type) {
    case error:
        err = v
        if len(vs) > 0 {
            err = fmt.Errorf("Extra arguments in call to %v at %v with error '%v': %v\n", CallerInfo(1), CallerInfo(2), err, vs)
        }
    case string:
        err = fmt.Errorf(v, vs...)
    default:
        err = fmt.Errorf("Unknown type of 1st argument in call to %v: %T\n", CallerInfo(1), v)
    }

    f = func(pref string) {
        Logger.Print(pref + "Error: " + fmt.Sprintln(err))
    }
    logPrefix(skip + 1, f)
    return err
}

// LogErr logs error. Error may be either provided as 'error' and in that case
// there should be no more arguments, or as format string with arguments for
// fmt.Errorf().
func LogErr(errFmt interface{}, vs ...interface{}) error { return logErr2(1, errFmt, vs...) }

// LogFatal logs error and exits with 1. See LogErr.
func LogFatal(errFmt interface{}, vs ...interface{}) {
    logErr2(1, errFmt, vs...)
    os.Exit(1)
}
