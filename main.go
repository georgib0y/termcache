package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Config struct {
	cacheDir string
}

var versionFlag bool

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Shows termcache version")
	flag.BoolVar(&versionFlag, "v", false, "Shows termcache version")
}

func main() {
	flag.Parse()

	cfg := Config{
		cacheDir: "cache",
	}

	if versionFlag {
		fmt.Println("v0.0.1")
		return
	}

	err := os.MkdirAll(cfg.cacheDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	cmdHash, err := hashCommand(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	log.Println(cmdHash)

	ok, err := writeCached(cfg, cmdHash, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	if ok {
		return
	}

	log.Println("not found")

	out, err := runCommand(flag.Args())
	if err != nil {
		log.Fatal(err)
	}

	if err = writeNewCache(cfg, cmdHash, out); err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
}

func hashCommand(args []string) (string, error) {
	buf := bytes.Buffer{}
	b64enc := base64.NewEncoder(base64.StdEncoding, &buf)

	for _, a := range args {
		_, err := io.WriteString(b64enc, a)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// writes to w if hash is found in db, returns true if written
func writeCached(cfg Config, cmdHash string, w io.Writer) (bool, error) {
	file, err := os.Open(path.Join(cfg.cacheDir, cmdHash))

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if _, err = io.Copy(w, file); err != nil {
		return false, err
	}

	return true, nil
}

func writeNewCache(cfg Config, cmdHash, value string) error {
	file, err := os.Create(path.Join(cfg.cacheDir, cmdHash))
	if err != nil {
		return err
	}

	_, err = io.WriteString(file, value)
	return err
}

func runCommand(args []string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)

	out := strings.Builder{}

	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	return out.String(), err
}
