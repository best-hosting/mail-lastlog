
package intervals

import (
    "fmt"

    . "bh/lastlog/pkg/types"
)

// TODO: Use debug print instead of just print.

// By using generics,
// - check at compile time, that elements type T has Ord instantiation with
// itself.
// - check at compile time, that both 'Start' and 'stop' types are the same.
type I[T Ord[T]] struct {
    Start T
    End T
}

func (i I[T]) String() string {
    return fmt.Sprintf("{%v - %v}", i.Start, i.End)
}

type Intervals[T Ord[T], K ToOrd[T]] struct {
    I []I[T]
    // FIXME: Can i parse several logs, which use the same intervals,
    // simultaneously? Should i include mutex here then? Or use two channels
    // and separate go routine for managing locks?
}

func (iv *Intervals[T, K]) Filter(out chan<- K, in <-chan K, last T) {
    if iv.I == nil {
        iv.I = make([]I[T], 0)
    }
    ts := iv.I
    fmt.Printf("intervals.FilterBy(): ## Started with last %v, and intervals %v\n", last, ts)
    v, ok := <-in
    if !ok {
        fmt.Printf("intervals.FilterBy(): End of stream\n")
        return
    }

    p := v.ToOrd()
    fmt.Printf("intervals.FilterBy(): Got p = %v from %v\n", p, v)

    // j is index of next (strictly greater) interval. It's possible, that j == len(ts).
    var j int
    for ; j < len(ts) && ts[j].Start.Le(p); j++ {
    }
    fmt.Printf("intervals.FilterBy(): Found j = %v\n", j)

    // i is index where to insert new interval t. Thus i >= 0 always.
    var i int
    var t I[T]
    if j > 0 && p.Le(ts[j-1].End) {
        i = j - 1
        t = ts[i]

        if last.Le(t.End) {
            fmt.Printf("intervals.FilterBy(): Completely contained in (%v) %v, discard all\n", i, t)
            return
        }

        fmt.Printf("intervals.FilterBy(): Discard %v\n", v)

    } else {
        i = j
        t = I[T]{p, p}
        fmt.Printf("intervals.FilterBy(): Send %v\n", v)
        out<- v
    }
    fmt.Printf("intervals.FilterBy(): Found i = %v, t = %v\n", i, t)

    for {
        v, ok := <-in
        if !ok {
            fmt.Printf("intervals.FilterBy(): End of stream\n")
            break
        }

        p = v.ToOrd()
        fmt.Printf("intervals.FilterBy(): Got p = %v from %v\n", p, v)

        if t.End.Lt(p) {
            for ; j < len(ts) && ts[j].End.Lt(p); j++ {
                fmt.Printf("intervals.FilterBy(): Skip interval j = %v %v\n", j, ts[j])
            }

            if j < len(ts) && ts[j].Start.Le(p) {
                fmt.Printf("intervals.FilterBy(): Merge with interval (%v) %v\n", j, ts[j])
                t.End = ts[j].End
                j += 1

                if last.Le(t.End) {
                    fmt.Printf("intervals.FilterBy(): Completely contained in %v, discard the rest\n", t)
                    break
                }

                fmt.Printf("intervals.FilterBy(): Discard %v\n", v)

            } else {
                fmt.Printf("intervals.FilterBy(): Update End to %v\n", p)
                t.End = p
                fmt.Printf("intervals.FilterBy(): Send %v\n", v)
                out<- v
            }

        } else {
            fmt.Printf("intervals.FilterBy(): Discard %v\n", v)
        }
        fmt.Printf("intervals.FilterBy(): Current interval %v\n", t)
    }

    fmt.Printf("intervals.FilterBy(): Got indexes i = %v j = %v\n", i, j)
    // j == i - insert new element
    // j - i == 1 - do nothing
    // j - i >  1 - remove merged elements
    if j == i {
        fmt.Printf("intervals.FilterBy(): Insert new interval\n")
        ts = append(ts[:i], append(make([]I[T], 1), ts[i:]...)...)
    } else if j - i > 1 {
        fmt.Printf("intervals.FilterBy(): Delete extra intervals\n")
        ts = append(ts[:i+1], ts[j:]...)
    }
    ts[i] = t
    iv.I = ts

    fmt.Printf("intervals.FilterBy(): Resulting intervals %v\n", ts)
    return
}

