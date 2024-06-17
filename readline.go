
package main

import (
    "io"
    "os"
    "fmt"
)

func readLine(r io.ReadSeeker) ([]byte, error) {
    res := make([]byte, 0)
    buf := make([]byte, 13)

    for {
        n, err := r.Read(buf)
        fmt.Printf("Read '%s'\n", buf[:n])

        if n > 0 {
            for i, c := range buf[:n] {
                if c == '\n' {
                    fmt.Printf("eol found at pos %v (%v)\n", i, n)
                    res = append(res, buf[:i]...)
                    if _, err := r.Seek(int64(i + 1 - n), 1); err != nil {
                        panic(err)
                    }
                    return res, nil
                }
            }

            res = append(res, buf[:n]...)
        }

        if err != nil {
            if err == io.EOF {
                return res, io.EOF
            }
            panic(err)
        }
    }
}

func main() {
    f, err := os.Open("1.log")
    if err != nil {
        panic(err)
    }
    for {
        l, err := readLine(f)
        if len(l) > 0 {
            fmt.Printf("Got '%s'\n", l)
        }
        if err != nil {
            fmt.Printf("Got error %v\n", err)
            break
        }
    }
}
