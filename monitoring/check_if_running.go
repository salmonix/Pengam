package monitoring

import (
	l "log"
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

func IfRunning() (bool, error) {

	myPid := os.Getpid()

	processes, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			l.Println(err.Error())
			continue
		}
		if (name == "GoWorker") && (int(p.Pid) != myPid) {
			l.Printf("Killing %s with PID %d", name, p.Pid)
			p.Kill()
		}
	}
	return true, nil

}
