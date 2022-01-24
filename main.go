package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func printPercents(percents []int) {
	for _, v := range percents {
		fmt.Printf("%d ", v)
	}
	fmt.Print("\n")
}

func main() {
	fileNum := len(os.Args)

	percents := make([]int, 0, fileNum)
	
	wg := sync.WaitGroup{}
	var index int
	for _, fileURI := range os.Args[1:] {
		fileName := filepath.Base(fileURI)
		if fileName != "" {
			percents = append(percents, 0)
			wg.Add(1)
			go downloadFile(fileName, fileURI, &percents[index], &wg)
			index++
		}
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				printPercents(percents)
			}
		}
	}()
	wg.Wait()
}

const bufSize = 32768

func downloadFile(fileName string, fileURI string, currentPercent *int, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create the file
	out, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating file", fileName, err)
		return
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(fileURI)
	if err != nil {
		fmt.Println("Error getting URI", fileURI, err)
		return
	}
	defer resp.Body.Close()

	fileLength := resp.ContentLength

	// Check server response
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: bad status", resp.Status, "for file", fileName)
		return
	}

	// Write the body to file
	var downloaded int64
	for {
		num, err := io.CopyN(out, resp.Body, bufSize)
		if err != nil && err != io.EOF {
			fmt.Println("Error downloading file", fileName, err)
			return
		}

		downloaded += num
		*currentPercent = int(downloaded * 100 / fileLength)

		if err == io.EOF {
			break
		}
	}
}
