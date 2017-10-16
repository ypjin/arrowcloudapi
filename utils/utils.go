package utils

import (
	"arrowcloudapi/utils/log"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jeffail/gabs"
)

// FormatEndpoint formats endpoint
func FormatEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(endpoint, "http://") &&
		!strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return endpoint
}

// ParseEndpoint parses endpoint to a URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = FormatEndpoint(endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ParseRepository splits a repository into two parts: project and rest
func ParseRepository(repository string) (project, rest string) {
	repository = strings.TrimLeft(repository, "/")
	repository = strings.TrimRight(repository, "/")
	if !strings.ContainsRune(repository, '/') {
		rest = repository
		return
	}
	index := strings.LastIndex(repository, "/")
	project = repository[0:index]
	rest = repository[index+1:]
	return
}

// GenerateRandomString generates a random string
func GenerateRandomString() string {
	length := 32
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func PrettyPrintObject(obj interface{}) (err error) {

	//fmt.Printf("%+v\n", obj);
	objJson, err := json.Marshal(obj)
	//fmt.Println(string(objJson))
	if err != nil {
		return
	}

	objJson2, err := gabs.ParseJSON(objJson)
	if err != nil {
		return
	}

	log.Debug(objJson2.StringIndent("", "  "))
	return
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Debug(string(b))
}

func SaveFile(fileName string, byteData []byte) error {

	parentDir := filepath.Dir(fileName)
	_, err := os.Stat(parentDir)
	if err != nil {
		err := os.MkdirAll(parentDir, 0777)
		if err != nil {
			return err
		}
	}

	fo, err := os.Create(fileName)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create file %v, %v", fileName, err))
	}
	if _, err := fo.Write(byteData); err != nil {
		return errors.New(fmt.Sprintf("Failed to save file, %v", err))
	}

	return nil
}

// Check if a exists in the provided list
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StringInSliceWithIndex(a string, list []string) (int, bool) {
	for i, b := range list {
		if b == a {
			return i, true
		}
	}
	return -1, false
}

func RemoveElementInSlice(index int, list []string) []string {
	return append(list[:index], list[index+1:]...)
}
