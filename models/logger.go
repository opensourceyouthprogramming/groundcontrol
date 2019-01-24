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
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
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

	nextID uint64

	debugCounter   int64
	infoCounter    int64
	warningCounter int64
	errorCounter   int64
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
		nodes:    nodes,
		subs:     subs,
		cap:      cap,
		systemID: systemID,
	}
}

// Cap returns the capacity of the logger.
func (l *Logger) Cap() int {
	return l.cap
}

// Add adds a log entry.
func (l *Logger) Add(
	level LogLevel,
	message string,
	meta interface{},
) (string, error) {
	if logLevelPriorities[level] < logLevelPriorities[l.level] {
		return "", nil
	}

	metaJSON := ""
	if meta != nil {
		b, err := json.Marshal(meta)
		if err != nil {
			return "", err
		}
		metaJSON = string(b)
	}
	log.Printf("%s\t%s %s", level, message, string(metaJSON))

	id := atomic.AddUint64(&l.nextID, 1)
	now := date.NowFormatted()
	logEntry := LogEntry{
		ID:        relay.EncodeID(NodeTypeLogEntry, fmt.Sprint(id)),
		Level:     level,
		CreatedAt: now,
		Message:   message,
		MetaJSON:  string(metaJSON),
	}
	l.nodes.MustStoreLogEntry(logEntry)
	l.subs.Publish(LogEntryAdded, logEntry.ID)

	l.nodes.MustLockSystem(l.systemID, func(system System) {
		if len(system.LogEntryIDs) >= l.cap*2 {
			// TODO: remove previous log entries
			logEntryIDs := make([]string, l.cap, l.cap*2)
			copy(logEntryIDs, system.LogEntryIDs[:l.cap])
			system.LogEntryIDs = logEntryIDs
		}

		system.LogEntryIDs = append([]string{logEntry.ID}, system.LogEntryIDs...)
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
func (l *Logger) Debug(message string, meta interface{}) string {
	id, err := l.Add(LogLevelDebug, message, meta)
	if err != nil {
		panic(err)
	}
	return id
}

// Info adds an info entry.
func (l *Logger) Info(message string, meta interface{}) string {
	id, err := l.Add(LogLevelInfo, message, meta)
	if err != nil {
		panic(err)
	}
	return id
}

// Warning adds a warning entry.
func (l *Logger) Warning(message string, meta interface{}) string {
	id, err := l.Add(LogLevelWarning, message, meta)
	if err != nil {
		panic(err)
	}
	return id
}

// Error adds an error entry.
func (l *Logger) Error(message string, meta interface{}) string {
	id, err := l.Add(LogLevelError, message, meta)
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