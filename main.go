package main

import (
	"time"
	"fmt"
	"os"
	"sync"
	"bufio"
	"io"
	"strings"
	"math"
	"regexp"
	"flag"
);

/* 
	Thanks to Ohm Patel for his post on parsing big files in Go. Got my started quickly \m/
	https://medium.com/swlh/processing-16gb-file-in-seconds-go-lang-3982c235dfa2
*/

type options struct {
	hasMinLength bool
	minLength int

	hasMaxLength bool
	maxLength int

	hasSpecialCharReq bool
	minSpecialChars int
} 

func main() {

	s := time.Now()
	
	fileName := flag.String("f", "", "Password list")
	minLength := flag.Int("min", 0, "Minimum Length")
	maxLength := flag.Int("max", 0, "Maximum Length")
	specialCharLengthReq := flag.Int("sp", 0, "Minimum amount of special chars required")
	flag.Parse()

	if *maxLength > 0 { 
		if *minLength > *maxLength {
			fmt.Println("Min length must be less|equal to max length")
			return
		}	
	}

	opts := &options{}
	opts.hasMinLength = *minLength > 0
	opts.minLength = *minLength
	opts.hasMaxLength = *maxLength > 0
	opts.maxLength = *maxLength
	opts.hasSpecialCharReq = *specialCharLengthReq > 0
	opts.minSpecialChars = *specialCharLengthReq

	// begin
	file, err := os.Open(*fileName)
	
	if err != nil {
		fmt.Println("Cannot read the file", err)
		fmt.Println(*fileName)
		return
	}
	
	defer file.Close() //close after checking err

	filestat, err := file.Stat()
	if err != nil {
		fmt.Println("Could not get the file stat")
		return
	}

	fileSize := filestat.Size()
	offset := fileSize - 1
	lastLineSize := 0

	for {
		b := make([]byte, 1)
		n, err := file.ReadAt(b, offset)
		if err != nil {
			fmt.Println("Error reading file ", err)
			break
		}
		char := string(b[0])
		if char == "\n" {
			break
		}
		offset--
		lastLineSize += n
	}

	lastLine := make([]byte, lastLineSize)
	_, err = file.ReadAt(lastLine, offset+1)

	if err != nil {
		fmt.Println("Could not able to read last line with offset", offset, "and lastline size", lastLineSize)
		return
	}

	Process(file, opts)

	fmt.Println("\nTime taken - ", time.Since(s))
}

func Process(f *os.File, opts *options) error {

	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, 250*1024)
		return lines
	}}

	stringPool := sync.Pool{New: func() interface{} {
		lines := ""
		return lines
	}}

	r := bufio.NewReader(f)

	var wg sync.WaitGroup

	for {
		buf := linesPool.Get().([]byte)

		n, err := r.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err != nil {
				fmt.Println(err)
				break
			}
			if err == io.EOF {
				break
			}
			return err
		}

		nextUntillNewline, err := r.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntillNewline...)
		}

		wg.Add(1)
		go func() {
			ProcessChunk(buf, &linesPool, &stringPool, opts)
			wg.Done()
		}()

	}

	wg.Wait()
	return nil
}

func ProcessChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool, opts *options) {

	var wg2 sync.WaitGroup

	logs := stringPool.Get().(string)
	logs = string(chunk)

	linesPool.Put(chunk)

	logsSlice := strings.Split(logs, "\n")

	stringPool.Put(logs)

	chunkSize := 300
	n := len(logsSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {
		wg2.Add(1)
		go func(s int, e int) {
			defer wg2.Done() //to avoid deadlocks
			for i := s; i < e; i++ {
				text := logsSlice[i]
				if len(text) == 0 {
					continue
				}

				currentLength := len(text)
				isValid := true

				if opts.hasMinLength && currentLength < opts.minLength {
					isValid = false
					continue
				} 
				
				if opts.hasMaxLength && currentLength > opts.maxLength {
					isValid = false
					continue
				}

				if opts.hasSpecialCharReq {
					var re = regexp.MustCompile(`(?m)([^A-Za-z0-9])`)
					specialChars := re.FindAllString(text, -1)
					if len(specialChars) < opts.minSpecialChars {
						isValid = false
						continue
					}
				}

				if isValid {
					fmt.Println(text)
				}
			}
		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(logsSlice)))))
	}

	wg2.Wait()
	logsSlice = nil
}
