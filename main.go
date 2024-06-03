
package main

import (
    "fmt"
    "time"
    //"net/mail"
    "strings"
    "encoding/json"

    . "bh/lastlog/pkg/common"
    "bh/lastlog/pkg/parser"
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



/*
// ParseTime() parses time in dovecot log, guessing correct year (because
// dovecot log files do not contain year).
func parseTime(v *Log) Parser {
    // Parse with current year and fix later, if that's wrong.
    t, err := time.Parse("2006 Jan _2 15:04:05", fmt.Sprintf("%v %s", l.mtime.Year(), v.match[0]))
    if err != nil {
        fmt.Printf("Error: %v, skipping\n", err)
        return failParse
    }
    fmt.Printf("%v\n", t.Format("2006/01/02 15:04:06"))

    // File mtime should always be after any timestamp inside file. If it's
    // not the case, record's timestamp is from previous year (this is only
    // true, if file contains strictly less, than a year of data, though).
    if t.After(v.mtime) {
        v.data.Time = t.AddDate(-1, 0, 0)
    } else {
        v.data.Time = t
    }

    v.match = v.match[1:]
    return parseMethod
}
*/

func readLog(allData map[User]*UserData, dl parser.Parser) {
    for d := range dl.Parse() {
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
            v.Last = d.Time
            v.Count += 1
            ud.Last = v
        }
        fmt.Printf("Curr %v\n", allData)
    }
}

func main() {
    allData := make(map[User]*UserData)

    /*
    f, err := os.Open(file)
    if err != nil {
        panic(err)
    }
    fi, err := os.Stat(file)
    if err != nil {
        panic(err)
    }
    mtime := fi.ModTime()
    fmt.Printf("mtime = %v\n", mtime.Format("2006/01/02 15:04:06"))
    l := Log{mtime: mtime}
    t, _ := l.parseTime("Oct 21 10:27:23")
    fmt.Printf("t = %v\n", t.Format("2006/01/02 15:04:06"))
    */

    dl, err := dovecot.NewLog("1.log")
    if err != nil {
        panic(err)
    }
    readLog(allData, dl)

    el, err := exim4.NewLog("2.log")
    if err != nil {
        panic(err)
    }
    readLog(allData, el)

    //info := make(LoginData)

    /*
    //re := regexp.MustCompile(`(\w+ \d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, method=([^,]+), rip=([0-9.]+), .*?, ((TLS), )?session`)
    re := regexp.MustCompile(`(\w+ +\d+ \d+:\d+:\d+) .*(pop3|imap)-login: Login: user=<([^>]+)>, .*, rip=([0-9.]+),`)
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        //fmt.Printf("Got '%s'\n", scanner.Text())
        match := re.FindStringSubmatch(scanner.Text())
        if len(match) > 0 {
            //fmt.Printf("%q\n", match)

            last, err := time.Parse("Jan _2 15:04:05", match[1])
            if err != nil {
                panic(err)
            }

            fmt.Printf("%v - %v\n", last.Format("2006/01/02 15:04:06"), mtime.Sub(last))
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
    */
    //fmt.Printf("%v\n", allData)

    bs, err := json.MarshalIndent(allData, "", "\t")
    if err != nil {
        panic(err)
    }
    fmt.Printf("%s\n", string(bs))
}
