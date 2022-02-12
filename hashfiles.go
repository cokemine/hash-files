package hashfiles

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rclone/rclone/backend/onedrive/quickxorhash"
)

func ListDir(filePath string) (result []string, err error) {
	depthMap := make(map[string]uint)

	err = filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		depth := strings.Count(path, string(os.PathSeparator)) - strings.Count(filePath, string(os.PathSeparator))
		depthMap[path] = uint(depth)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			result = append(result, path)
		}
		return nil
	})

	sort.SliceStable(result, func(i, j int) bool { return depthMap[result[i]] < depthMap[result[j]] })

	return
}

func TransformPath(filePath string) string {
	return strings.Replace(filePath, "\\", "/", -1)
}

func GetMD5(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := md5.New()
	return getHash(h, file)
}

func GetSHA1(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha1.New()
	return getHash(h, file)
}

func GetSHA256(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	return getHash(h, file)
}

func GetQuickXORHash(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := quickxorhash.New()
	return getHash(h, file)
}

func getHash(h hash.Hash, file io.Reader) (string, error) {
	buf := make([]byte, 1<<20)

	for {
		n, err := file.Read(buf)
		if n == 0 {
			if err == nil {
				continue
			} else if err == io.EOF {
				break
			} else {
				return "", err
			}
		}
		io.Copy(h, bytes.NewReader(buf[:n]))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func WriteFile(filePath string, content string) (err error) {
	return ioutil.WriteFile(filePath, []byte(content), 0644)
}
