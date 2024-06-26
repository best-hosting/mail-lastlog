
package dovecot

import (
    "fmt"
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
    t, err := time.ParseInLocation("2006 Jan _2 15:04:05", fmt.Sprintf("%v %s", mtime.Year(), p.Match[0]), time.Local)
    if err != nil {
        LogErr("Failed to parse time: %v", err)
        return parser.Fail
    }

    Logfn("Got mtime %v", mtime)
    // File mtime should always be after any timestamp inside file. If it's
    // not the case, record's timestamp is from previous year (this is only
    // true, if file contains strictly less, than a year of data, though).
    if mtime.Before(t) {
        t = t.AddDate(-1, 0, 0)
    }

    p.Data.Time = t
    return p.Next(parseMethod)
}

func parseMethod(p *parser.P[Result]) parser.Fn[Result] {
    m, err := ToMethod(p.Match[0])
    if err != nil {
        LogErr("Failed to parse method with '%v'", err)
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
        LogErr("Can't parse ip '%v'", p.Match[0])
        return parser.Fail
    }

    p.Data.IP = IP(v.String())
    return parser.Done
}

