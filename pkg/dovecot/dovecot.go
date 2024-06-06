
package dovecot

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
    l := log.L[Result]{Last: last}

    fi, err := os.Stat(file)
    if err != nil {
        return nil, err
    }

    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    l.Input = f

    // Parsers are run in the order of matches. And this order is hardcoded in
    // their return values. See below.
    fn := func(p *parser.P[Result]) parser.Fn[Result] {
        p.Data = Result{}
        // FIXME: Should it be pointer? Or otherwise how new last time will be
        // seen py parseTime() ?
        return parseTime(fi.ModTime(), l.Last, p)
    }
    l.Parser = parser.NewP(`(\w+ +\d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, .*, rip=([0-9.]+),`, fn)

    return &l, nil
}

// FIXME:
// func NewTailLog() ....

// ParseTime() parses time in dovecot log, guessing correct year (because
// dovecot log files do not contain year).
func parseTime(mtime time.Time, last time.Time, p *parser.P[Result]) parser.Fn[Result] {
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
    if t.After(mtime) {
        t = t.AddDate(-1, 0, 0)
    }

    if t.Before(last) {
        fmt.Printf("dovecot.parseTime(): Warning: Skip record %v as already parsed\n", p.Data)
        return parser.Fail
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

