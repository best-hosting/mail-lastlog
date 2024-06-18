
package store

import (
    "fmt"
    "time"
    //"net/mail"
    "strings"

    . "bh/lastlog/pkg/types"
    "bh/lastlog/pkg/log"
    "bh/lastlog/pkg/intervals"
)

type IPData struct {
    IP IP   `json:"ip"`
    Method Method   `json:"method"`
    Last time.Time  `json:"last"`
    Count int   `json:"count"`
}

func (v *IPData) String() string {
    return fmt.Sprintf("IPData{%v, %v, %v, %v}", v.Last.Format("02/01 15:04:05"), v.IP, v.Method, v.Count)
}

type UserData struct {
    // FIXME: Does not need this.
    User User   `json:"user"`
    Last *IPData    `json:"last"`
    Data map[IP]map[Method]*IPData  `json:"data"`
}

func (v *UserData) String() string {
    var b strings.Builder
    fmt.Fprintf(&b, "UserData{%v, %v,", v.User, v.Last)
    for _, e := range v.Data {
        fmt.Fprintf(&b, "\n\t\t%v", e)
    }
    fmt.Fprintf(&b, "\n\t},")
    return b.String()
}

type Store struct {
    Data map[User]*UserData
    Intervals map[string]*intervals.Intervals[Time, Result]
}

func New() *Store {
    s := Store {
        Data: make(map[User]*UserData),
        Intervals: make(map[string]*intervals.Intervals[Time, Result]),
    }
    return &s
}

func (s *Store) Add(d *Result) {
    if _, ok := s.Data[d.User]; !ok {
        s.Data[d.User] = &UserData {
                            User: d.User,
                            Data: make(map[IP]map[Method]*IPData),
                        }
    }
    ud := s.Data[d.User]

    if _, ok := ud.Data[d.IP]; !ok {
        ud.Data[d.IP] = make(map[Method]*IPData)
    }

    if v, ok := ud.Data[d.IP][d.Method]; !ok {
        v = &IPData {
                IP: d.IP,
                Method: d.Method,
                Last: d.Time,
                Count: 1,
            }
        ud.Data[d.IP][d.Method] = v
        ud.Last = v
    } else {
        v.Count += 1
        // FIXME: Should i check, that time does not go backward?
        if d.Time.After(v.Last) {
            v.Last = d.Time
        }
        if d.Time.After(ud.Last.Last) {
            ud.Last = v
        }
    }
    fmt.Printf("store.Add(): Curr %v\n", s.Data)
}

func (s *Store) ReadLogs(l *log.L[Time, Result], files []string) {
    for _, f := range files {
        if err := log.OpenFile[Result](l, f); err != nil {
            panic(err)
        }

        ch := make(chan Result)
        go func() {
            defer close(ch)
            if err := l.Parse(ch); err != nil {
                fmt.Printf("Error during reading '%v' file: '%v'", f, err)
                return
            }
        }()
        for d := range ch {
            s.Add(&d)
        }
    }
}

