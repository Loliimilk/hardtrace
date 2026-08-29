package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

type CpuTimesVector struct {
	CPU       string    `json:"cpu"`
	User      []float64 `json:"user"`
	System    []float64 `json:"system"`
	Idle      []float64 `json:"idle"`
	Nice      []float64 `json:"nice"`
	Iowait    []float64 `json:"iowait"`
	Irq       []float64 `json:"irq"`
	Softirq   []float64 `json:"softirq"`
	Steal     []float64 `json:"steal"`
	Guest     []float64 `json:"guest"`
	GuestNice []float64 `json:"guest_nice"`
}

type CPUVector struct {
	CpuTimes []CpuTimesVector `json:"cpu_times"`
}

func getTimesVector(duration time.Duration, interval time.Duration, startTime time.Time) (CPUVector, error) {
	var cpuDataTimes []CpuTimesVector

	stopTimer := time.After(duration)

MainLoop:
	for {

		select {

		case <-stopTimer:

			fmt.Println("collection time is over")

			endTime := time.Now()

			timeDuration := endTime.Sub(startTime)

			fmt.Printf("Finished at: %s\n", endTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("Total runtime: %s\n", timeDuration.Round(time.Minute))
			break MainLoop

		default:

			fmt.Println("starting new operation")

			statTimes, err := cpu.Times(true)
			if err != nil {
				fmt.Println("failed to read CPU times:", err)
				return CPUVector{}, err
			}

			if len(cpuDataTimes) == 0 {
				for _, stat := range statTimes {
					cpuDataTimes = append(cpuDataTimes, CpuTimesVector{
						CPU:       stat.CPU,
						User:      []float64{},
						System:    []float64{},
						Idle:      []float64{},
						Nice:      []float64{},
						Iowait:    []float64{},
						Irq:       []float64{},
						Softirq:   []float64{},
						Steal:     []float64{},
						Guest:     []float64{},
						GuestNice: []float64{},
					})
				}
			}

			for i, stat := range statTimes {

				cpuDataTimes[i].User = append(cpuDataTimes[i].User, stat.User)
				cpuDataTimes[i].System = append(cpuDataTimes[i].System, stat.System)
				cpuDataTimes[i].Idle = append(cpuDataTimes[i].Idle, stat.Idle)
				cpuDataTimes[i].Nice = append(cpuDataTimes[i].Nice, stat.Nice)
				cpuDataTimes[i].Iowait = append(cpuDataTimes[i].Iowait, stat.Iowait)
				cpuDataTimes[i].Irq = append(cpuDataTimes[i].Irq, stat.Irq)
				cpuDataTimes[i].Softirq = append(cpuDataTimes[i].Softirq, stat.Softirq)
				cpuDataTimes[i].Steal = append(cpuDataTimes[i].Steal, stat.Steal)
				cpuDataTimes[i].Guest = append(cpuDataTimes[i].Guest, stat.Guest)
				cpuDataTimes[i].GuestNice = append(cpuDataTimes[i].GuestNice, stat.GuestNice)

			}
			time.Sleep(interval)

		}
	}

	return CPUVector{
		CpuTimes: cpuDataTimes,
	}, nil

}

func metricJSON(cpuVector CPUVector) {

	jsData, err := json.MarshalIndent(cpuVector, "", " ")
	if err != nil {
		fmt.Println("failed to create JSON:", err)
		return
	}

	err = os.MkdirAll("metrics", 0755)
	if err != nil {
		fmt.Println("failed to create metrics directory:", err)
		return
	}

	err = os.WriteFile("metrics/metric.json", jsData, 0644)
	if err != nil {
		fmt.Println("failed to write metric.json:", err)
		return

	}
}

func main() {

	startTime := time.Now()
	fmt.Printf("Started at: %s\n", startTime.Format("2006-01-02 15:04:05"))

	duration := 1 * time.Minute
	interval := 2 * time.Second

	cpuVector, err := getTimesVector(duration, interval, startTime)
	if err != nil {
		fmt.Println("failed to collect CPU times:", err)
		return
	}

	metricJSON(cpuVector)

}
