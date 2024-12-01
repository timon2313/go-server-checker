package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ServerStats struct {
	LoadAverage  int
	TotalMemory  int
	UsedMemory   int
	TotalDisk    int
	UsedDisk     int
	TotalNetwork int
	UsedNetwork  int
}

var failedAttempts int

// >= 3 error requests
func handleFailedAttempt() {
	failedAttempts++
	if failedAttempts >= 3 {
		failedAttempts = 0
		fmt.Println("Unable to fetch server statistics after multiple attempts.")
	}
}

// parse server statistics
func fetchServerStats(url string) (*ServerStats, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch server stats: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	rawData := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(rawData) != 7 {
		return nil, fmt.Errorf("invalid data length: expected 7, got %d", len(rawData))
	}

	data := make([]int, len(rawData))
	for i, value := range rawData {
		data[i], err = strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("invalid data format: %w", err)
		}
	}

	return &ServerStats{
		LoadAverage:  data[0],
		TotalMemory:  data[1],
		UsedMemory:   data[2],
		TotalDisk:    data[3],
		UsedDisk:     data[4],
		TotalNetwork: data[5],
		UsedNetwork:  data[6],
	}, nil
}

// analyze server statistics
func checkServerStats(stats *ServerStats) {
	if stats.LoadAverage > 30 {
		fmt.Printf("Load Average is too high: %d\n", stats.LoadAverage)
	}

	if stats.TotalMemory > 0 {
		memoryUsage := stats.UsedMemory * 100 / stats.TotalMemory
		if memoryUsage > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memoryUsage)
		}
	}

	if stats.TotalDisk > 0 {
		freeDiskSpace := stats.TotalDisk - stats.UsedDisk
		diskUsage := stats.UsedDisk * 100 / stats.TotalDisk
		if diskUsage > 90 {
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskSpace/(1024*1024))
		}
	}

	if stats.TotalNetwork > 0 {
		freeBandwidth := stats.TotalNetwork - stats.UsedNetwork
		networkUsage := stats.UsedNetwork * 100 / stats.TotalNetwork
		if networkUsage > 90 {
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeBandwidth/1000000)
		}
	}
}

func main() {
	const statsURL = "http://srv.msk01.gigacorp.local/_stats"

	for {
		stats, err := fetchServerStats(statsURL)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			handleFailedAttempt()
			time.Sleep(time.Second)
			continue
		}

		checkServerStats(stats)
		failedAttempts = 0

		time.Sleep(time.Second)
	}
}
