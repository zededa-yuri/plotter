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


func parseResult(path string) error {
	// fmt.Printf("parsing %s\n", path)


	// bytes, err := os.ReadFile(path)
	// check(err)

	// result, err := UnmarshalFio(bytes)

	// fmt.Printf("version is %s\n", result.FioVersion)
	// check(err)

	re := regexp.MustCompile(`\/fio-output\/([^\/]+)\/`)
	testName := string(re.FindSubmatch([]byte(path))[1])
	fmt.Printf("%s:  %s\n", path, testName)

	//	return fmt.Errorf("stop here for now")
	return nil
}

func pathWalker(path string, info os.FileInfo) error {
	if !strings.HasSuffix(path, ".json") {
		return nil
	}

	err := parseResult(path)

	return err
}

func main() {
	const walkPath = "/Users/yuri/eve-fio-output"

	err := filepath.Walk(walkPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return pathWalker(path, info)

		})
	if err != nil {
		log.Println(err)
	}
}
