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
	help		bool
	showVers	bool
	sizeOp		string
)

func init() {
	flag.StringVar(&sizeOp, "s", "k", "what units to use: k, m, g")
	flag.BoolVar(&showVers, "v", false, "show version/build information")
	flag.BoolVar(&help, "help", false, "display this help text")
	flag.Parse()
}

func showVersion() {
	println("Commit:  " + buildCommit)
	println("Date:    " + buildDate)
	println("Runtime: " + buildRuntime)
	os.Exit(0)
}

func usage() {
	fmt.Fprint(os.Stderr, "Usage: gmem [options]\n")
	fmt.Fprint(os.Stderr, "Display the amount of memory total, free, available (before swapping starts), and unavailable\n\n")
	fmt.Fprint(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	os.Exit(0)
}

func fatalExit(msg string, params ...interface{}) {
    fmt.Printf(msg + "\n", params...)
    os.Exit(1)
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

func verifyInfo(info map[string]int) error {
    keys := []string{"MemFree", "Active(file)", "Inactive(file)", "SReclaimable", "MemTotal"}
    for _, key := range keys {
        if _, ok := info[key]; ok {continue}
        return errors.New("'/proc/meminfo' does not have the needed field '" + key + "'")
    }
    return nil
}


func main() {
	if help {usage()} // Does not return
	if showVers {showVersion()} //Does not return

	var (
		file	[]byte
		err		error
	)
	// Read in meminfo and zoneinfo
	if file, err = ioutil.ReadFile("/proc/meminfo"); err != nil {
        fatalExit("Error: failed to read '/proc/meminfo' with error: %s", err)
	}
	meminfo := string(file)
	if file, err = ioutil.ReadFile("/proc/zoneinfo"); err != nil {
        fatalExit("Error: failed to read '/proc/zoneinfo' with error: %s", err)
	}
	zoneinfo := string(file)

	// Parse and verify meminfo
    info := make(map[string]int)
    matches := regexp.MustCompile(`(?m)^(\w+[^\s:]+):\s+(\d+)\s*kB`).FindAllStringSubmatch(meminfo, -1)
    for _, match := range matches {
        info[match[1]], _ = strconv.Atoi(match[2])
    }
    if err = verifyInfo(info); err != nil {
        fatalExit("Error %s", err.Error())
    }

	// Calculate low watermark
	lows := regexp.MustCompile(`low\s+(\d+)`).FindAllStringSubmatch(zoneinfo, -1)
	wmarkLow := 0
	lowInt := 0
	for _, low := range lows {
		lowInt, _ = strconv.Atoi(low[1])
		wmarkLow += lowInt
	}

	// Estimate the ammount of memory available for userspace allocations, without causing swapping
	// Free memory cannot be taken below the low watermark, before the system starts swapping.
	memFree := info["MemFree"]
    memAvailable := memFree - wmarkLow

	// Not all the page cache can be freed, otherwise the system will start swapping. 
	// Assume at least half of the page cache, or the low watermark worth of cache, needs to stay.
    pagecache := info["Active(file)"] + info["Inactive(file)"]
    pagecache -= min(pagecache / 2, wmarkLow)
    memAvailable += pagecache

	// Part of the reclaimable swap consists of items that are in use, and cannot be freed.
	// Cap this estimate at the low watermark.
    slab_reclaimable := info["SReclaimable"]
    slab_reclaimable -= min(slab_reclaimable / 2, wmarkLow)
    memAvailable += slab_reclaimable

    if memAvailable < 0 {
      memAvailable = 0
    }
    memTotal := info["MemTotal"]
    memUnavailable := memTotal - memAvailable

    size := "kB"
    switch strings.ToLower(sizeOp) {
    case "m", "mb":
		memTotal = memTotal / 1024
		memFree = memFree / 1024
        memAvailable = memAvailable / 1024
        memUnavailable = memUnavailable / 1024
        size = "MB"
    case "g", "gb":
		memTotal = memTotal / 1048576
		memFree = memFree / 1048576
        memAvailable = memAvailable / 1048576
        memUnavailable = memUnavailable / 1048576
        size = "GB"
    }

    fmt.Printf("MemTotal:       %8d %s\n", memTotal, size)
    fmt.Printf("MemFree:        %8d %s\n", memFree, size)
    fmt.Printf("MemAvailable:   %8d %s\n", memAvailable, size)
    fmt.Printf("MemUnavailable: %8d %s\n", memUnavailable, size)
}
