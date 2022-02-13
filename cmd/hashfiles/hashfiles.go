package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	hashfiles "github.com/cokemine/hash-files"
)

var (
	dirPath     = flag.String("dir", "", "The directory to be hashed")
	algo        = flag.String("algo", "md5", "The hash algorithm to use, multiple algorithms can be specified by comma separated")
	parallelNum = flag.Int("parallel", runtime.NumCPU(), "The number of parallel workers")
	verbose     = flag.Bool("verbose", false, "Verbose output log")
)

func main() {
	flag.Parse()

	if *dirPath == "" {
		*dirPath, _ = os.Getwd()
	}

	*dirPath = strings.TrimSuffix(*dirPath, string(os.PathSeparator)) + string(os.PathSeparator)

	files, err := hashfiles.ListDir(*dirPath)
	if err != nil {
		log.Fatal(err)
	}
	var timeConsumed int64

	for _, algo := range strings.Split(*algo, ",") {
		var hashFn func(filePath string) (string, error)
		var wg sync.WaitGroup
		algo := strings.TrimSpace(algo)

		switch algo {
		case "md5":
			hashFn = hashfiles.GetMD5
		case "sha1":
			hashFn = hashfiles.GetSHA1
		case "sha256":
			hashFn = hashfiles.GetSHA256
		case "quickxorhash":
			hashFn = hashfiles.GetQuickXORHash
		default:
			log.Fatal(errors.New("unsupported algorithm"))
		}

		result := make([]string, len(files))
		fileWorkers := SliceChunk(files, *parallelNum)

		timeStart := time.Now().Unix()

		for i := range fileWorkers {
			l := len(fileWorkers[i])
			wg.Add(l)
			for j := range fileWorkers[i] {
				go func(j int, file string) {
					defer wg.Done()

					encoded, err := hashFn(file)
					if err != nil {
						panic(err)
					}

					fileName := hashfiles.TransformPath(strings.TrimPrefix(file, *dirPath))
					result[i*l+j] = fmt.Sprintf("%s %s\n", encoded, fileName)

					if *verbose {
						fmt.Printf("%s [%d,%d]: %s, result: %s\n", algo, i*l+j+1, len(files), fileName, encoded)
					}

				}(j, fileWorkers[i][j])
			}
			wg.Wait()
		}

		timeEnd := time.Now().Unix()
		timeConsumed += timeEnd - timeStart
		log.Printf("%s: %d files, %d seconds\n", algo, len(files), timeEnd-timeStart)

		err := hashfiles.WriteFile(filepath.Join(*dirPath, fmt.Sprintf("%ssum.txt", algo)), strings.Join(result, "\n"))
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("All files hashed, %d seconds\n", timeConsumed)
}

func SliceChunk(s []string, size int) [][]string {
	var ret [][]string
	for size < len(s) {
		s, ret = s[size:], append(ret, s[:size:size])
	}
	ret = append(ret, s)
	return ret
}
