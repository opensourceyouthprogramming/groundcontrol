// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

// ProcessManager manages creating and running jobs.
type ProcessManager struct {
	nodes *NodeManager
	log   *Logger
	subs  *pubsub.PubSub

	getProjectPath ProjectPathGetter

	systemID string

	nextID   uint64
	commands sync.Map

	runningCounter int64
	doneCounter    int64
	failedCounter  int64
}

// NewProcessManager creates a ProcessManager.
// TODO: Add Clean() method to terminate all processes on shutdown.
func NewProcessManager(
	nodes *NodeManager,
	log *Logger,
	subs *pubsub.PubSub,
	getProjectPath ProjectPathGetter,
	systemID string,
) *ProcessManager {
	return &ProcessManager{
		nodes:          nodes,
		log:            log,
		subs:           subs,
		getProjectPath: getProjectPath,
		systemID:       systemID,
	}
}

// CreateGroup creates a new ProcessGroup and returns its ID.
func (p *ProcessManager) CreateGroup(taskID string) string {
	id := relay.EncodeID(
		NodeTypeProcessGroup,
		fmt.Sprint(atomic.AddUint64(&p.nextID, 1)),
	)

	group := ProcessGroup{
		ID:        id,
		CreatedAt: date.NowFormatted(),
		TaskID:    taskID,
	}

	p.nodes.MustStoreProcessGroup(group)
	p.subs.Publish(ProcessGroupUpserted, id)

	p.nodes.MustLockSystem(p.systemID, func(system System) {
		system.ProcessGroupIDs = append(
			[]string{id},
			system.ProcessGroupIDs...,
		)

		p.nodes.MustStoreSystem(system)
	})

	return id
}

// Run launches a new Process and adds it to a ProcessGroup.
func (p *ProcessManager) Run(
	command string,
	processGroupID string,
	projectID string,
) string {
	id := relay.EncodeID(
		NodeTypeProcess,
		fmt.Sprint(atomic.AddUint64(&p.nextID, 1)),
	)

	process := Process{
		ID:             id,
		Command:        command,
		ProcessGroupID: processGroupID,
		ProjectID:      projectID,
	}

	p.nodes.MustStoreProcess(process)
	p.nodes.MustLockProcessGroup(processGroupID, func(processGroup ProcessGroup) {
		processGroup.ProcessIDs = append([]string{id}, processGroup.ProcessIDs...)
		p.nodes.MustStoreProcessGroup(processGroup)
	})

	p.exec(id)

	return id
}

// Start starts a process that was stopped.
func (p *ProcessManager) Start(processID string) error {
	var processError error

	err := p.nodes.LockProcess(processID, func(process Process) {
		switch process.Status {
		case ProcessStatusRunning, ProcessStatusStopping:
			processError = ErrNotStopped
			return
		case ProcessStatusDone:
			atomic.AddInt64(&p.doneCounter, -1)
		case ProcessStatusFailed:
			atomic.AddInt64(&p.failedCounter, -1)
		}

	})
	if err != nil {
		return err
	}
	if processError != nil {
		return processError
	}

	p.exec(processID) // will publish metrics

	return nil
}

// Stop stops a running process.
func (p *ProcessManager) Stop(processID string) error {
	var processError error

	err := p.nodes.LockProcess(processID, func(process Process) {
		if process.Status != ProcessStatusRunning {
			processError = ErrNotRunning
			return
		}

		process.Status = ProcessStatusStopping
		p.nodes.MustStoreProcess(process)

		p.subs.Publish(ProcessUpserted, processID)
		p.subs.Publish(ProcessGroupUpserted, process.ProcessGroupID)

		actual, ok := p.commands.Load(processID)
		if !ok {
			panic("command not found")
		}
		cmd := actual.(*exec.Cmd)

		pgid, processError := syscall.Getpgid(cmd.Process.Pid)
		if processError != nil {
			return
		}

		processError = syscall.Kill(-pgid, syscall.SIGINT)
		if processError != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	return processError
}

// Clean terminates all running processes.
func (p *ProcessManager) Clean(ctx context.Context) {
	waitGroup := sync.WaitGroup{}

	p.commands.Range(func(k, _ interface{}) bool {
		processID := k.(string)

		p.log.InfoWithOwner(processID, "stopping process")

		if err := p.Stop(processID); err != nil {
			p.log.ErrorWithOwner(processID, "failed to stop process because %s", err.Error())
			return true
		}

		waitGroup.Add(1)

		processCtx, cancel := context.WithCancel(ctx)

		go func() {
			<-processCtx.Done()
			waitGroup.Done()
		}()

		p.subs.Subscribe(processCtx, ProcessUpserted, func(msg interface{}) {
			id := msg.(string)
			if id != processID {
				return
			}

			process := p.nodes.MustLoadProcess(id)

			switch process.Status {
			case ProcessStatusDone, ProcessStatusFailed:
				p.log.InfoWithOwner(processID, "process stopped")
				cancel()
			}
		})

		return true
	})

	waitGroup.Wait()
}

func (p *ProcessManager) exec(id string) {
	p.nodes.MustLockProcess(id, func(process Process) {
		project := process.Project(p.nodes)
		workspace := project.Workspace(p.nodes)

		dir := p.getProjectPath(
			workspace.Slug,
			project.Repository,
			project.Branch,
		)

		stdout := CreateLineWriter(p.log.InfoWithOwner, project.ID)
		stderr := CreateLineWriter(p.log.WarningWithOwner, project.ID)
		cmd := exec.Command("bash", "-l", "-c", process.Command)
		cmd.Dir = dir
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		err := cmd.Start()
		if err == nil {
			process.Status = ProcessStatusRunning
			atomic.AddInt64(&p.runningCounter, 1)
		} else {
			process.Status = ProcessStatusFailed
			atomic.AddInt64(&p.failedCounter, 1)
		}

		p.nodes.MustStoreProcess(process)
		p.subs.Publish(ProcessUpserted, id)
		p.subs.Publish(ProcessGroupUpserted, process.ProcessGroupID)
		p.publishMetrics()

		if err != nil {
			p.log.ErrorWithOwner(project.ID, "process failed because %s", err.Error())
			stdout.Close()
			stderr.Close()
			return
		}

		p.log.InfoWithOwner(project.ID, "process is running")
		p.commands.Store(id, cmd)

		go func() {
			err := cmd.Wait()

			p.nodes.MustLockProcess(id, func(process Process) {
				p.commands.Delete(id)

				if err == nil {
					process.Status = ProcessStatusDone
					atomic.AddInt64(&p.doneCounter, 1)
					p.log.InfoWithOwner(project.ID, "process done")
				} else {
					process.Status = ProcessStatusFailed
					atomic.AddInt64(&p.failedCounter, 1)
					p.log.ErrorWithOwner(project.ID, "process failed because %s", err.Error())
				}

				atomic.AddInt64(&p.runningCounter, -1)
				p.nodes.MustStoreProcess(process)
			})

			p.subs.Publish(ProcessUpserted, id)
			p.subs.Publish(ProcessGroupUpserted, process.ProcessGroupID)
			p.publishMetrics()

			stdout.Close()
			stderr.Close()
		}()
	})
}

func (p *ProcessManager) publishMetrics() {
	system := p.nodes.MustLoadSystem(p.systemID)

	p.nodes.MustLockProcessMetrics(system.ProcessMetricsID, func(metrics ProcessMetrics) {
		metrics.Running = int(atomic.LoadInt64(&p.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&p.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&p.failedCounter))
		p.nodes.MustStoreProcessMetrics(metrics)
	})

	p.subs.Publish(ProcessMetricsUpdated, system.ProcessMetricsID)
}

// CreateLineWriter creates a writer with a line splitter.
// Remember to call close().
func CreateLineWriter(
	write func(ownerID, message string, a ...interface{}) string,
	ownerID string,
	a ...interface{},
) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)

	go func() {
		for scanner.Scan() {
			write(ownerID, scanner.Text(), a...)

			// Don't kill the poor browser.
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return w
}
