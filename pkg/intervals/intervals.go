
package intervals

import (
    "fmt"

    . "bh/lastlog/pkg/common"
)

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
    // TODO: Can i parse several logs, which use the same intervals,
    // simultaneously? Should i include mutex here then? Or use two channels
    // and separate go routine for managing locks?
}

func (iv *Intervals[T, K]) Filter(out chan<- K, in <-chan K, last T) {
    if iv.I == nil {
        iv.I = make([]I[T], 0)
    }
    ts := iv.I
    LogDfn("## Started with last %v, and intervals %v", last, ts)
    v, ok := <-in
    if !ok {
        LogDfn("End of stream")
        return
    }

    p := v.ToOrd()
    LogDfn("Got p = %v from %v", p, v)

    // j is index of next (strictly greater) interval. It's possible, that j == len(ts).
    var j int
    for ; j < len(ts) && ts[j].Start.Le(p); j++ {
    }
    LogDfn("Found j = %v", j)

    // i is index where to insert new interval t. Thus i >= 0 always.
    var i int
    var t I[T]
    if j > 0 && p.Le(ts[j-1].End) {
        i = j - 1
        t = ts[i]

        if last.Le(t.End) {
            LogDfn("Completely contained in (%v) %v, discard all", i, t)
            return
        }

        LogDfn("Discard %v", v)

    } else {
        i = j
        t = I[T]{p, p}
        LogDfn("Send %v", v)
        out<- v
    }
    LogDfn("Found i = %v, t = %v", i, t)

    for {
        v, ok := <-in
        if !ok {
            LogDfn("End of stream")
            break
        }

        p = v.ToOrd()
        LogDfn("Got p = %v from %v", p, v)

        if t.End.Lt(p) {
            for ; j < len(ts) && ts[j].End.Lt(p); j++ {
                LogDfn("Skip interval j = %v %v", j, ts[j])
            }

            if j < len(ts) && ts[j].Start.Le(p) {
                LogDfn("Merge with interval (%v) %v", j, ts[j])
                t.End = ts[j].End
                j += 1

                if last.Le(t.End) {
                    LogDfn("Completely contained in %v, discard the rest", t)
                    break
                }

                LogDfn("Discard %v", v)

            } else {
                LogDfn("Update End to %v", p)
                t.End = p
                LogDfn("Send %v", v)
                out<- v
            }

        } else {
            LogDfn("Discard %v", v)
        }
        LogDfn("Current interval %v", t)
    }

    LogDfn("Got indexes i = %v j = %v", i, j)
    // j == i - insert new element
    // j - i == 1 - do nothing
    // j - i >  1 - remove merged elements
    if j == i {
        LogDfn("Insert new interval")
        ts = append(ts[:i], append(make([]I[T], 1), ts[i:]...)...)
    } else if j - i > 1 {
        LogDfn("Delete extra intervals")
        ts = append(ts[:i+1], ts[j:]...)
    }
    ts[i] = t
    iv.I = ts

    LogDfn("Resulting intervals %v", ts)
    return
}

