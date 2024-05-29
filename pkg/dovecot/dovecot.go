
package dovecot

import (
    "os"
    "fmt"
    "bufio"
    "time"
    "io"
    "net"

    . "bh/lastlog/pkg/common"
    "bh/lastlog/pkg/parser"
)

type DovecotLog struct {
    file string
    mtime time.Time
    input io.Reader

    parser *parser.P
    last time.Time
    items chan parser.RData
}

func NewDovecotLog(file string) (*DovecotLog, error) {
    l := DovecotLog{file: file}

    fi, err := os.Stat(file)
    if err != nil {
        return nil, err
    }
    l.mtime = fi.ModTime()

    f, err := os.Open(l.file)
    if err != nil {
        return nil, err
    }
    l.input = f

    // Parsers are run in the order of matches. And this order is hardcoded in
    // their return values. See below.
    fn := func(p *parser.P) parser.ParseFn { return parseTime(&l, p) }
    l.parser = parser.New(`(\w+ +\d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, .*, rip=([0-9.]+),`, fn)

    return &l, nil
}

func (l *DovecotLog) Parse() <-chan parser.RData {
    l.items = make(chan parser.RData)

    go func() {
        defer close(l.items)

        scanner := bufio.NewScanner(l.input)
        for scanner.Scan() {
            fmt.Printf("Read '%s'\n", scanner.Text())
            ch := l.parser.Run(scanner.Text())
            d, ok := <-ch
            if ok {
                l.last = d.Time
                l.items <- *d
            }
        }
        if err := scanner.Err(); err != nil {
            fmt.Printf("dovecot.Parse(): Scanner error %v\n", err)
        }
    }()

    return l.items
}

// ParseTime() parses time in dovecot log, guessing correct year (because
// dovecot log files do not contain year).
func parseTime(l *DovecotLog, p *parser.P) parser.ParseFn {
    // Parse with current year and fix later, if that's wrong.
    t, err := time.Parse("2006 Jan _2 15:04:05", fmt.Sprintf("%v %s", l.mtime.Year(), p.Match[0]))
    if err != nil {
        fmt.Printf("Error: %v, skipping\n", err)
        return parser.Fail
    }
    //fmt.Printf("Parsed time %v\n", t.Format("2006/01/02 15:04:06"))

    // File mtime should always be after any timestamp inside file. If it's
    // not the case, record's timestamp is from previous year (this is only
    // true, if file contains strictly less, than a year of data, though).
    if t.After(l.mtime) {
        p.Data.Time = t.AddDate(-1, 0, 0)
    } else {
        p.Data.Time = t
    }

    if p.Data.Time.Before(l.last) {
        fmt.Printf("dovecot.ParseTime(): Skip record %v as already parsed\n", p.Data)
        return parser.Fail
    }

    return p.Next(parseMethod)
}

func parseMethod(p *parser.P) parser.ParseFn {
    m, err := ToMethod(p.Match[0])
    if err != nil {
        fmt.Printf("dovecot.ParseMethod(): Failed to parse method with '%v'\n", err)
        return parser.Fail
    }
    p.Data.Method = m

    return p.Next(parseUser)
}

func parseUser(p *parser.P) parser.ParseFn {
    p.Data.User = User(p.Match[0])

    return p.Next(parseIP)
}

func parseIP(p *parser.P) parser.ParseFn {
    if v := net.ParseIP(p.Match[0]); v == nil {
        fmt.Printf("dovecot.ParseIP(): Can't parse ip " + p.Match[0])
        return parser.Fail
    } else {
        p.Data.IP = IP(v.String())
    }

    return parser.Done
}

