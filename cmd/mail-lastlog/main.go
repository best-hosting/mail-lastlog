
package main

import (
    "fmt"
    //"net/mail"
    "encoding/json"
    "flag"
    "os"
    "io"
    "path/filepath"

    //"bh/lastlog/pkg/parser"
    . "bh/lastlog/pkg/types"
    "bh/lastlog/pkg/dovecot"
    "bh/lastlog/pkg/store"
    "bh/lastlog/pkg/intervals"
    //"bh/lastlog/pkg/exim4"
)

// FIXME: Different per-user views: group all entries by IP (i.e. several IP
// may have several methods used) or split by method.


type Config struct {
    DovecotLogs []string
    Exim4Logs []string
    Store string
}

var confFile string
const (
    defConfFile = "./mail-lastlog.conf"
    confFileUsage = "Path to config file"
)

func init() {
    flag.StringVar(&confFile, "config", defConfFile, confFileUsage)
    flag.StringVar(&confFile, "c", defConfFile, confFileUsage + " (shorthand)")
}

func readJson(file string, v any) error {
    f, err := os.Open(file)
    if err != nil {
        return err
    }
    defer f.Close()

    bs, err := io.ReadAll(f)
    if err != nil {
        return err
    }

    if err := json.Unmarshal(bs, v); err != nil {
        return err
    }
    fmt.Printf("readJson(): Read from '%v':\n%#v\n", file, v)

    return nil
}

func writeJson(file string, v any) error {
    bs, err := json.MarshalIndent(v, "", "\t")
    if err != nil {
        return err
    }
    fmt.Printf("writeJson(): Writing to '%v':\n%s\n", file, string(bs))

    f, err := os.CreateTemp(filepath.Dir(file), "mail-lastlog")
    if err != nil {
        return err
    }
    defer os.Remove(f.Name())

    if _, err := f.Write(bs); err != nil {
        return err
    }
    defer f.Close()

    if err := os.Rename(f.Name(), file); err != nil {
        return err
    }
    return nil
}

// TODO: Write tests running against test/data .
func main() {
    flag.Parse()


    fmt.Printf("main(): Use config file '%v''\n", confFile)
    conf := Config{}
    if err := readJson(confFile, &conf); err != nil {
        panic(err)
    }

    fmt.Printf("main(): Use db file '%v'\n", conf.Store)
    store := store.New()
    readJson(conf.Store, store)

    if _, ok := store.Intervals["Dovecot"]; !ok {
        store.Intervals["Dovecot"] = &intervals.Intervals[Time, Result]{}
    }
    dl := dovecot.NewLog(store.Intervals["Dovecot"])
    store.ReadLogs(dl, conf.DovecotLogs)

    if err := writeJson(conf.Store, store); err != nil {
        panic(err)
    }
}
