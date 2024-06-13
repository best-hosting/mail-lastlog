
package types

import (
    "fmt"
    "time"
)

// By using generics,
// - check at compile time, that elements type T has Ord instantiation with
// itself.
type ToOrd[T Ord[T]] interface {
    ToOrd() T
}

// By using generics,
// - i may constrain later, that certain type impelements Ord on itself with
// '[T Ord[T]]' constraint.  And this will be checked at compile time (which
// wouldn't be possible, if Ord is concrete (non-generic) interface and Le has
// type Le(Ord)).
// - types implementing this interface will implement methods with arguments
// of concrete type (instead of arguments of Ord interface type, which require
// type assertions to check what's inside Ord interface and panic-ing if
// there's something unexpected).
type Ord[T any] interface {
    Lt(T) bool
    Le(T) bool
}

type Time struct {
    time.Time
}

func (t Time) String() string {
    return t.Format("2006/02/03 15:04:06")
}

var _ Ord[Time] = Time{}
func (t Time) Lt(t1 Time) bool { return t.Before(t1.Time) }
func (t Time) Le(t1 Time) bool { return t.Lt(t1) || t.Equal(t1.Time) }


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

type Result struct {
    User User
    IP IP
    Method Method
    Time time.Time
}

var _ ToOrd[Time] = Result{}
func (r Result) ToOrd() Time { return Time{r.Time} }
