
package dovecot

import (
    "fmt"
    "time"
    "net"

    . "bh/lastlog/pkg/types"
    "bh/lastlog/pkg/parser"
    "bh/lastlog/pkg/intervals"
    "bh/lastlog/pkg/log"
)

func NewLog(i *intervals.Intervals[Time, Result]) *log.L[Time, Result] {
    l := log.L[Time, Result]{}
    l.Intervals = i

    // Parsers are run in the order of matches. And this order is hardcoded in
    // their return values. See below.
    fn := func(p *parser.P[Result]) parser.Fn[Result] {
        p.Data = Result{}
        return parseTime(&l.ModTime, p)
    }
    l.Parser = parser.NewP(`(\w+ +\d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, .*, rip=([0-9.]+),`, fn)

    return &l
}

// ParseTime() parses time in dovecot log, guessing correct year (because
// dovecot log files do not contain year).
func parseTime(mtime *Time, p *parser.P[Result]) parser.Fn[Result] {
    // Parse with current year and fix later, if that's wrong.
    t, err := time.Parse("2006 Jan _2 15:04:05", fmt.Sprintf("%v %s", mtime.Year(), p.Match[0]))
    if err != nil {
        fmt.Printf("dovecot.parseTime(): Error: Failed to parse time with '%v'\n", err)
        return parser.Fail
    }
    //fmt.Printf("Parsed time %v\n", t.Format("2006/01/02 15:04:06"))

    // File mtime should always be after any timestamp inside file. If it's
    // not the case, record's timestamp is from previous year (this is only
    // true, if file contains strictly less, than a year of data, though).
    if t.After(mtime.Time) {
        t = t.AddDate(-1, 0, 0)
    }

    p.Data.Time = t
    return p.Next(parseMethod)
}

func parseMethod(p *parser.P[Result]) parser.Fn[Result] {
    m, err := ToMethod(p.Match[0])
    if err != nil {
        fmt.Printf("dovecot.parseMethod(): Error: Failed to parse method with '%v'\n", err)
        return parser.Fail
    }
    p.Data.Method = m

    return p.Next(parseUser)
}

func parseUser(p *parser.P[Result]) parser.Fn[Result] {
    p.Data.User = User(p.Match[0])

    return p.Next(parseIP)
}

func parseIP(p *parser.P[Result]) parser.Fn[Result] {
    v := net.ParseIP(p.Match[0])
    if v == nil {
        fmt.Printf("dovecot.parseIP(): Error: Can't parse ip " + p.Match[0])
        return parser.Fail
    }

    p.Data.IP = IP(v.String())
    return parser.Done
}

