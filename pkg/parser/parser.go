
package parser

import (
    "fmt"
    "time"
    "regexp"

    . "bh/lastlog/pkg/common"
)

type Parser interface {
    Parse() <-chan Result
}

type Result struct {
    User User
    IP IP
    Method Method
    Time time.Time
}

type P struct {
    Rx *regexp.Regexp
    Fn Fn
    Match []string      // Regexp submatches.
    Last Result         // Last successfully parsed data.
    Data *Result        // Current incomplete result.

    items chan *Result
}

func New(rx string, fn Fn) *P {
    p := P {
        Rx: regexp.MustCompile(rx),
        Fn: fn,
    }
    return &p
}

type Fn func (*P) Fn

func Emit(p *P) {
    fmt.Printf("parser.Emit(): Emit %#v\n", p.Data)
    p.Last = *p.Data
    p.items<- p.Data
}

func (p *P) Next(f Fn) Fn {
    if len(p.Match) <= 1 {
        panic("parser.Next(): Too few submatches")
    }
    p.Match = p.Match[1:]
    return f
}

func Fail(p *P) Fn {
    fmt.Printf("parser.Fail(): Failed parse at %v\n", p.Data)
    return nil
}

func Done(p *P) Fn {
    fmt.Printf("parser.Done(): Done with %#v\n", p.Data)
    Emit(p)
    return nil
}

func (p *P) Run(s string) <-chan *Result {
    p.Data = &Result{}
    p.items = make(chan *Result)

    go func() {
        defer close(p.items)

        p.Match = p.Rx.FindStringSubmatch(s)
        if len(p.Match) > 0 {
            for f := p.Next(p.Fn); f != nil; {
                f = f(p)
            }

        } else {
            fmt.Printf("Parser.Run(): Rx does not match\n")
        }
    }()

    return p.items
}

