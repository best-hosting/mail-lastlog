
package exim4

import (
    "os"
    "fmt"
    "time"
    "net"

    . "bh/lastlog/pkg/common"
    "bh/lastlog/pkg/parser"
    "bh/lastlog/pkg/log"
)

func NewLog(file string, last time.Time) (*log.L[Result], error) {
    l := log.L[Result]{}

    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    l.Input = f

    // Parsers are run in the order of matches. And this order is hardcoded in
    // their return values. See below.
    fn := func (p *parser.P[Result]) parser.Fn[Result] { return parseTime(last, p) }
    // Note, that such usage of submatches as here in IP address regexp is
    // wrong: i verify IP address in parseIP() anyway, so i need the simplest
    // regexp, which will capture the longest possible IP (i.e. i don't need
    // even try to filter out wrong IP-s at regexp level). It's written here
    // only as example of p.NextN().
    l.Parser = parser.NewP(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) [^ ]+ <= [^ ]+ H=[^ ]+ \[(([0-9]+.){3}[0-9]+)\] .* A=[^:]+:([^ ]+)`, fn)

    return &l, nil
}

func parseTime (last time.Time, p *parser.P[Result]) parser.Fn[Result] {
    t, err := time.Parse("2006-01-02 15:04:05", p.Match[0])
    if err != nil {
        fmt.Printf("exim4.parseTime(): Error: Failed to parse time with '%v'\n", err)
        return parser.Fail
    }

    if t.Before(last) {
        fmt.Printf("exim4.parseTime(): Warning: Skip record %v as already parsed\n", p.Data)
        return parser.Fail
    }

    p.Data.Time = t
    return p.Next(parseIP)
}

func parseIP(p *parser.P[Result]) parser.Fn[Result] {
    v := net.ParseIP(p.Match[0])
    if v == nil {
        fmt.Printf("exim4.parseIP(): Can't parse ip " + p.Match[0])
        return parser.Fail
    }

    p.Data.IP = IP(v.String())
    return p.NextN(2, parseUser)
}

func parseUser(p *parser.P[Result]) parser.Fn[Result] {
    // FIXME: Parse mail address.
    p.Data.User = User(p.Match[0])

    return parseMethod
}

func parseMethod(p *parser.P[Result]) parser.Fn[Result] {
    p.Data.Method = "smtp"
    return parser.Done
}

