package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

type TimeEntry struct {
	Timestamp string          `json:"timestamp"`
	CpuTimes  []cpu.TimesStat `json:"cpu_times"`
}

func getCpuInfo() (cpu.InfoStat, error) {

	statInfo, err := cpu.Info()
	if err != nil {
		return cpu.InfoStat{}, err
	}

	if len(statInfo) == 0 {
		return cpu.InfoStat{}, fmt.Errorf("no CPU information found")
	}

	return statInfo[0], nil

}

func streamCpuTimes(ctx context.Context, interval time.Duration, ch chan<- TimeEntry) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	defer close(ch)

	for {
		select {

		case <-ctx.Done():
			fmt.Println("[Goroutine]: Received stop signal. Stopping metric collection...")
			return

		case <-ticker.C:
			fmt.Println("\n[Goroutine]: Taking a new CPU measurement...")
			statTimes, err := cpu.Times(true)
			if err != nil {
				fmt.Println("failed to read CPU times:", err)
				continue
			}

			entry := TimeEntry{
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
				CpuTimes:  statTimes,
			}

			select {
			case ch <- entry:
			case <-ctx.Done():
				return
			}
		}

	}
}

func metricJSON(file *os.File, entry TimeEntry) {

	encoder := json.NewEncoder(file)

	if err := encoder.Encode(entry); err != nil {
		fmt.Println("failed to encode and write JSON:", err)
	}
}

func main() {

	cpuInfo, err := getCpuInfo()
	if err != nil {
		fmt.Println("failed to get CPU info:", err)
		return
	}

	fmt.Printf("Started collection for: %s\n", cpuInfo.ModelName)

	if err := os.MkdirAll("metrics", 0755); err != nil {
		fmt.Println("failed to create metrics directory:", err)
		return
	}

	file, err := os.OpenFile("metrics/metric.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("failde to open file", err)
		return
	}
	defer file.Close()

	timesChan := make(chan TimeEntry)

	//ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interval := 1 * time.Second

	go streamCpuTimes(ctx, interval, timesChan)

	fmt.Println("Press [ENTER] at any time to stop the program...")

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			fmt.Println("\n[Main]: User input detected! Signaling cancellation...")
			cancel()
		}
	}()

	for currentEntry := range timesChan {
		metricJSON(file, currentEntry)
		fmt.Printf("[Record]: Added vector for %s\n", currentEntry.Timestamp)
	}

	fmt.Println("Program completed successfully.")
}
