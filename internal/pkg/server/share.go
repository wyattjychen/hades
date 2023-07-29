package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

	// 只有指定-v时才会显示版本，暂时不需要显示版本，可以除去
	if NodeConfigOps.Version {
		if isNode {
			fmt.Printf("%s Version:%s\n", NodeModule, Version)
		} else {
			fmt.Printf("%s Version:%s\n", MasterModule, Version)
		}
		os.Exit(0)
	}

	var nodeType string
	if isNode {
		nodeType = "node"
	} else {
		nodeType = "master"
	}

	// TODO:pprof
	// if apiConfigOptions.EnablePProfile {
	// 	go func() {
	// 		fmt.Printf("enable pprof http server at:%d\n", apiConfigOptions.PProfilePort)
	// 		fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", apiConfigOptions.PProfilePort), nil))
	// 	}()
	// }
	// TODO:HealthCheck

	// TODO:区分环境 目前先不区分
	//var env = config.Env(NodeConfigOps.Environment)

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