
package common

import (
    "fmt"
)

type User string

type IP string

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

