
package parser

import (
    "fmt"
    "time"
    "regexp"

    . "bh/lastlog/pkg/common"
)

type Parser interface {
    Parse() <-chan RData
}

type RData struct {
    User User
    IP IP
    Method Method
    Time time.Time
}

type P struct {
    Rx *regexp.Regexp
    Fn Fn
    Match []string
    Data *RData
    Items chan *RData
}

func New(rx string, fn Fn) *P {
    p := P {
        Rx: regexp.MustCompile(rx),
        Fn: fn,
    }
    return &p
}

type Fn func (*P) Fn

func Fail(p *P) Fn {
    fmt.Printf("parser.Fail(): Failed parse at %v\n", p.Data)
    return nil
}

func Done(p *P) Fn {
    fmt.Printf("parser.Done(): Done with %#v\n", p.Data)
    p.Items<- p.Data
    return nil
}

func (p *P) Next(f Fn) Fn {
    if len(p.Match) <= 1 {
        panic("parser.Next(): Too few submatches")
    }
    p.Match = p.Match[1:]
    return f
}

func (p *P) Run(s string) <-chan *RData {
    p.Data = &RData{}
    p.Items = make(chan *RData)

    go func() {
        defer close(p.Items)

        p.Match = p.Rx.FindStringSubmatch(s)
        if len(p.Match) > 0 {
            for f := p.Next(p.Fn); f != nil; {
                f = f(p)
            }

        } else {
            fmt.Printf("Parser.Run(): Rx does not match\n")
        }
    }()

    return p.Items
}

