
package intervals

import (
    "fmt"
    "sync"
    "testing"

    . "bh/lastlog/pkg/common"
)

type myData struct {
    v myInt
}

var _ ToOrd[myInt] = myData{}
func (s myData) ToOrd() myInt { return s.v }

type myInt int

var _ Ord[myInt] = (myInt)(0)

// Methods are implemented on concrete types, thus no type assertions and
// runtime panic needed.
func (x myInt) Lt(y myInt) bool {
    return int(x) < int(y)
}
func (x myInt) Le(y myInt) bool {
    return int(x) <= int(y)
}

type Data[T Ord[T], K ToOrd[T]] struct {
    in  []I[T]
    out []I[T]
    gen func() []K
    res []K
    last T
}

var data = []Data[myInt, myData]{
    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {0, 11} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 0; i < 12; i++ {
                xs = append(xs, myData{myInt(i)})
            }
            return xs
        },
        res: []myData{ {0}, {1}, {4}, {9}, {10}, {11}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {0, 0}, {2, 3}, {5, 7}, {8, 8} },
        gen: func() (xs []myData) {
            xs = []myData{ {0} }
            return
        },
        res: []myData{ {0} },
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        gen: func() (xs []myData) {
            xs = []myData{ {2} }
            return
        },
        res: []myData{},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        gen: func() (xs []myData) {
            xs = []myData{ myData{myInt(3)} }
            return
        },
        res: []myData{},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        gen: func() (xs []myData) {
            xs = []myData{ myData{myInt(6)} }
            return
        },
        res: []myData{},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        gen: func() (xs []myData) {
            xs = []myData{ myData{myInt(8)} }
            return
        },
        res: []myData{},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {0, 1}, {2, 3}, {5, 7}, {8, 8} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 0; i < 2; i++ {
                xs = append(xs, myData{myInt(i)})
            }
            return xs
        },
        res: []myData{{0}, {1}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {0, 7}, {8, 8} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 0; i < 2; i++ {
                xs = append(xs, myData{myInt(i)})
            }
            xs = append(xs, myData{myInt(5)})
            return xs
        },
        res: []myData{myData{0}, myData{1}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {0, 7}, {8, 8} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 0; i < 2; i++ {
                xs = append(xs, myData{myInt(i)})
            }
            xs = append(xs, myData{myInt(6)})
            return xs
        },
        res: []myData{myData{0}, myData{1}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {0, 7}, {8, 8} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 0; i < 2; i++ {
                xs = append(xs, myData{myInt(i)})
            }
            xs = append(xs, myData{myInt(7)})
            return xs
        },
        res: []myData{myData{0}, myData{1}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {1, 8} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 0; i < 4; i++ {
                xs = append(xs, myData{myInt(pow(2, i))})
            }
            return xs
        },
        res: []myData{myData{1}, myData{4}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {4, 16} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 2; i < 5; i++ {
                xs = append(xs, myData{myInt(pow(2, i))})
            }
            return xs
        },
        res: []myData{{4}, {16}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {5, 7}, {8, 16} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 3; i < 5; i++ {
                xs = append(xs, myData{myInt(pow(2, i))})
            }
            return xs
        },
        res: []myData{{16}},
        last: 10,
    },

    {
        in:  []I[myInt]{ {2, 3}, {5, 7}, {8, 8} },
        out: []I[myInt]{ {2, 3}, {5, 7}, {8, 8}, {16, 16} },
        gen: func() []myData {
            xs := make([]myData, 0)
            for i := 4; i < 5; i++ {
                xs = append(xs, myData{myInt(pow(2, i))})
            }
            return xs
        },
        res: []myData{{16}},
        last: 10,
    },
}

func pow(n, p int) int {
    if p == 0 {
        return 1
    }

    x := n
    for i := 1; i < p; i++ {
        x *= n
    }
    return x
}

func runFtest(td Data[myInt, myData], t *testing.T) {
    s := Intervals[myInt, myData]{ make([]I[myInt], len(td.in)) }
    copy(s.I, td.in)

    done := make(chan any)
    in := make(chan myData)
    out := make(chan myData)

    go func() {
        defer close(in)
        for _, v := range td.gen() {
            select {
            case <-done: return
            case in <- v:
            }
        }
    }()

    rs := make([]myData, 0)
    var wg sync.WaitGroup
    wg.Add(1)
    go func() {
        defer wg.Done()
        for d := range out {
            rs = append(rs, d)
        }
    }()

    s.Filter(out, in, td.last)
    close(done)
    close(out)
    wg.Wait()

    if len(rs) != len (td.res) {
        t.Fatalf("%v: Results length differ: want %v, got %v", t.Name(), td.res, rs)
    }
    for i := 0; i < len(rs); i++ {
        if rs[i] != td.res[i] {
            t.Fatalf("%v: Results differ at position %v: want %v, got %v", t.Name(), i, td.res, rs)
        }
    }

    if len(s.I) != len(td.out) {
        t.Fatalf("%v: Result intervals length differ: want %v, got %v", t.Name(), td.out, s.I)
    }
    for i := 0; i < len(s.I); i++ {
        if s.I[i] != td.out[i] {
            t.Fatalf("%v: Result intervals differ at position %v: want %v, got %v", t.Name(), i, td.out, s.I)
        }
    }
}

func TestFilterBy(t *testing.T) {
    for i, td := range data {
        ok := t.Run(fmt.Sprintf("%v", i), func(t *testing.T) { runFtest(td, t) })
        if !ok {
            t.FailNow()
        }
    }
    return
}

