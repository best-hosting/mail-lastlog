
package main

import (
    "fmt"
    "time"
    //"net/mail"
    "strings"
    "encoding/json"

    . "bh/lastlog/pkg/types"
    //"bh/lastlog/pkg/parser"
    "bh/lastlog/pkg/log"
    "bh/lastlog/pkg/dovecot"
    "bh/lastlog/pkg/exim4"
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
    fmt.Fprintf(&b, "UserData{%v, %v, ", v.User, v.Last)
    for _, e := range v.Data {
        fmt.Fprintf(&b, "\n\t%v", e)
    }
    fmt.Fprintf(&b, "\n")
    return b.String()
}

// FIXME: Different per-user views: group all entries by IP (i.e. several IP
// may have several methods used) or split by method.


type lastlogData map[User]*UserData

func (allData lastlogData) Add(d *Result) {
    if _, ok := allData[d.User]; !ok {
        allData[d.User] = &UserData {
                            User: d.User,
                            Data: make(map[IP]map[Method]*IPData),
                        }
    }
    ud := allData[d.User]

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
    fmt.Printf("Curr %v\n", allData)
}

func (allData lastlogData) readLog(l *log.L[Result]) {
    for d := range l.Parse() {
        allData.Add(&d)
    }
}

func main() {
    allData := lastlogData(make(map[User]*UserData))

    dl, err := dovecot.NewLog("1.log", time.Time{})
    if err != nil {
        panic(err)
    }
    allData.readLog(dl)

    el, err := exim4.NewLog("2.log", time.Time{})
    if err != nil {
        panic(err)
    }
    allData.readLog(el)

    bs, err := json.MarshalIndent(allData, "", "\t")
    if err != nil {
        panic(err)
    }
    fmt.Printf("%s\n", string(bs))
}
