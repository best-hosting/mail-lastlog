
package exim4

import (
    "time"
    "net"

    . "bh/lastlog/pkg/common"
    "bh/lastlog/pkg/parser"
    "bh/lastlog/pkg/intervals"
    "bh/lastlog/pkg/log"
)

func NewLog(i *intervals.Intervals[Time, Result]) *log.L[Time, Result] {
    l := log.L[Time, Result]{}
    if i == nil {
        l.Intervals = &intervals.Intervals[Time, Result]{}
    } else {
        l.Intervals = i
    }

    // Note, that such usage of submatches as here in IP address regexp is
    // wrong: i verify IP address in parseIP() anyway, so i need the simplest
    // regexp, which will capture the longest possible IP (i.e. i don't need
    // even try to filter out wrong IP-s at regexp level). It's written here
    // only to show an example of p.NextN() usage.
    l.Parser = parser.NewP(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) [^ ]+ <= [^ ]+ H=[^ ]+ \[(([0-9]+.){3}[0-9]+)\] .* A=[^:]+:([^ ]+)`, parseTime)

    return &l
}

func parseTime (p *parser.P[Result]) parser.Fn[Result] {
    t, err := time.Parse("2006-01-02 15:04:05", p.Match[0])
    if err != nil {
        LogErr("Failed to parse time: %v", err)
        return parser.Fail
    }

    p.Data.Time = t
    return p.Next(parseIP)
}

func parseIP(p *parser.P[Result]) parser.Fn[Result] {
    v := net.ParseIP(p.Match[0])
    if v == nil {
        LogErr("Can't parse ip '%v'", p.Match[0])
        return parser.Fail
    }

    p.Data.IP = IP(v.String())
    return p.NextN(2, parseUser)
}

func parseUser(p *parser.P[Result]) parser.Fn[Result] {
    p.Data.User = User(p.Match[0])

    return parseMethod
}

func parseMethod(p *parser.P[Result]) parser.Fn[Result] {
    p.Data.Method = "smtp"
    return parser.Done
}

