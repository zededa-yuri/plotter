package main

import (
	"fmt"
	"path/filepath"
	"log"
	"os"
)

func pathWalker(path string, info os.FileInfo) error {
	fmt.Printf("visiting file %s\n", path)
	return nil
}

func main() {
	const walkPath = "/Users/yuri/eve-fio-output"

	err := filepath.Walk(walkPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			pathWalker(path, info)
			return nil
		})
	if err != nil {
		log.Println(err)
	}
}
