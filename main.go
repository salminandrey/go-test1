package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL       = "http://srv.msk01.gigacorp.local/_stats"
	maxErrors       = 3
	loadAvgLimit    = 30.0
	memoryThreshold = 0.8
	diskThreshold   = 0.9
	networkThreshold = 0.9
)


func fetchStats() (string, error) {
	resp, err := http.Get(serverURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("invalid response status: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseStats(data string) ([]float64, error) {
	parts := strings.Split(data, ",")
	if len(parts) != 7 {
		return nil, errors.New("invalid data format")
	}

	values := make([]float64, len(parts))
	for i, part := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value %q: %v", part, err)
		}
		values[i] = val
	}
	return values, nil
}

func checkThresholds(stats []float64) {
	loadAvg := stats[0]
	memTotal := stats[1]
	memUsed := stats[2]
	diskTotal := stats[3]
	diskUsed := stats[4]
	netTotal := stats[5]
	netUsed := stats[6]

	if loadAvg > loadAvgLimit {
		fmt.Printf("Load Average is too high: %f.0", loadAvg)
	}

	memUsage := memUsed / memTotal
	if memUsage > memoryThreshold {
		fmt.Printf("Memory usage too high: %f.0%%\n", memUsage*100)
	}

	freeDiskMB := (diskTotal - diskUsed) / (1024 * 1024)
	if diskUsed/diskTotal > diskThreshold {
		fmt.Printf("Free disk space is too low: %f.0 Mb left\n", freeDiskMB)
	}

	freeNetMbit := ((netTotal - netUsed) * 8) / (1024 * 1024)
	if netUsed/netTotal > networkThreshold {
		fmt.Printf("Network bandwidth usage high: %f.0 Mbit/s available\n", freeNetMbit)
	}
}

func main() {
	errorCount := 0
	for {
		data, err := fetchStats()
		if err != nil {
			errorCount++
			fmt.Println("Error fetching data:", err)
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(5 * time.Second)
			continue
		}

		stats, err := parseStats(data)
		if err != nil {
			errorCount++
			fmt.Println("Error parsing data:", err)
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(5 * time.Second)
			continue
		}

		errorCount = 0
		checkThresholds(stats)
		time.Sleep(10 * time.Second)
	}
}