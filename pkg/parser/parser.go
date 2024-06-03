
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

func (p *P) NextN(k int, f Fn) Fn {
    if k < 1 {
        panic(fmt.Sprintf("parser.NextN(): Error: Can't advance on less, than 1 submatch: %v", k))
    } else if len(p.Match) <= k {
        panic(fmt.Sprintf("parser.Next(): Too few submatches: advance on %v requested, but i have only %v (%#v)", k, len(p.Match), p.Match))
    }
    p.Match = p.Match[k:]
    return f
}

func (p *P) Next(f Fn) Fn {
    return p.NextN(1, f)
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
            fmt.Printf("parser.P.Run(): Submatches %#v\n", p.Match)
            for f := p.Next(p.Fn); f != nil; {
                f = f(p)
            }

        } else {
            fmt.Printf("Parser.Run(): Rx does not match\n")
        }
    }()

    return p.items
}

