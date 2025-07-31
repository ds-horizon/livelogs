package shell

import (
	"bufio"
	"os/exec"

	"github.com/dream11/livelogs/pkg/logger"
)

var log logger.Logger

// Exec : execute given command
func Exec(command string) int {
	log.Debug("Executing:" + command)

	cmd := exec.Command("bash", "-c", command)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		log.Error("Unable to start cmd execution. " + err.Error())
		return 1
	}

	scannerOut := bufio.NewScanner(stdout)
	for scannerOut.Scan() {
		m := scannerOut.Text()
		log.Debug(m)
	}

	scannerErr := bufio.NewScanner(stderr)
	for scannerErr.Scan() {
		m := scannerErr.Text()
		log.Error(m)
	}

	err = cmd.Wait()
	if err != nil {
		log.Error(err.Error())
		return 1
	}

	return 0
}
