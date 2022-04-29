package log

import (
	api "github.com/zdarovich/cowboy_shooters/api/log"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

// Log struct where dir refers to the directory where the log is stored
type Log struct {
	Dir   string
	store *store
}

// Row struct to marshal logs as json
type Row struct {
	Time time.Time
	Msg  string
}

// NewLog creating and setting up the log instance
func NewLog(dir string) (*Log, error) {
	log := &Log{
		Dir: dir,
	}
	return log, log.setup()
}

func (l *Log) setup() error {
	storeFile, err := os.OpenFile(
		path.Join(l.Dir, fmt.Sprintf("%s%s", "log", ".txt")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return err
	}
	if l.store, err = newStore(storeFile); err != nil {
		return err
	}
	return nil
}

func (l *Log) Append(record *api.Record) (uint64, error) {
	row := Row{time.Now(), string(record.Value)}
	bytes, err := json.Marshal(row)
	if err != nil {
		return 0, err
	}
	off, _, err := l.store.Append(bytes)
	if err != nil {
		return 0, err
	}

	return off, err
}

func (l *Log) Close() error {
	if err := l.store.Close(); err != nil {
		return err
	}
	return nil
}

// Remove remove closes the log, and removes the data
func (l *Log) Remove() error {
	if err := l.Close(); err != nil {
		return err
	}
	return os.RemoveAll(l.Dir)
}

// Reset removes the log, and creates a new log
func (l *Log) Reset() error {
	if err := l.Remove(); err != nil {
		return err
	}

	return l.setup()
}
