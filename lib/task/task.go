package task

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
)

const (
	Init = iota
	Running
	Succeeded
	Failed
	Skipped
	Aborted
)

type key int

const BuildID key = 0

type Task struct {
	Name           string
	Command        string
	Directory      string
	Env            map[string]string
	Parallel       []*Task
	Serial         []*Task
	Stdout         *bytes.Buffer
	Stderr         *bytes.Buffer
	CombinedOutput *bytes.Buffer
	Status         int
	Cmd            *exec.Cmd
	Include        string
	OnlyIf         string   `yaml:"only_if"`
	WaitFor        *WaitFor `yaml:"wait_for"`
}

type outputHandler struct {
	task   *Task
	writer io.Writer
	copy   io.Writer
	mu     *sync.Mutex
}

func (t *Task) Run(ctx context.Context, cancel context.CancelFunc, prevTask *Task) error {
	if t.Command == "" {
		return nil
	}

	if t.Directory != "" {
		re := regexp.MustCompile(`\$[A-Z1-9\-_]+`)
		matches := re.FindAllString(t.Directory, -1)
		for _, m := range matches {
			env := os.Getenv(strings.TrimPrefix(m, "$"))
			t.Directory = strings.Replace(t.Directory, m, env, -1)
		}
	}

	buildID := ctx.Value(BuildID)

	if t.OnlyIf != "" {
		cmd := exec.Command("sh", "-c", t.OnlyIf)
		cmd.Dir = t.Directory

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("BUILD_ID=%s", buildID))
		for k, v := range t.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}

		err := cmd.Run()

		if err != nil {
			log.Warnf("[%s] Skipped because only_if failed: %s", t.Name, err)
			return nil
		}
	}

	if t.WaitFor != nil {
		err := t.wait()
		if err != nil {
			return err
		}
	}

	log.Infof("[%s] Start task", t.Name)
	log.Infof("[%s] Command: %s ", t.Name, t.Command)

	t.Cmd = exec.Command("sh", "-c", t.Command)
	t.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	t.Cmd.Dir = t.Directory

	t.Cmd.Env = os.Environ()
	t.Cmd.Env = append(t.Cmd.Env, fmt.Sprintf("BUILD_ID=%s", buildID))
	for k, v := range t.Env {
		t.Cmd.Env = append(t.Cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if prevTask != nil && prevTask.Stdout != nil {
		t.Cmd.Stdin = bytes.NewBuffer(prevTask.Stdout.Bytes())
	}

	t.Stdout = new(bytes.Buffer)
	t.Stderr = new(bytes.Buffer)
	t.CombinedOutput = new(bytes.Buffer)

	var mu sync.Mutex
	t.Cmd.Stdout = &outputHandler{t, t.Stdout, t.CombinedOutput, &mu}
	t.Cmd.Stderr = &outputHandler{t, t.Stderr, t.CombinedOutput, &mu}

	if err := t.Cmd.Start(); err != nil {
		t.Status = Failed
		return err
	}

	t.Status = Running

	go func(t *Task) {
		for {
			select {
			case <-ctx.Done():
				if t.Status == Running {
					t.Status = Aborted
					t.Cmd.Process.Kill()
					pgid, err := syscall.Getpgid(t.Cmd.Process.Pid)
					if err == nil {
						syscall.Kill(-pgid, syscall.SIGTERM)
					}
					log.Warnf("[%s] aborted", t.Name)
				}
				return
			}
		}
	}(t)

	t.Cmd.Wait()

	if t.Cmd.ProcessState.Success() {
		t.Status = Succeeded
	} else if t.Status == Running {
		t.Status = Failed
		cancel()
		return errors.New("Task failed")
	}

	if t.Status == Succeeded {
		log.Infof("[%s] End task", t.Name)
	}

	return nil
}

func (o *outputHandler) Write(b []byte) (int, error) {
	log.Infof("[%s] %s", o.task.Name, strings.TrimSuffix(string(b), "\n"))

	o.mu.Lock()
	defer o.mu.Unlock()
	o.writer.Write(b)
	o.copy.Write(b)

	return len(b), nil
}
