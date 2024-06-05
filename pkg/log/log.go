
package log

import (
    "fmt"
    "bufio"
    "time"
    "io"

    "bh/lastlog/pkg/parser"
)

type Log[T any] struct {
    Input io.Reader
    Parser *parser.P[T]
    Last time.Time
}

// Parse() reads 'Input' reader line by line sequentially, calls 'Parser'
// parser on each line and resends its results to channel. Line parsers are
// running synchronously.
func (l *Log[T]) Parse() <-chan T {
    items := make(chan T)

    go func() {
        defer close(items)

        scanner := bufio.NewScanner(l.Input)
        for scanner.Scan() {
            fmt.Printf("Read '%s'\n", scanner.Text())
            for d := range l.Parser.Run(scanner.Text()) {
                items <- d
            }
        }
        if err := scanner.Err(); err != nil {
            fmt.Printf("dovecot.Parse(): Scanner error %v\n", err)
        }
    }()

    return items
}
