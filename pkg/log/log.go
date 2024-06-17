
package log

import (
    "fmt"
    "bufio"
    "io"
    "os"

    . "bh/lastlog/pkg/types"
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
// func (l *L[T, K]) Parse(upstream chan<- K) {
func (l *L[T, K]) Parse() <-chan K {
    upstream := make(chan K)
    filterIn := make(chan K)

    done := make(chan any)
    go func() {
        defer close(filterIn)

        scanner := bufio.NewScanner(l.Input)
        // 1. Read first line.
        // 2. Find oldest interval, where t.Before(start) is true
        // 3. Init end to 0
        // 4. If t.After(end) send result and update te.
        // 5. At each te update compare it with next tstart. If it's after,
        //    merge intervals.
        //  During initial interval selection and merge, check Mtime against
        //  te of choosed interval. If mtime is before, shortcicuit parsing.
        for scanner.Scan() {
            fmt.Printf("log.L.Parse(): Read '%s'\n", scanner.Text())
            // FIXME: Send 'filterIn' channel directly to parser instead of
            // resending its results here.
            ch := l.Parser.Run(scanner.Text())
            //l.Parser.Run(filterIn, scanner.Text())
            for d := range ch {
                select {
                case filterIn<- d:
                case <-done:
                    for range ch {
                    }
                    return
                }
            }
        }

        if err := scanner.Err(); err != nil {
            panic(fmt.Sprintf("log.L.Parse(): Scanner error %v\n", err))
        }
    }()

    go func() {
        defer close(upstream)

        //l.Intervals = intervals.FilterBy(l.Intervals, filterIn, upstream, l.ModTime)
        l.Intervals.Filter(upstream, filterIn, l.ModTime)
        close(done)
    }()

    return upstream
}

// TODO: Open gzip-ed files properly.
func OpenFile[K ToOrd[Time]](l *L[Time, K], file string) error {
    fi, err := os.Stat(file)
    if err != nil {
        return err
    }
    l.ModTime = Time{fi.ModTime()}

    f, err := os.Open(file)
    if err != nil {
        return err
    }
    l.Input = f

    return nil
}

// FIXME:
// func NewTailLog() ....

