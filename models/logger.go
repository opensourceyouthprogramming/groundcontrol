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
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"groundcontrol/pubsub"
	"groundcontrol/relay"
)

var logLevelPriorities = map[LogLevel]int{
	LogLevelDebug:   0,
	LogLevelInfo:    1,
	LogLevelWarning: 2,
	LogLevelError:   3,
}

// Logger logs messages.
type Logger struct {
	nodes *NodeManager
	subs  *pubsub.PubSub

	cap      int
	level    LogLevel
	systemID string

	lastID      uint64
	logEntryIDs []string
	head        int

	debugCounter   int64
	infoCounter    int64
	warningCounter int64
	errorCounter   int64

	stdoutLog *log.Logger
	stderrLog *log.Logger
}

// NewLogger creates a Logger with given capacity and level.
func NewLogger(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	cap int,
	level LogLevel,
	systemID string,
) *Logger {
	return &Logger{
		nodes:       nodes,
		subs:        subs,
		cap:         cap,
		level:       level,
		systemID:    systemID,
		logEntryIDs: make([]string, cap*2),
		stdoutLog:   log.New(os.Stdout, "", log.LstdFlags),
		stderrLog:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

// Add adds a log entry.
func (l *Logger) Add(
	level LogLevel,
	ownerID string,
	message string,
) (string, error) {
	if logLevelPriorities[level] < logLevelPriorities[l.level] {
		return "", nil
	}

	log := l.stdoutLog
	if logLevelPriorities[level] >= logLevelPriorities[LogLevelWarning] {
		log = l.stderrLog
	}

	if ownerID != "" {
		log.Printf("%s\t<%s> %s", level, ownerID, message)
	} else {
		log.Printf("%s\t%s", level, message)
	}

	id := atomic.AddUint64(&l.lastID, 1)
	now := DateTime(time.Now())
	logEntry := LogEntry{
		ID:        relay.EncodeID(NodeTypeLogEntry, fmt.Sprint(id)),
		Level:     level,
		CreatedAt: now,
		Message:   message,
		OwnerID:   ownerID,
	}
	l.nodes.MustStoreLogEntry(logEntry)
	l.subs.Publish(LogEntryAdded, logEntry.ID)

	l.nodes.MustLockSystem(l.systemID, func(system System) {
		if l.head >= l.cap*2 {
			copy(l.logEntryIDs, l.logEntryIDs[l.cap:])
			l.head = l.cap

			for _, oldEntryID := range l.logEntryIDs[l.head:] {
				oldEntry := l.nodes.MustLoadLogEntry(oldEntryID)

				switch oldEntry.Level {
				case LogLevelDebug:
					atomic.AddInt64(&l.debugCounter, -1)
				case LogLevelInfo:
					atomic.AddInt64(&l.infoCounter, -1)
				case LogLevelWarning:
					atomic.AddInt64(&l.warningCounter, -1)
				case LogLevelError:
					atomic.AddInt64(&l.errorCounter, -1)
				}

				l.nodes.MustDeleteLogEntry(oldEntryID)
			}
		}

		l.logEntryIDs[l.head] = logEntry.ID

		l.head++

		system.LogEntryIDs = l.logEntryIDs[:l.head]
		l.nodes.MustStoreSystem(system)
	})

	switch level {
	case LogLevelDebug:
		atomic.AddInt64(&l.debugCounter, 1)
	case LogLevelInfo:
		atomic.AddInt64(&l.infoCounter, 1)
	case LogLevelWarning:
		atomic.AddInt64(&l.warningCounter, 1)
	case LogLevelError:
		atomic.AddInt64(&l.errorCounter, 1)
	}

	l.publishMetrics()

	return logEntry.ID, nil
}

// Debug adds a debug entry.
func (l *Logger) Debug(message string, a ...interface{}) string {
	id, err := l.Add(LogLevelDebug, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Info adds an info entry.
func (l *Logger) Info(message string, a ...interface{}) string {
	id, err := l.Add(LogLevelInfo, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Warning adds a warning entry.
func (l *Logger) Warning(message string, a ...interface{}) string {
	id, err := l.Add(LogLevelWarning, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// Error adds an error entry.
func (l *Logger) Error(message string, a ...interface{}) string {
	id, err := l.Add(LogLevelError, "", fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// DebugWithOwner adds a debug entry with an owner.
func (l *Logger) DebugWithOwner(ownerID string, message string, a ...interface{}) string {
	id, err := l.Add(LogLevelDebug, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// InfoWithOwner adds an info entry with an owner.
func (l *Logger) InfoWithOwner(ownerID string, message string, a ...interface{}) string {
	id, err := l.Add(LogLevelInfo, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// WarningWithOwner adds a warning entry with an owner.
func (l *Logger) WarningWithOwner(ownerID string, message string, a ...interface{}) string {
	id, err := l.Add(LogLevelWarning, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

// ErrorWithOwner adds an error entry with an owner.
func (l *Logger) ErrorWithOwner(ownerID string, message string, a ...interface{}) string {
	id, err := l.Add(LogLevelError, ownerID, fmt.Sprintf(message, a...))
	if err != nil {
		panic(err)
	}
	return id
}

func (l *Logger) publishMetrics() {
	system := l.nodes.MustLoadSystem(l.systemID)

	l.nodes.MustLockLogMetrics(system.LogMetricsID, func(metrics LogMetrics) {
		metrics.Debug = int(atomic.LoadInt64(&l.debugCounter))
		metrics.Info = int(atomic.LoadInt64(&l.infoCounter))
		metrics.Warning = int(atomic.LoadInt64(&l.warningCounter))
		metrics.Error = int(atomic.LoadInt64(&l.errorCounter))
		l.nodes.MustStoreLogMetrics(metrics)
	})

	l.subs.Publish(LogMetricsUpdated, system.LogMetricsID)
}
