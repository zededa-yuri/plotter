package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"regexp"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type Result struct {
	FioData Fio
	JsonPath string
	TestName string
}

type resultsMap map[string][]*Result

func parseResult(path string, allResults resultsMap) error {
	fmt.Printf("parsing %s\n", path)

	re := regexp.MustCompile(`\/fio-output\/([^\/]+)\/`)
	testName := string(re.FindSubmatch([]byte(path))[1])
	// fmt.Printf("%s:  %s\n", path, testName)

	bytes, err := os.ReadFile(path)
	check(err)

	FioData, err := UnmarshalFio(bytes)
	check(err)

	Result := Result{
		FioData: FioData,
		JsonPath: path,
		TestName: testName,
	}

	allResults[testName] = append(allResults[testName], &Result)
	//	return fmt.Errorf("stop here for now")
	return nil
}

func pathWalker(path string, info os.FileInfo, allResults resultsMap) error {
	if !strings.HasSuffix(path, ".json") {
		return nil
	}

	err := parseResult(path, allResults)

	return err
}

func main() {
	const walkPath = "/Users/yuri/eve-fio-output"

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
}
