package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Result struct {
	jobsNr       int
	SysStat      SysStat
	SysStatLog   SysStatLog
	FioData      Fio
	JsonPath     string
	TestName     string
	testFullName string
	TestFamily   string
}

type resultsMap map[string][]*Result

func parseResult(path string, allResults resultsMap) error {
	fmt.Printf("parsing %s\n", path)

	re := regexp.MustCompile(`\/([a-z]+)-jobs\d+-bs\d+.?-iodepth\d+`)
	testName := string(re.FindSubmatch([]byte(path))[1])

	dirName := filepath.Dir(path)
	testFamily := filepath.Base(filepath.Dir(dirName))
	testFullName := filepath.Base(dirName)
	// fmt.Printf("%s:  %s\n", path, testName)

	bytes, err := os.ReadFile(path)
	check(err)

	SysStatData, err := UnmarshalSysStat(bytes)
	check(err)

	bytes, err = os.ReadFile(fmt.Sprintf("%s/%s-guest/result.json", dirName, testFullName))
	check(err)

	jsonString := string(bytes)

	re = regexp.MustCompile(`Starting (\d+) process`)
	jobsNrStr := string(re.FindSubmatch(bytes)[1])
	jobsNr, err := strconv.Atoi(jobsNrStr)
	check(err)

	jsonStartsAt := strings.Index(jsonString, "{")
	jsonEndsAt := strings.LastIndex(jsonString, "}")
	jsonString = jsonString[jsonStartsAt : jsonEndsAt+2]
	bytes = []byte(jsonString)

	FioData, err := UnmarshalFio(bytes)
	check(err)

	bytes, err = os.ReadFile(fmt.Sprintf("%s/sys_stats_log.json", dirName))
	check(err)

	SysStatLogData, err := UnmarshalSysStatLog(bytes)
	check(err)

	Result := Result{
		jobsNr:       jobsNr,
		SysStat:      SysStatData,
		SysStatLog:   SysStatLogData,
		FioData:      FioData,
		JsonPath:     path,
		TestName:     testName,
		testFullName: testFullName,
		TestFamily:   testFamily,
	}

	allResults[testFamily] = append(allResults[testFamily], &Result)
	//	return fmt.Errorf("stop here for now")
	return nil
}

func pathWalker(path string, info os.FileInfo, allResults resultsMap) error {
	if !strings.HasSuffix(path, "sys_stats.json") {
		return nil
	}

	err := parseResult(path, allResults)

	return err
}

var header = []string{
	"Test",
	"JobsNR",
	"Depth",
	"ReadBW",
	"WriteBW",
	"Write Lat Max ms",
	"Write Lat stddev ms",
	"Write clat p99 ms",
}

func genRecord(res *Result) []string {
	testName := fmt.Sprintf("%s %s", res.TestName,
		res.FioData.Jobs[0].JobOptions.Rw)
	record := []string{testName}

	job := res.FioData.Jobs[0]
	record = append(record, job.JobOptions.Numjobs)
	record = append(record, job.JobOptions.Iodepth)
	record = append(record, fmt.Sprintf("%d", job.Read.BW))
	record = append(record, fmt.Sprintf("%d", job.Write.BW))
	record = append(record, fmt.Sprintf("%.2f", float64(job.Write.LatNS.Max)/1000000))
	record = append(record, fmt.Sprintf("%.2f", job.Write.LatNS.Stddev/1000000))
	record = append(record, fmt.Sprintf("%.2f", float64(job.Write.ClatNS.Percentile["99.000000"])/1000000))
	return record
}

func filterTests(rw string) bool {
	// if rw == "write" {
	// 	return false
	// }

	if rw == "write" {
		return false
	}

	return true
}

func printTable(allResults resultsMap) {
	outfile := "file.tsv"
	f, err := os.Create(outfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Comma = ';'

	w.Write(header)
	for k := range allResults {
		curGroup := allResults[k]
		for _, curTest := range curGroup {
			if filterTests(curTest.FioData.Jobs[0].JobOptions.Rw) {
				continue
			}
			record := genRecord(curTest)
			w.Write(record)
		}
	}
}

func getMemUsage(res *Result) (int64, int64, int64) {
	SysStatData := res.SysStat
	SysStatLogData := res.SysStatLog

	testStartedAt, err := time.Parse("2006-01-02T15:04:05-0700", SysStatData.BeforeTest.Date)
	check(err)
	testFinishedAt, err := time.Parse("2006-01-02T15:04:05-0700", SysStatData.AfterTest.Date)
	check(err)
	memTotal := SysStatData.BeforeTest.Memory.MemTotal
	freeMemMin := int64(0)
	freeMemMax := memTotal
	memSum := int64(0)
	samplesNr := int64(0)

	for _, logEntry := range SysStatLogData {
		curTime, err := time.Parse("2006-01-02T15:04:05-0700", logEntry.Date)
		check(err)

		if curTime.Before(testStartedAt) {
			continue
		}

		memSum += int64(logEntry.Memory.MemFree)
		samplesNr++

		if logEntry.Memory.MemFree < freeMemMax {
			freeMemMax = logEntry.Memory.MemFree
		}

		if logEntry.Memory.MemFree > freeMemMin {
			freeMemMin = logEntry.Memory.MemFree
		}

		if curTime.After(testFinishedAt) {
			break
		}
	}

	usedMax := memTotal - freeMemMin
	usedMin := memTotal - freeMemMax
	usedAvg := memTotal - memSum/samplesNr

	return usedMax, usedMin, usedAvg
}

func getCpuUsage(res *Result) int {
	sysStat := res.SysStat

	diffUser := sysStat.AfterTest.CPU.Usage - sysStat.BeforeTest.CPU.Usage
	diffSys := sysStat.AfterTest.CPU.System - sysStat.BeforeTest.CPU.System
	diffIdle := sysStat.AfterTest.CPU.Idle - sysStat.BeforeTest.CPU.Idle

	/* XXX: rerun tests to collect all the cpu stats from /proc/stat */
	return int((diffUser + diffSys) * 100 / (diffUser + diffSys + diffIdle))
}

func genExcelRow(res *Result, f *excelize.File, family string, row_nr int) {
	job := res.FioData.Jobs[0]
	f.SetCellValue(family, fmt.Sprintf("A%d", row_nr), res.testFullName)
	f.SetCellValue(family, fmt.Sprintf("B%d", row_nr), res.jobsNr)

	iodepth, err := strconv.Atoi(job.JobOptions.Iodepth)
	check(err)
	f.SetCellValue(family, fmt.Sprintf("C%d", row_nr), iodepth)

	f.SetCellValue(family, fmt.Sprintf("D%d", row_nr), job.Write.BW)
	f.SetCellValue(family, fmt.Sprintf("E%d", row_nr), job.Read.BW)
	f.SetCellValue(family, fmt.Sprintf("F%d", row_nr), float64(job.Write.LatNS.Max)/1000000)
	f.SetCellValue(family, fmt.Sprintf("G%d", row_nr), float64(job.Write.LatNS.Min)/1000000)
	f.SetCellValue(family, fmt.Sprintf("H%d", row_nr), float64(job.Write.LatNS.Stddev)/1000000)
	f.SetCellValue(family, fmt.Sprintf("I%d", row_nr), float64(job.Write.ClatNS.Percentile["99.000000"])/1000000)

	f.SetCellValue(family, fmt.Sprintf("J%d", row_nr), float64(job.Read.LatNS.Max)/1000000)
	f.SetCellValue(family, fmt.Sprintf("K%d", row_nr), float64(job.Read.LatNS.Min)/1000000)
	f.SetCellValue(family, fmt.Sprintf("L%d", row_nr), float64(job.Read.LatNS.Stddev)/1000000)
	f.SetCellValue(family, fmt.Sprintf("M%d", row_nr), float64(job.Read.ClatNS.Percentile["99.000000"])/1000000)

	f.SetCellValue(family, fmt.Sprintf("N%d", row_nr), job.Write.Iops)
	f.SetCellValue(family, fmt.Sprintf("O%d", row_nr), job.Read.Iops)
	//f.SetCellValue(family, fmt.Sprintf("J%d", row_nr), float64(job.Write.LatNS.Stddev)/1000000)

	memMax, memMin, memAvg := getMemUsage(res)
	f.SetCellValue(family, fmt.Sprintf("P%d", row_nr), memMax)
	f.SetCellValue(family, fmt.Sprintf("Q%d", row_nr), memMin)
	f.SetCellValue(family, fmt.Sprintf("R%d", row_nr), memAvg)

	f.SetCellValue(family, fmt.Sprintf("S%d", row_nr), getCpuUsage(res))

}

func genExcelSheet(results []*Result, f *excelize.File, family string) {
	index := f.NewSheet(family)
	f.SetCellValue(family, "A1", "Test")
	f.SetCellValue(family, "B1", "JobsNR")
	f.SetCellValue(family, "C1", "Depth")
	f.SetCellValue(family, "D1", "WriteBW KiB/s")
	f.SetCellValue(family, "E1", "ReadBW KiB/s")

	f.SetCellValue(family, "F1", "Write LatMax")
	f.SetCellValue(family, "G1", "Write LatMin")
	f.SetCellValue(family, "H1", "Write Lat stddev")
	f.SetCellValue(family, "I1", "Write cLat p99 ms")

	f.SetCellValue(family, "J1", "Read LatMax")
	f.SetCellValue(family, "K1", "Read LatMin")
	f.SetCellValue(family, "L1", "Read Lat stddev")
	f.SetCellValue(family, "M1", "Read cLat p99 ms")

	f.SetCellValue(family, "N1", "Write IOPS")
	f.SetCellValue(family, "O1", "Read IOPS")

	f.SetCellValue(family, "P1", "Mem Min")
	f.SetCellValue(family, "Q1", "Mem Max")
	f.SetCellValue(family, "R1", "Mem Avg")

	f.SetCellValue(family, "S1", "CPU %")

	row_nr := 2
	for _, test := range results {
		genExcelRow(test, f, family, row_nr)
		row_nr++
	}
	f.SetActiveSheet(index)
}

func genExcel(allResults resultsMap) {
	f := excelize.NewFile()

	for family := range allResults {
		genExcelSheet(allResults[family], f, family)
	}

	if err := f.SaveAs("restults.xlsx"); err != nil {
		fmt.Println(err)
	}
}
func main() {
	// const walkPath = "/Users/yuri/eve-fio-output"
	const walkPath = "results"

	allResults := make(resultsMap)
	err := filepath.Walk(walkPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return pathWalker(path, info, allResults)

		})
	if err != nil {
		log.Println(err)
	}

	printTable(allResults)
	genExcel(allResults)
}
