
package store

import (
    "fmt"
    //"net/mail"
    "strings"

    . "bh/lastlog/pkg/common"
    "bh/lastlog/pkg/log"
    "bh/lastlog/pkg/intervals"
)

type IPData struct {
    IP IP   `json:"ip"`
    Method Method   `json:"method"`
    Time Time  `json:"time"`
    Count int   `json:"count"`
}

func (v *IPData) String() string {
    return fmt.Sprintf("IPData{%v, %v, %v, %v}", v.Time.Format("2006/01/02 15:04:05"), v.IP, v.Method, v.Count)
}

type UserData struct {
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

func (s *Store) Add(r *Result) {
    LogDfn("Adding %v", r)
    if _, ok := s.Data[r.User]; !ok {
        s.Data[r.User] = &UserData {
                            User: r.User,
                            Data: make(map[IP]map[Method]*IPData),
                        }
    }
    ud := s.Data[r.User]

    if _, ok := ud.Data[r.IP]; !ok {
        ud.Data[r.IP] = make(map[Method]*IPData)
    }

    if ipd, ok := ud.Data[r.IP][r.Method]; !ok {
        v := IPData {
                IP: r.IP,
                Method: r.Method,
                Time: Time{r.Time},
                Count: 1,
            }
        ud.Data[r.IP][r.Method] = &v
        if ud.Last == nil || ud.Last.Time.Before(r.Time) {
            ud.Last = &v
        }

    } else {
        ipd.Count += 1
        if ipd.Time.Before(r.Time) {
            ipd.Time = Time{r.Time}
        }
        if ud.Last.Time.Before(r.Time) {
            ud.Last = ipd
        }
    }
    LogDfn("Curr %v", s.Data)
}

func (s *Store) ReadLogs(l *log.L[Time, Result], files []string) {
    for _, f := range files {
        if err := log.OpenFile[Result](l, f); err != nil {
            LogErr("Can't open file '%v': %v", f, err)
            continue
        }

        ch := make(chan Result)
        go func() {
            defer close(ch)
            if err := l.Parse(ch); err != nil {
                LogErr("Can't read file '%v': %v", f, err)
                return
            }
        }()

        for r := range ch {
            s.Add(&r)
        }
    }
}

