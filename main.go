
package main

import (
    "os"
    "fmt"
    "bufio"
    "regexp"
    "time"
    //"net/mail"
    "net"
    "strings"
    "encoding/json"
)

type Method string
var Imap Method = "imap"
var Pop3 Method = "pop3"
var Smtp Method = "smtp"
var Web  Method = "web"
var Unknown Method = ""
func ToMethod(v string) (Method, error) {
    switch v {
        case "imap": return Imap, nil
        case "pop3": return Pop3, nil
        case "smtp": return Smtp, nil
        case "web": return Web, nil
    }
    return Unknown, fmt.Errorf("Unknown method %v", v)
}

type IP string
type IPData struct {
    IP IP   `json:"ip"`
    Method Method   `json:"method"`
    Last time.Time  `json:"last"`
    Count int   `json:"count"`
}

func (v *IPData) String() string {
    return fmt.Sprintf("IPData{%v, %v, %v, %v}", v.Last.Format("02/01 15:04:05"), v.IP, v.Method, v.Count)
}

type User string
type UserData struct {
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

type DovecotLog struct {
    mtime time.Time
}

// parseTime() parses time in dovecot log, guessing correct year (because
// dovecot log files do not contain year).
func (l *DovecotLog) ParseTime(ts string) (time.Time, error) {
    // Parse with current year and fix later, if that's wrong.
    t, err := time.Parse("2006 Jan _2 15:04:05", fmt.Sprintf("%v %s", l.mtime.Year(), ts))
    if err != nil {
        return time.Time{}, err
    }
    //fmt.Printf("%v\n", t.Format("2006/01/02 15:04:06"))

    // File mtime should always be after any timestamp inside file. If it's
    // not the case, record's timestamp is from previous year (this is only
    // true, if file contains strictly less, than a year of data, though).
    if t.After(l.mtime) {
        return t.AddDate(-1, 0, 0), nil
    }
    return t, nil
}

func main() {
    allData := make(map[User]*UserData)

    file := "1.log"
    f, err := os.Open(file)
    if err != nil {
        panic(err)
    }
    fi, err := os.Stat(file)
    if err != nil {
        panic(err)
    }
    mtime := fi.ModTime()
    //fmt.Printf("mtime = %v\n", mtime.Format("2006/01/02 15:04:06"))
    dovecot := DovecotLog{mtime: mtime}

    //info := make(LoginData)

    //re := regexp.MustCompile(`(\w+ \d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, method=([^,]+), rip=([0-9.]+), .*?, ((TLS), )?session`)
    re := regexp.MustCompile(`(\w+ +\d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, .*, rip=([0-9.]+),`)
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        //fmt.Printf("Got '%s'\n", scanner.Text())
        match := re.FindStringSubmatch(scanner.Text())
        if len(match) > 0 {
            //fmt.Printf("%q\n", match)

            last, err := dovecot.ParseTime(match[1])
            if err != nil {
                panic(err)
            }

            method, err := ToMethod(match[2])
            if err != nil {
                panic(err)
            }

            user := User(match[3])
            var ip IP
            if v := net.ParseIP(match[4]); v == nil {
                panic("Can't parse ip " + match[4])
            } else {
                ip = IP(v.String())
            }

            //fmt.Printf("%v %v %v %v\n", last.Format("02/01 15:04:05"), method, user, ip)
            if _, ok := allData[user]; !ok {
                allData[user] = &UserData {
                                    User: user,
                                    Data: make(map[IP]map[Method]*IPData),
                                }
            }
            ud := allData[user]

            if _, ok := ud.Data[ip]; !ok {
                ud.Data[ip] = make(map[Method]*IPData)
            }

            if v, ok := ud.Data[ip][method]; !ok {
                v = &IPData {
                        IP: ip,
                        Method: method,
                        Last: last,
                        Count: 1,
                    }
                ud.Data[ip][method] = v
                ud.Last = v
            } else {
                v.Last = last
                v.Count += 1
                ud.Last = v
            }
        }
        //fmt.Printf("%v\n", allData)
    }
    if err := scanner.Err(); err != nil {
        panic(err)
    }
    //fmt.Printf("%v\n", allData)
    bs, err := json.Marshal(allData)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%s\n", string(bs))
}
