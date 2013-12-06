package main

import (
	"regexp"
	"flag"
	"os"
	"io/ioutil"
	"log"
	"strconv"
	"fmt"
)

// To be set at build
var buildCommit string
var buildDate string
var buildRuntime string

// Variables for flags
var (
	available	bool
	used		bool
	size		string
	showVers    bool
)

func init() {
	// Find location of binary
	flag.BoolVar(&available, "available", false, "print out available memory")
	flag.BoolVar(&used, "used", false, "print out used memory")
	flag.StringVar(&size, "size", "k", "what units to use: k, m, g")
	flag.BoolVar(&showVers, "v", false, "show version/build information")
	flag.Parse()
}

func main() {
	if showVers {
		println("Commit:  " + buildCommit)
		println("Date:    " + buildDate)
		println("Runtime: " + buildRuntime)
		os.Exit(0)
	}
	var (
		file		[]byte
		err			error
	)
	if file, err = ioutil.ReadFile("/proc/meminfo"); err != nil {
		log.Fatal("Failed to read '/proc/meminfo' with error: %s", err)
	}
	//meminfo := string(file)
	if file, err = ioutil.ReadFile("/proc/zoneinfo"); err != nil {
		log.Fatal("Failed to read '/proc/zoneinfo' with error: %s", err)
	}
	zoneinfo := string(file)

	reg := regexp.MustCompile("low\\s+(\\d+)")
	lows := reg.FindAllStringSubmatch(zoneinfo, -1)
	wmarkLow := 0
	lowInt := 0
	for _, low := range lows {
		if lowInt, err = strconv.Atoi(low[0]); err != nil {
			log.Fatal("Failed to parse watermark lows from /proc/zoneinfo")
		}
		wmarkLow += lowInt
	}
	fmt.Println(wmarkLow)

}
