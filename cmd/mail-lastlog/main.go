
package main

import (
    //"net/mail"
    "encoding/json"
    "flag"
    "os"
    "io"
    "path/filepath"

    //"bh/lastlog/pkg/parser"
    . "bh/lastlog/pkg/common"
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
    LogDfn("Read from '%v':\n\t%#v", file, v)

    return nil
}

func writeJson(file string, v any) error {
    bs, err := json.MarshalIndent(v, "", "\t")
    if err != nil {
        return err
    }
    LogDfn("Writing to '%v':\n%s", file, string(bs))

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

func run(conf Config) error {
    LogDfn("Use db file '%v'", conf.Store)
    store := store.New()
    readJson(conf.Store, store)

    if _, ok := store.Intervals["Dovecot"]; !ok {
        store.Intervals["Dovecot"] = &intervals.Intervals[Time, Result]{}
    }
    dl := dovecot.NewLog(store.Intervals["Dovecot"])
    store.ReadLogs(dl, conf.DovecotLogs)

    if err := writeJson(conf.Store, store); err != nil {
        return err
    }
    return nil
}

func main() {
    flag.Parse()

    LogDfn("Use config file '%v''", confFile)
    conf := Config{}
    if err := readJson(confFile, &conf); err != nil {
        panic(err)
    }
    run(conf)
}

