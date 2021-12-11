package web

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"sync"
)

type Database struct {
	mutex sync.Mutex
	values map[string]string
}

type DatabaseSnapshot struct {
	values map[string]string
}

type Command struct {
	Op string `json:"op,omitempty"`
	Key string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

func (d *Database) set(key, value string) interface{} {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.values[key] = value
	return nil
}

func (d *Database) get(key string) interface{} {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value := d.values[key]
	return value
}

// Implements FSM for Database
func (d *Database) Apply(l *raft.Log) interface{} {
	var c Command
	if err := json.Unmarshal(l.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshall cmd: %v", err))
	}

	switch c.Op {
	case "set":
		return d.set(c.Key, c.Value)
	case "get":
		return d.get(c.Key)
	default:
		panic(fmt.Sprintf("command unknown: %s", c.Op))
	}
}

// Implements FSM for Database
func (d *Database) Snapshot() (raft.FSMSnapshot, error)  {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	o := make(map[string]string)
	for k, v := range d.values {
		o[k] = v
	}
	return &DatabaseSnapshot{values: o}, nil
}

// Implements FSM for Database
func (d *Database) Restore(rc io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(rc).Decode(&o); err != nil {
		return err
	}

	d.values = o
	return nil
}

func NewDatabase() *Database {
	return &Database{
		mutex:  sync.Mutex{},
		values: make(map[string]string),
	}
}

// Implements FSMSnapshot for DatabaseSnapshot
func (ds *DatabaseSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		buff, err := json.Marshal(ds.values)
		if err != nil {
			return err
		}

		if _, err := sink.Write(buff); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
	}

	return err
}

// Implements FSMSnapshot for DatabaseSnapshot
func (ds *DatabaseSnapshot) Release() {
}
