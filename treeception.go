package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
)

var targetfile = flag.String("file", "/fs/home/kerola/rio_testdata/utf-8-keys.txt", "file containing the data")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to file")
var checkafter = flag.Bool("check", false, "check data after sorting")
var threads = flag.Int("threads", runtime.NumCPU(), "amount of threads to use")
var serial = flag.Bool("serial", false, "Use non-paraller version of the algorithm")

var start, read, parsed, end, checked time.Time
var cpuprof, memprof *os.File

var targetdepth uint

func init() {
	flag.Parse()

	runtime.GOMAXPROCS(*threads)
	*threads = runtime.GOMAXPROCS(-1)
	var i, count uint
	for count < uint(*threads+1) {
		count = 1 << i
		i++
	}
	targetdepth = i
	println(*threads, "/", runtime.NumCPU(), "depth:", targetdepth)
}

func main() {
	if *cpuprofile != "" {
		createCPUProf()
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		createMemoryProf()
		defer memprof.Close()
	}

	start = time.Now()
	f := Read(*targetfile)
	read = time.Now()
	fmt.Printf("File %s read in %v\n", *targetfile, read.Sub(start))

	rows := Cut(f)
	parsed = time.Now()
	fmt.Printf("File parsed into string array in %v\n", parsed.Sub(read))

	sorted := MergeSort(rows)
	end = time.Now()
	fmt.Printf("File sorted in %v\n", end.Sub(parsed))

	fmt.Printf("Total: %v\n", end.Sub(start))

	if memprof != nil {
		pprof.WriteHeapProfile(memprof)
	}

	if *checkafter {
		CheckOrder(sorted)
	}

}

// Goes through the sorted pointers, if any string should come after the next one, data is not properly sorted.
func CheckOrder(data []*string) {
	fmt.Println("Checking data")
	temp := *data[0]
	for _, s := range data {
		if temp > *s {
			log.Fatalf("%s should come before %s", temp, *s)
			return
		}
		temp = *s
	}
	fmt.Println("All fine.")
	checked = time.Now()
	fmt.Printf("Data checked in %v\n", checked.Sub(end))
}

// Reads file into a slice of bytes.
func Read(target string) []byte {
	var data []byte
	var err error

	if data, err = ioutil.ReadFile(target); err != nil {
		log.Fatalf("Error reading %s: %s", target, err)
	}
	return data
}

// Byte slice to slice of strings.
func Cut(data []byte) (rows []string) {
	return strings.Split(string(data), "\n")
}

// Write CPU profile data for debugging purposes
func createCPUProf() {
	var err error
	cpuprof, err = os.Create(*cpuprofile)
	if err != nil {
		log.Fatalf("%s", err.Error())
		return
	}
	pprof.StartCPUProfile(cpuprof)
}

// Write memory profile data for debugging purposes
func createMemoryProf() {
	var err error
	memprof, err = os.Create(*memprofile)
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
}

// Sort using the sort found in standard library.
func Sort(data []string) {
	sort.Strings(data)
}
