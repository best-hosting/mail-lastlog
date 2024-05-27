
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
type IPInfo struct {
    IP IP
    Method Method
    Last time.Time
    Count int
}

func (v *IPInfo) String() string {
    return fmt.Sprintf("IPInfo{%v, %v, %v, %v}", v.Last.Format("02/01 15:04:05"), v.IP, v.Method, v.Count)
}

type User string
type UserInfo struct {
    User User
    Last *IPInfo
    Info map[IP]map[Method]*IPInfo
}

func (v *UserInfo) String() string {
    var b strings.Builder
    fmt.Fprintf(&b, "UserInfo{%v, %v, ", v.User, v.Last)
    for _, e := range v.Info {
        fmt.Fprintf(&b, "\n\t%v", e)
    }
    fmt.Fprintf(&b, "\n")
    return b.String()
}

// FIXME: Different per-user views: group all entries by IP (i.e. several IP
// may have several methods used) or split by method.

func main() {
    allInfo := make(map[User]*UserInfo)

    f, err := os.Open("mail.log")
    if err != nil {
        panic(err)
    }
    //info := make(LoginInfo)

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
            if _, ok := allInfo[user]; !ok {
                allInfo[user] = &UserInfo {
                                    User: user,
                                    Info: make(map[IP]map[Method]*IPInfo),
                                }
            }
            ui := allInfo[user]

            if _, ok := ui.Info[ip]; !ok {
                ui.Info[ip] = make(map[Method]*IPInfo)
            }

            if v, ok := ui.Info[ip][method]; !ok {
                v = &IPInfo {
                        IP: ip,
                        Method: method,
                        Last: last,
                        Count: 1,
                    }
                ui.Info[ip][method] = v
                ui.Last = v
            } else {
                v.Last = last
                v.Count += 1
                ui.Last = v
            }
        }
        //fmt.Printf("%v\n", allInfo)
    }
    if err := scanner.Err(); err != nil {
        panic(err)
    }
    fmt.Printf("%v\n", allInfo)
}
