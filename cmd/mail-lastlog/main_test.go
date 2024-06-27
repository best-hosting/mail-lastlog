
package main

import (
    "testing"
    "os"
    "bufio"
    "fmt"
    "path/filepath"
)

const testDir = "testdata"
var conf Config
var resStore string

func TestRun(t *testing.T) {
    if err := run(conf); err != nil {
        t.Fatalf("%v: main.run() exited with error %v", t.Name(), err)
    }

    fmt.Printf("%#v\n", conf)
    f1, err := os.Open(resStore)
    if err != nil {
        t.Fatalf("%v: Can't open expected results file '%v': %v", t.Name(), resStore, err)
    }
    defer f1.Close()
    want := bufio.NewScanner(f1)

    f2, err := os.Open(conf.Store)
    if err != nil {
        t.Fatalf("%v: Can't open generated results file '%v': %v", t.Name(), conf.Store, err)
    }
    defer f2.Close()
    got := bufio.NewScanner(f2)

    for want.Scan() {
        if !got.Scan() {
            t.Fatalf("%v: Results file ended, while expecting '%v'", t.Name(), want.Text())
        }
        xs := want.Text()
        ys := got.Text()
        if len(xs) != len(ys) {
            t.Fatalf("%v: Results have different length:\n\twant '%v',\n\tgot  '%v'", t.Name(), xs, ys)
        }
        for i := 0; i < len(xs); i++ {
            if xs[i] != ys[i] {
                t.Fatalf("%v: Results differ at position %v:\n\twant '%v',\n\tgot  '%v'", t.Name(), i, xs, ys)
            }
        }
    }
    if err := want.Err(); err != nil {
        t.Fatalf("%v: Error, while reading expected results file '%v': %v", t.Name(), resStore, err)
    }
}

func TestMain(m *testing.M) {
    if err := readJson(filepath.Join(testDir, "mail-lastlog.conf"), &conf); err != nil {
        panic(err)
    }
    resStore = conf.Store
    conf.Store = filepath.Join(testDir, "mail-lastlog.test.json")

    os.Remove(conf.Store)
    os.Exit(m.Run())
}

