
package log

import (
    "bufio"
    "io"
    "os"
    "path/filepath"
    "compress/gzip"

    . "bh/lastlog/pkg/common"
    "bh/lastlog/pkg/parser"
    "bh/lastlog/pkg/intervals"
)

type L[T Ord[T], K ToOrd[T]] struct {
    Input io.Reader
    ModTime T  // log file mtime
    Parser *parser.P[K]
    //Intervals []intervals.Interval[T]
    Intervals *intervals.Intervals[T, K]
}

// Parse() reads 'L.Input' reader line by line calling 'L.Parser' on each
// line and resending its results to returned channel. Line parsers are run
// synchronously.
func (l *L[T, K]) Parse(up chan<- K) error {
    filterIn := make(chan K)

    done := make(chan any)
    errCh := make(chan error)
    // Read and parse.
    go func() {
        var err error

        defer func() {
            if err != nil {
                errCh <- err
            }
            close(errCh)
        }()

        defer close(filterIn)

        scanner := bufio.NewScanner(l.Input)
        for scanner.Scan() {
            LogDfn("Read '%s'", scanner.Text())
            // To run several parsers concurrently, i need to use different
            // parser.P struct-s for them.
            l.Parser.Run(filterIn, scanner.Text())
            select {
            case <-done: return
            default:
            }
        }

        if err := scanner.Err(); err != nil {
            err = LogErr("Scanner error '%v'", err)
            return
        }
    }()

    l.Intervals.Filter(up, filterIn, l.ModTime)
    close(done) // Stop reading file further.
    // Consume all remaining already parsed tokens.
    for range filterIn {
    }
    err, ok := <-errCh
    if ok && err != nil {
        return err
    }

    return nil
}

func OpenFile[K ToOrd[Time]](l *L[Time, K], file string) error {
    var err error

    fi, err := os.Stat(file)
    if err != nil {
        return err
    }
    l.ModTime = Time{fi.ModTime()}

    f, err := os.Open(file)
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            f.Close()
        }
    }()

    if filepath.Ext(file) == ".gz" {
        z, err := gzip.NewReader(f)
        if err != nil {
            return err
        }
        l.Input = z
    } else {
        l.Input = f
    }

    return nil
}

// FIXME:
// func NewTailLog() ....

