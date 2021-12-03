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

func getMemUser(res *Result) (int64, int64, int64) {
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

func getCpuUser(res *Result) int {
	sysStat := res.SysStat

	before := sysStat.BeforeTest.CPU
	after := sysStat.AfterTest.CPU

	totalBefore := before.User +
		before.Nice +
		before.System +
		before.Idle +
		before.Iowait +
		before.IRQ +
		before.Softirq +
		before.Steal +
		before.Guest

	totalAfter := after.User +
		after.Nice +
		after.System +
		after.Idle +
		after.Iowait +
		after.IRQ +
		after.Softirq +
		after.Steal +
		after.Guest

	usedBefore := totalBefore - before.Idle - before.Iowait
	usedAfter := totalAfter - after.Idle - after.Iowait

	return int((usedAfter - usedBefore) * 100 / (totalAfter - totalBefore))
}

func genExcelRow(headers *headersMap, res *Result, f *excelize.File, family string, row_nr int) {
	job := res.FioData.Jobs[0]
	f.SetCellValue(family, getCell(headers, "Test", row_nr), res.testFullName)
	f.SetCellValue(family, getCell(headers, "rw", row_nr), res.FioData.GlobalOptions.RW)
	f.SetCellValue(family, getCell(headers, "JobsNR", row_nr), res.jobsNr)

	iodepth, err := strconv.Atoi(job.JobOptions.Iodepth)
	check(err)
	f.SetCellValue(family, getCell(headers, "Depth", row_nr), iodepth)

	f.SetCellValue(family, getCell(headers, "WriteBW KiB/s", row_nr), job.Write.BW)
	f.SetCellValue(family, getCell(headers, "ReadBW KiB/s", row_nr), job.Read.BW)
	f.SetCellValue(family, getCell(headers, "Write LatMax", row_nr), float64(job.Write.LatNS.Max)/1000000)
	f.SetCellValue(family, getCell(headers, "Write LatMin", row_nr), float64(job.Write.LatNS.Min)/1000000)
	f.SetCellValue(family, getCell(headers, "Write Lat stddev", row_nr), float64(job.Write.LatNS.Stddev)/1000000)
	f.SetCellValue(family, getCell(headers, "Write cLat p99 ms", row_nr), float64(job.Write.ClatNS.Percentile["99.000000"])/1000000)

	f.SetCellValue(family, getCell(headers, "Read LatMax", row_nr), float64(job.Read.LatNS.Max)/1000000)
	f.SetCellValue(family, getCell(headers, "Read LatMin", row_nr), float64(job.Read.LatNS.Min)/1000000)
	f.SetCellValue(family, getCell(headers, "Read Lat stddev", row_nr), float64(job.Read.LatNS.Stddev)/1000000)
	f.SetCellValue(family, getCell(headers, "Read cLat p99 ms", row_nr), float64(job.Read.ClatNS.Percentile["99.000000"])/1000000)

	f.SetCellValue(family, getCell(headers, "Write IOPS", row_nr), job.Write.Iops)
	f.SetCellValue(family, getCell(headers, "pRead IOPS", row_nr), job.Read.Iops)

	memMax, memMin, memAvg := getMemUser(res)
	f.SetCellValue(family, getCell(headers, "Mem Min", row_nr), memMax)
	f.SetCellValue(family, getCell(headers, "Mem Max", row_nr), memMin)
	f.SetCellValue(family, getCell(headers, "Mem Avg", row_nr), memAvg)

	f.SetCellValue(family, getCell(headers, "CPU %", row_nr), getCpuUser(res))
}

type headersMap map[string]string

func addHead(f *excelize.File, headers *headersMap, head string, sheet string) {
	nextIndex := len(*headers) + 1
	column, err := excelize.ColumnNumberToName(nextIndex)
	check(err)
	(*headers)[head] = column

	f.SetCellValue(sheet, fmt.Sprintf("%s1", column), head)
}

func getCell(headers *headersMap, header string, row int) string {
	column := (*headers)[header]
	return fmt.Sprintf("%s%d", column, row)
}

func genExcelSheet(results []*Result, f *excelize.File, family string) {
	index := f.NewSheet(family)
	headers := make(headersMap)
	addHead(f, &headers, "Test", family)

	addHead(f, &headers, "rw", family)

	addHead(f, &headers, "JobsNR", family)
	addHead(f, &headers, "Depth", family)
	addHead(f, &headers, "WriteBW KiB/s", family)
	addHead(f, &headers, "ReadBW KiB/s", family)

	addHead(f, &headers, "Write LatMax", family)
	addHead(f, &headers, "Write LatMin", family)
	addHead(f, &headers, "Write Lat stddev", family)
	addHead(f, &headers, "Write cLat p99 ms", family)

	addHead(f, &headers, "Read LatMax", family)
	addHead(f, &headers, "Read LatMin", family)
	addHead(f, &headers, "Read Lat stddev", family)
	addHead(f, &headers, "Read cLat p99 ms", family)

	addHead(f, &headers, "Write IOPS", family)
	addHead(f, &headers, "pRead IOPS", family)

	addHead(f, &headers, "Mem Min", family)
	addHead(f, &headers, "Mem Max", family)
	addHead(f, &headers, "Mem Avg", family)

	addHead(f, &headers, "CPU %", family)

	row_nr := 2
	for _, test := range results {
		genExcelRow(&headers, test, f, family, row_nr)
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
	const walkPath = "results-proper-cpu"

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
