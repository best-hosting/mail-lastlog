
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

type Log struct {
    file string
    mtime time.Time
    input io.Reader

    parser *parser.P
}

func NewLog(file string) (*Log, error) {
    l := Log{file: file}

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
    fn := func(p *parser.P) parser.Fn { return parseTime(&l, p) }
    l.parser = parser.New(`(\w+ +\d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, .*, rip=([0-9.]+),`, fn)

    return &l, nil
}

var _ parser.Parser = (*Log)(nil)
func (l *Log) Parse() <-chan parser.Result {
    items := make(chan parser.Result)

    go func() {
        defer close(items)

        scanner := bufio.NewScanner(l.input)
        for scanner.Scan() {
            fmt.Printf("Read '%s'\n", scanner.Text())
            for d := range l.parser.Run(scanner.Text()) {
                items <- *d
            }
        }
        if err := scanner.Err(); err != nil {
            fmt.Printf("dovecot.Parse(): Scanner error %v\n", err)
        }
    }()

    return items
}

// ParseTime() parses time in dovecot log, guessing correct year (because
// dovecot log files do not contain year).
func parseTime(l *Log, p *parser.P) parser.Fn {
    // Parse with current year and fix later, if that's wrong.
    t, err := time.Parse("2006 Jan _2 15:04:05", fmt.Sprintf("%v %s", l.mtime.Year(), p.Match[0]))
    if err != nil {
        fmt.Printf("dovecot.parseTime(): Error: %v, skipping\n", err)
        return parser.Fail
    }
    //fmt.Printf("Parsed time %v\n", t.Format("2006/01/02 15:04:06"))

    // File mtime should always be after any timestamp inside file. If it's
    // not the case, record's timestamp is from previous year (this is only
    // true, if file contains strictly less, than a year of data, though).
    if t.After(l.mtime) {
        t = t.AddDate(-1, 0, 0)
    }

    if t.Before(p.Last.Time) {
        fmt.Printf("dovecot.parseTime(): Warning: Skip record %v as already parsed\n", p.Data)
        return parser.Fail
    }

    p.Data.Time = t
    return p.Next(parseMethod)
}

func parseMethod(p *parser.P) parser.Fn {
    m, err := ToMethod(p.Match[0])
    if err != nil {
        fmt.Printf("dovecot.parseMethod(): Failed to parse method with '%v'\n", err)
        return parser.Fail
    }
    p.Data.Method = m

    return p.Next(parseUser)
}

func parseUser(p *parser.P) parser.Fn {
    p.Data.User = User(p.Match[0])

    return p.Next(parseIP)
}

func parseIP(p *parser.P) parser.Fn {
    v := net.ParseIP(p.Match[0])
    if v == nil {
        fmt.Printf("dovecot.parseIP(): Can't parse ip " + p.Match[0])
        return parser.Fail
    }

    p.Data.IP = IP(v.String())
    return parser.Done
}

