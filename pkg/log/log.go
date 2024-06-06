
package log

import (
    "fmt"
    "bufio"
    "time"
    "io"

    "bh/lastlog/pkg/parser"
)

type HasTime interface {
    GetTime() time.Time
}

type L[T HasTime] struct {
    Input io.Reader
    Parser *parser.P[T]
    Last time.Time
}

// Parse() reads 'L.Input' reader line by line calling 'L.Parser' on each
// line and resending its results to returned channel. Line parsers are run
// synchronously. Timestamp returned by last parser is saved in 'L.Last'.
func (l *L[T]) Parse() <-chan T {
    items := make(chan T)

    go func() {
        defer close(items)

        scanner := bufio.NewScanner(l.Input)
        for scanner.Scan() {
            fmt.Printf("log.L.Parse(): Read '%s'\n", scanner.Text())
            for d := range l.Parser.Run(scanner.Text()) {
                // FIXME: Should i check, that new last time is after previous
                // one? And if it is, what to do? Discard result?
                l.Last = d.GetTime()
                items <- d
            }
        }

        if err := scanner.Err(); err != nil {
            panic(fmt.Sprintf("log.L.Parse(): Scanner error %v\n", err))
        }
    }()

    return items
}
