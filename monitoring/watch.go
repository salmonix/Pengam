package monitoring

import (
	"errors"
	l "log"
	"math"
	"time"

	c "GoWorker/config"
	d "GoWorker/datastore"
	e "GoWorker/entropy"

	"github.com/shirou/gopsutil/v3/process"
)

type Metrics struct {
	MemoryPercent float64
	CPUPercent    float64
	KSoftirqdCPU  float64
}

func Monitor() {
	go startWatching()
}

func startWatching() {

	p, err := getPIDByName("GoWorker")
	if err != nil {
		l.Println(err.Error())
	}

	ksoft, err := getPIDByName("ksoftirqd/0")
	if err != nil {
		l.Println(err.Error())
	}

	if err != nil {
		l.Println(err.Error())
		return
	}

	lastWritten := new(Metrics)

	// This is the main monitoring loop for the given process
	send := false
	l.Printf("Monitoring pid %d GoWorker ", p.Pid)

	for {

		memPercent, _ := p.MemoryPercent()
		send = (calcChange(lastWritten.MemoryPercent, float64(memPercent)) > 5.0)

		cpuPercent, _ := p.CPUPercent()
		send = (calcChange(lastWritten.CPUPercent, cpuPercent) > 1.0) || send

		ksoftCPU, _ := ksoft.CPUPercent()
		send = (calcChange(lastWritten.KSoftirqdCPU, ksoftCPU) > 1.0) || send

		//	l.Printf(" %f %f %f ", memPercent, cpuPercent, ksoftCPU)

		if send {
			send = false
			d.WriteToDb("monitoring", "INSERT INTO metrics.monitor(clustername, cpu_percent, memory_percent, ksoftirqd) VALUES($1,$2,$3,$4)",
				[]interface{}{c.Clustername, cpuPercent, memPercent, ksoftCPU})
			lastWritten.CPUPercent = cpuPercent
			lastWritten.MemoryPercent = float64(memPercent)
			lastWritten.KSoftirqdCPU = ksoftCPU
		}

		// handling case of high memory usage - currently ifa cache consumes memory a lot
		if memPercent > 65 {
			e.FreeRandomHighKeys(20, 2048)
			l.Println("Reached mem limit - cleaning up keys")
		}

		time.Sleep(15 * time.Second)
	}

}

func getPIDByName(processName string) (*process.Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		name, _ := p.Name()
		if name == processName {
			return p, nil
		}
	}
	return nil, errors.New("can't find process " + processName)
}

func calcChange(a float64, b float64) float64 {
	return math.Abs(a - b)
}
