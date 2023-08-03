package server

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
)

const (
	MasterModule = "hades/master"
	NodeModule   = "hades/node"

	shutdownMaxAge = 10 * time.Second
	shutdownWait   = 1500 * time.Millisecond
)

type NodeConfigOptions struct {
	flags.Options
	Master         string `short:"n" long:"stype"  description:"Master or Node"`
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
