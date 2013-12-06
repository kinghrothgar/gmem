package main

import (
	"regexp"
	"flag"
	"os"
	"io/ioutil"
	"strconv"
	"fmt"
    "errors"
    "strings"
)

// To be set at build
var buildCommit string
var buildDate string
var buildRuntime string

// Variables for flags
var (
	sizeOp		    string
	showVers        bool
)

func init() {
	// Find location of binary
	flag.StringVar(&sizeOp, "size", "m", "what units to use: k, m, g")
	flag.BoolVar(&showVers, "v", false, "show version/build information")
	flag.Parse()
}

func fatal(msg string, params ...interface{}) {
    fmt.Printf(msg + "\n", params...)
    os.Exit(1)
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func digits(number int) (count int) {
    count = 1
    for number > 0 {
        count++
        number = number / 10
    }
    return
}

func verifyInfo(info map[string]int) error {
    keys := []string{"MemFree", "Active(file)", "Inactive(file)", "SReclaimable", "MemTotal"}
    for _, key := range keys {
        if _, ok := info[key]; ok {continue}
        return errors.New("/proc/meminfo does not have the needed field '" + key + "'")
    }
    return nil
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
        fatal("Error: failed to read '/proc/meminfo' with error: %s", err)
	}
	meminfo := string(file)
	if file, err = ioutil.ReadFile("/proc/zoneinfo"); err != nil {
        fatal("Error: failed to read '/proc/zoneinfo' with error: %s", err)
	}
	zoneinfo := string(file)

	lows := regexp.MustCompile(`low\s+(\d+)`).FindAllStringSubmatch(zoneinfo, -1)
	wmarkLow := 0
	lowInt := 0
	for _, low := range lows {
		lowInt, _ = strconv.Atoi(low[1])
		wmarkLow += lowInt
	}

    info := make(map[string]int)
    matches := regexp.MustCompile(`(?m)^(\w+[^\s:]+):\s+(\d+)\s*kB`).FindAllStringSubmatch(meminfo, -1)
    for _, match := range matches {
        info[match[1]], _ = strconv.Atoi(match[2])
    }

    if err = verifyInfo(info); err != nil {
        fatal("Error %s", err.Error())
    }

    available := info["MemFree"] - wmarkLow
    pagecache := info["Active(file)"] + info["Inactive(file)"]
    pagecache -= min(pagecache / 2, wmarkLow)
    available += pagecache

    slab_reclaimable := info["SReclaimable"]
    slab_reclaimable -= min(info["SReclaimable"] / 2, wmarkLow)
    available += slab_reclaimable

    if available < 0 {
      available = 0
    }
    total := info["MemTotal"]
    used := total - available

    size := "KB"
    switch strings.ToLower(sizeOp) {
    case "m", "mb":
        available = available / 1024
        used = used / 1024
        size = "MB"
    case "g", "gb":
        available = available / (1048576)
        used = used / (1048576)
        size = "GB"
    }

    fmt.Printf("Total:     %12d %s\n", total, size)
    fmt.Printf("Used:      %12d %s\n", used, size)
    fmt.Printf("Available: %12d %s\n", available, size)
}
