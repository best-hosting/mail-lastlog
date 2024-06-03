
package exim4

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
    last time.Time
    items chan parser.RData
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
    fn := parseTime
    l.parser = parser.New(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) [^ ]+ <= [^ ]+ H=[^ ]+ \[([0-9.]+)\] .* A=[^:]+:([^ ]+)`, fn)

    return &l, nil
}

func (l *Log) Parse() <-chan parser.RData {
    items := make(chan parser.RData)

    go func() {
        defer close(items)

        scanner := bufio.NewScanner(l.input)
        for scanner.Scan() {
            fmt.Printf("Read '%s'\n", scanner.Text())
            ch := l.parser.Run(scanner.Text())
            d, ok := <-ch
            if ok {
                l.last = d.Time
                items <- *d
            }
        }
        if err := scanner.Err(); err != nil {
            fmt.Printf("huy.Parse(): Scanner error %v\n", err)
        }
    }()

    return items
}

func parseTime (p *parser.P) parser.Fn {
    t, err := time.Parse("2006-01-02 15:04:05", p.Match[0])
    if err != nil {
        fmt.Printf("exim4.parseTime(): Error: %v, skipping\n", err)
        return parser.Fail
    }
    p.Data.Time = t

    fmt.Printf("time %v\n", p.Data)
    return p.Next(parseIP)
}

func parseIP(p *parser.P) parser.Fn {
    if v := net.ParseIP(p.Match[0]); v == nil {
        fmt.Printf("exim4.parseIP(): Can't parse ip " + p.Match[0])
        return parser.Fail
    } else {
        p.Data.IP = IP(v.String())
    }

    fmt.Printf("ip %v\n", p.Data)
    return p.Next(parseUser)
}

func parseUser(p *parser.P) parser.Fn {
    p.Data.User = User(p.Match[0])

    fmt.Printf("user %v\n", p.Data)
    return parseMethod
}

func parseMethod(p *parser.P) parser.Fn {
    p.Data.Method = "smtp"
    fmt.Printf("method %v\n", p.Data)
    return parser.Done
}

