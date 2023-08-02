package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/jessevdk/go-flags"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

const (
	Version      = "v0.0.1"
	MasterModule = "hades/master"
	NodeModule   = "hades/node"

	shutdownMaxAge = 10 * time.Second
	shutdownWait   = 1500 * time.Millisecond
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

type NodeConfigOptions struct {
	flags.Options
	Master         string `short:"n" long:"stype"  description:"Master or Node"`
	Environment    string `short:"e" long:"env" description:"Use ApiServer environment" default:"testing"`
	EnablePProfile bool   `short:"p" long:"enable-pprof"  description:"enable pprof"`
	PProfilePort   int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8188"`
	ConfigFileName string `short:"c" long:"config" description:"Use ApiServer config file" default:""`
	Balance        int    `short:"b" long:"balance-algorithm" description:"balance algorithm" default:"0"`
}

func formatTime(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}

func NewNode() (string, string, error) {
	parser := flags.NewParser(&NodeConfigOps, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return "", "", err
	}
	isNode := false
	if NodeConfigOps.Master == "" {
		fmt.Println("no nodetype get, start node default.")
		isNode = true
	} else if NodeConfigOps.Master == "master" {
		fmt.Println("start hades master.")
	} else if NodeConfigOps.Master == "node" {
		fmt.Println("start hades node.")
		isNode = true
	} else {
		fmt.Println("please input correct node type, master or node.")
		os.Exit(0)
	}

	var nodeType string
	if isNode {
		nodeType = "node"
	} else {
		nodeType = "master"
	}

	// pprof
	if NodeConfigOps.EnablePProfile {
		go func() {
			fmt.Printf("enable pprof http server at:%d\n", NodeConfigOps.PProfilePort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", NodeConfigOps.PProfilePort), nil))
		}()
	}

	var configFile = NodeConfigOps.ConfigFileName

	return nodeType, configFile, nil

}

func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}
