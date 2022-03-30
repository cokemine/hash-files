package main

import (
	"bufio"
	"fmt"
	"github.com/cokemine/hashfiles/pkg/hashfiles"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Empty struct{}

func getHashFn(algo string) func(filePath string) (string, error) {
	switch algo {
	case "md5":
		return hashfiles.GetMD5
	case "sha1":
		return hashfiles.GetSHA1
	case "sha256":
		return hashfiles.GetSHA256
	case "quickxorhash":
		return hashfiles.GetQuickXORHash
	default:
		log.Fatalf("unsupported algorithm: %s", algo)
	}
	return nil
}

func main() {
	pwd, _ := os.Getwd()

	var (
		algo        string
		dirPath     string
		parallelNum int
		verbose     bool
	)

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "dir",
			Aliases:     []string{"d"},
			Value:       pwd,
			Usage:       "The directory to be hashed",
			Destination: &dirPath,
		},
		&cli.StringFlag{
			Name:        "algo",
			Aliases:     []string{"a"},
			Value:       "md5",
			Usage:       "The hash algorithm to use, multiple algorithms can be specified by comma separated",
			Destination: &algo,
		},
		&cli.IntFlag{
			Name:        "parallel",
			Aliases:     []string{"n"},
			Value:       runtime.NumCPU(),
			Usage:       "The number of parallel workers",
			Destination: &parallelNum,
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Value:       false,
			Usage:       "Verbose output log",
			Destination: &verbose,
		},
	}

	var timeConsumed int64
	timeStart := time.Now().Unix()

	app := &cli.App{
		Name:  "hashfiles",
		Usage: "Recursively generate checksum of all files in a directory",
		Flags: flags,
		Commands: []*cli.Command{
			{
				Name:  "hash",
				Usage: "Hash files",
				Flags: flags,
				Action: func(c *cli.Context) error {
					dirPath = strings.TrimSuffix(dirPath, string(os.PathSeparator)) + string(os.PathSeparator)
					algo = strings.ToLower(algo)
					files, err := hashfiles.ListDir(dirPath)
					if err != nil {
						log.Fatal(err)
					}

					for _, algo := range strings.Split(algo, ",") {
						algo := strings.TrimSpace(algo)

						filesChan, wg := make(chan Empty, parallelNum), sync.WaitGroup{}
						result := make([]string, len(files))

						hashFn := getHashFn(algo)

						for i := range files {
							wg.Add(1)
							filesChan <- Empty{}

							go func(i int, file string) {
								defer func() {
									wg.Done()
									<-filesChan
								}()

								encoded, err := hashFn(file)
								if err != nil {
									log.Fatal(err)
								}

								fileName := hashfiles.TransformPath(strings.TrimPrefix(file, dirPath))
								result[i] = fmt.Sprintf("%s %s\n", encoded, fileName)

								if verbose {
									log.Printf("%s [%d,%d]: %s, result: %s\n", algo, i+1, len(files), fileName, encoded)
								}
							}(i, files[i])
						}

						wg.Wait()

						timeEnd := time.Now().Unix()
						timeConsumed += timeEnd - timeStart
						log.Printf("%s: %d files, %d seconds\n", algo, len(files), timeEnd-timeStart)

						err := hashfiles.WriteFile(filepath.Join(dirPath, fmt.Sprintf("%ssum.txt", algo)), strings.Join(result, "\n"))
						if err != nil {
							log.Fatal(err)
						}
					}

					log.Printf("All files hashed, %d seconds\n", timeConsumed)

					return nil
				},
			},
			{
				Name:  "verify",
				Usage: "Verify files",
				Flags: flags,
				Action: func(c *cli.Context) error {
					dirPath = strings.TrimSuffix(dirPath, string(os.PathSeparator)) + string(os.PathSeparator)
					algo = strings.ToLower(algo)

					for _, algo := range strings.Split(algo, ",") {
						algo := strings.TrimSpace(algo)
						hashFn := getHashFn(algo)
						sumFile := path.Join(dirPath, fmt.Sprintf("%ssum.txt", algo))
						file, err := os.Open(sumFile)
						if err != nil {
							log.Fatal(err)
						}
						r := bufio.NewReader(file)
						i, tot, unMatched := 0, 0, 0

						filesChan, wg := make(chan Empty, parallelNum), sync.WaitGroup{}

						for {
							line, err := r.ReadString('\n')
							line = strings.TrimSpace(line)
							if err == io.EOF {
								break
							} else if err != nil {
								log.Fatal(err)
							}
							if line == "" {
								continue
							}
							i = i + 1

							wg.Add(1)
							filesChan <- Empty{}

							go func(i int, line string) {
								defer func() {
									wg.Done()
									<-filesChan
								}()

								sum := strings.SplitN(line, " ", 2)
								fileHash, err := hashFn(path.Join(dirPath, sum[1]))
								if err != nil {
									log.Fatal(err)
								}
								if fileHash != sum[0] {
									unMatched++
									log.Printf("[Not Matched] %s [%d]: %s\n", algo, i, sum[1])
								} else if verbose {
									log.Printf("[Matched] %s [%d]: %s\n", algo, i, sum[1])
								}
								tot++
							}(i, line)
						}

						wg.Wait()

						timeEnd := time.Now().Unix()
						timeConsumed += timeEnd - timeStart
						log.Printf("%s: %d files, %d seconds, UnMatched count: %d\n", algo, tot, timeEnd-timeStart, unMatched)

						err = file.Close()

						if err != nil {
							log.Fatal(err)
						}

					}

					log.Printf("All files Verified, %d seconds\n", timeConsumed)

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
