
package parser

import (
    "fmt"
    "regexp"
)

// Parser function.
type Fn[T any] func (*P[T]) Fn[T]

type P[T any] struct {
    rx *regexp.Regexp
    fn Fn[T]

    Match []string  // Regexp submatches.
    Data T          // Current parsed result.

    items chan T
}

// NewP() creates new parser.
func NewP[T any](rx string, fn Fn[T]) *P[T] {
    p := P[T] {
        rx: regexp.MustCompile(rx),
        fn: fn,
    }
    return &p
}

// Emit() sends current parsed data to results channel.
func (p *P[T]) Emit() {
    fmt.Printf("parser.P.Emit(): Emit %#v\n", p.Data)
    p.items<- p.Data
}

// NextN() advances 'P.Match' on k positions forward.
func (p *P[T]) NextN(k int, f Fn[T]) Fn[T] {
    if k < 1 {
        panic(fmt.Sprintf("parser.P.NextN(): Error: Can't advance on less, than 1 submatch: %v", k))
    } else if len(p.Match) <= k {
        panic(fmt.Sprintf("parser.P.NextN(): Too few submatches: advance on %v requested, but i have only %v (%#v)", k, len(p.Match), p.Match))
    }
    p.Match = p.Match[k:]
    return f
}

// Next() advances 'P.Match' on 1 position forward.
func (p *P[T]) Next(f Fn[T]) Fn[T] {
    return p.NextN(1, f)
}

// Fail() terminates parsing discarding current result.
func Fail[T any](p *P[T]) Fn[T] {
    fmt.Printf("parser.Fail(): Failed to parse with %#v\n", p.Data)
    return nil
}

// Done() terminates parsing sending current result.
func Done[T any](p *P[T]) Fn[T] {
    fmt.Printf("parser.Done(): Done with %#v\n", p.Data)
    p.Emit()
    return nil
}

// Run() starts async parsing by calling specified parsing function. Results
// (if any) will be delivered through returned channel.
func (p *P[T]) Run(s string) <-chan T {
    p.items = make(chan T)

    go func() {
        defer close(p.items)

        p.Match = p.rx.FindStringSubmatch(s)
        if len(p.Match) > 0 {
            fmt.Printf("parser.P.Run(): Initial submatches %#v\n", p.Match)
            for f := p.Next(p.fn); f != nil; {
                f = f(p)
            }

        } else {
            fmt.Printf("parser.P.Run(): Rx does not match\n")
        }
    }()

    return p.items
}

