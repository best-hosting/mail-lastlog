
package types

import (
    "fmt"
    "time"
    "encoding/json"
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
    return t.Format("2006/01/02 15:04:06")
}

var _ Ord[Time] = Time{}
func (t Time) Lt(t1 Time) bool { return t.Before(t1.Time) }
func (t Time) Le(t1 Time) bool { return t.Lt(t1) || t.Equal(t1.Time) }

var _ json.Marshaler = Time{}
func (t Time) MarshalJSON() ([]byte, error) {
    ft := time.DateTime + " -07:00"
    b := make([]byte, 0, len(ft) + len(`""`))
    b = append(b, '"')
    b = t.AppendFormat(b, ft)
    b = append(b, '"')
    return b, nil
}

// I can't overwrite time.Time's json encoding by implementing TextMarshaler
// interface for Time{}, because it embeds time.Time and therefore inherits
// its method set. And because time.Time already have implemented
// json.Marshaler interface, it'll be used instead of my TextMarshaler here.
//
// But i really should implement TextMarshaler/Unmarshaler to avoid unescaping
// json encoded string, which is not trivial, see time.UnmarshalJSON() code
// comments.
//
// On the other hand, i still want to embed time.Time to preserve all its
// methods and making Time{} wrapper almost invisible.
//
// Thus, as a workaround UnmarshalJSON() first unmarshals to string and then
// parses it.
var _ json.Unmarshaler = (*Time)(nil)
func (t *Time) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil {
        return err
    }
    v, err := time.Parse(time.DateTime + " -07:00", string(b))
    if err != nil {
        return err
    }
    t.Time = v
    return nil
}

// FIXME: Use net/mail.Address.
type User string

type IP string

type Method string
var UnknownMethod Method = ""
var Imap Method = "imap"
var Pop3 Method = "pop3"
var Smtp Method = "smtp"
var Web  Method = "web"

func ToMethod(v string) (Method, error) {
    switch v {
        case "imap": return Imap, nil
        case "pop3": return Pop3, nil
        case "smtp": return Smtp, nil
        case "web": return Web, nil
    }
    return UnknownMethod, fmt.Errorf("Unknown method %v", v)
}

type Result struct {
    User User
    IP IP
    Method Method
    Time time.Time
}

var _ ToOrd[Time] = Result{}
func (r Result) ToOrd() Time { return Time{r.Time} }
