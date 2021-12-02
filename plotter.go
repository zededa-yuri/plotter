package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	// "github.com/xuri/excelize/v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Result struct {
	SysStat  SysStat
	FioData  Fio
	JsonPath string
	TestName string
}

type resultsMap map[string][]*Result

func parseResult(path string, allResults resultsMap) error {
	fmt.Printf("parsing %s\n", path)

	re := regexp.MustCompile(`\/([a-z]+)-jobs\d+-bs\d+.?-iodepth\d+`)
	testName := string(re.FindSubmatch([]byte(path))[1])

	dirName := filepath.Dir(path)
	testFamily := filepath.Base(dirName)
	// fmt.Printf("%s:  %s\n", path, testName)

	bytes, err := os.ReadFile(path)
	check(err)

	SysStatData, err := UnmarshalSysStat(bytes)
	check(err)

	bytes, err = os.ReadFile(fmt.Sprintf("%s/%s-guest/result.json", dirName, testFamily))
	check(err)

	jsonString := string(bytes)
	jsonStartsAt := strings.Index(jsonString, "{")
	jsonEndsAt := strings.LastIndex(jsonString, "}")
	jsonString = jsonString[jsonStartsAt : jsonEndsAt+2]
	bytes = []byte(jsonString)

	FioData, err := UnmarshalFio(bytes)
	check(err)

	Result := Result{
		SysStat:  SysStatData,
		FioData:  FioData,
		JsonPath: path,
		TestName: testName,
	}

	allResults[testName] = append(allResults[testName], &Result)
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

func main() {
	// const walkPath = "/Users/yuri/eve-fio-output"
	const walkPath = "results/zfs_untuned_p4"

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
}
