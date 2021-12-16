package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/hashicorp/raft"
)

const (
	CMDSET       = "SET"
	CMDDEL       = "DEL"
	BDGLOGPREFIX = "rft:"
	BDGSSTPREFIX = "sst:"
	BDGDATPREFIX = "dat:"
	BDGU64PREFIX = "u64:"
)

type VpLogCmd struct {
	Op    string
	Key   string
	Value []byte
}

type VpRpcResponse struct {
	Error error
	Data  []byte
}

type VpKeyPage struct {
	Keys     []string
	NextKey  string
	PageSize uint16
}

type Snapshot struct{}

type IteratorRange struct{ from, to uint64 }

var (
	dbLogPrefix    = []byte(BDGLOGPREFIX) // Bucket names we perform transactions in
	dbDatPrefix    = []byte(BDGDATPREFIX)
	dbU64Prefix    = []byte(BDGU64PREFIX)
	dbSstPrefix    = []byte(BDGSSTPREFIX)
	ErrKeyNotFound = errors.New("not found")
)

/*
	BadgerStore provides access to Badger for Raft to store and retrieve log entries.
	It also provides key/value storage, and can be used as a LogStore and StableStore.

	See https://godoc.org/github.com/hashicorp/raft#StableStore
	and https://godoc.org/github.com/hashicorp/raft#LogStore
*/
type BadgerStore struct {
	db *badger.DB
}

func NewBadgerStore(path string) (*BadgerStore, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		log.Error("Badger store error", "cause", err)
	}
	store := &BadgerStore{db: db}
	return store, nil
}

/* ==================================================================================
                            Utility functions
================================================================================== */

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

func dataKeyOf(rawKey []byte) []byte {
	key := fmt.Sprintf("%s%s", dbDatPrefix, hex.EncodeToString(rawKey))
	if log.IsTrace() {
		log.Trace("badger key", "dat", key)
	}
	return []byte(key)
}

func logKeyOf(idxKey uint64) []byte {
	key := fmt.Sprintf("%s%d", dbLogPrefix, idxKey)
	if log.IsTrace() {
		log.Trace("badger key", "log", key)
	}
	return []byte(key)
}

func u64KeyOf(rawKey []byte) []byte {
	key := fmt.Sprintf("%s%s", dbU64Prefix, hex.EncodeToString(rawKey))
	if log.IsDebug() {
		log.Trace("badger key", "u64", key)
	}
	return []byte(key)
}

func sstKeyOf(rawKey []byte) []byte {
	key := fmt.Sprintf("%s%s", dbSstPrefix, hex.EncodeToString(rawKey))
	if log.IsTrace() {
		log.Trace("badger key", "sst", key)
	}
	return []byte(key)
}

func (b *BadgerStore) generateRanges(min, max uint64, batchSize int64) []IteratorRange {
	nSegments := int(math.Round(float64((max - min) / uint64(batchSize))))
	segments := []IteratorRange{}
	if (max - min) <= uint64(batchSize) {
		segments = append(segments, IteratorRange{from: min, to: max})
		return segments
	}
	for len(segments) < nSegments {
		nextMin := min + uint64(batchSize)
		segments = append(segments, IteratorRange{from: min, to: nextMin})
		min = nextMin + 1
	}
	segments = append(segments, IteratorRange{from: min, to: max})
	return segments
}

/* ==================================================================================
                            Log operations
================================================================================== */

// FirstIndex returns the first known index from the Raft log.
func (b *BadgerStore) FirstIndex() (uint64, error) {
	first := uint64(0)
	err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		it.Seek(dbLogPrefix)
		if it.ValidForPrefix(dbLogPrefix) {
			item := it.Item()
			k := string(item.Key()[len(dbLogPrefix):])
			idx, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				return err
			}
			first = idx
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return first, nil
}

// LastIndex returns the last known index from the Raft log.
func (b *BadgerStore) LastIndex() (uint64, error) {
	last := uint64(0)
	if err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		// see https://github.com/dgraph-io/badger/issues/436
		// and https://github.com/dgraph-io/badger/issues/347
		seekKey := append(dbLogPrefix, 0xFF)
		it.Seek(seekKey)
		if it.ValidForPrefix(dbLogPrefix) {
			item := it.Item()
			k := string(item.Key()[len(dbLogPrefix):])
			idx, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				return err
			}
			last = idx
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return last, nil
}

// GetLog is used to retrieve a log from Badger at a given index.
func (b *BadgerStore) GetLog(idx uint64, log *raft.Log) error {
	return b.db.View(func(txn *badger.Txn) error {
		item, _ := txn.Get(logKeyOf(idx))
		if item == nil {
			return raft.ErrLogNotFound
		}
		err := item.Value(func(val []byte) error {
			buf := bytes.NewBuffer(val)
			dec := gob.NewDecoder(buf)
			return dec.Decode(&log)
		})
		return err
	})
}

// StoreLogs is used to store a set of raft logs
func (b *BadgerStore) StoreLogs(logs []*raft.Log) error {
	maxBatchSize := b.db.MaxBatchSize()
	min := uint64(0)
	max := uint64(len(logs))
	ranges := b.generateRanges(min, max, maxBatchSize)
	for _, r := range ranges {
		txn := b.db.NewTransaction(true)
		defer txn.Discard()
		for index := r.from; index < r.to; index++ {
			log := logs[index]
			var out bytes.Buffer
			enc := gob.NewEncoder(&out)
			enc.Encode(log)
			if err := txn.Set(logKeyOf(log.Index), out.Bytes()); err != nil {
				return err
			}
		}
		if err := txn.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// StoreLog is used to store a single raft log
func (b *BadgerStore) StoreLog(log *raft.Log) error {
	return b.StoreLogs([]*raft.Log{log})
}

// DeleteRange is used to delete logs within a given range inclusively.
func (b *BadgerStore) DeleteRange(min, max uint64) error {
	maxBatchSize := b.db.MaxBatchSize()
	ranges := b.generateRanges(min, max, maxBatchSize)
	for _, r := range ranges {
		txn := b.db.NewTransaction(true)
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer txn.Discard()

		it.Rewind()
		minKey := logKeyOf(r.from) // Get the key to start at
		for it.Seek(minKey); it.ValidForPrefix(dbLogPrefix); it.Next() {
			item := it.Item()
			k := string(item.Key()[len(dbLogPrefix):]) // Get the index as a string to convert to uint64
			idx, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				it.Close()
				return err
			}
			if idx > r.to { // Handle out-of-range index
				break
			}
			delKey := logKeyOf(idx) // Delete in-range index
			if err := txn.Delete(delKey); err != nil {
				it.Close()
				return err
			}
		}
		it.Close()
		if err := txn.Commit(); err != nil {
			return err
		}
	}
	return nil
}

/* ==================================================================================
                            Raw access operations
================================================================================== */

func (b *BadgerStore) GetRaw(k []byte) ([]byte, error) {
	txn := b.db.NewTransaction(false)
	defer txn.Discard()
	item, err := txn.Get(k)
	if item == nil {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	v, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	if err := txn.Commit(); err != nil {
		return nil, err
	}
	return append([]byte(nil), v...), nil
}

func (b *BadgerStore) SetRaw(k []byte, v []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
}

func (b *BadgerStore) DeleteRaw(key []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

/* ==================================================================================
                            Data access operations
================================================================================== */

func (b *BadgerStore) KeysOf(prefix []byte, offset []byte, pageSize uint16) (*VpKeyPage, error) {
	keys := make([]string, 0)
	keyPfx := dataKeyOf(prefix)
	keyOff := dataKeyOf(offset)
	var nextKey string
	var i uint16 = 0
	if err := b.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		it.Seek(keyOff)
		for it.ValidForPrefix(keyPfx) {
			rk := strings.Replace(string(it.Item().Key()), BDGDATPREFIX, "", 1)
			k, _ := hex.DecodeString(rk)
			itKey := string(k)
			if i < pageSize {
				keys = append(keys, string(k))
				it.Next()
			} else {
				nextKey = itKey
				break
			}
			i++
		}
		return nil
	}); err != nil {
		return nil, err
	}
	res := VpKeyPage{
		Keys: keys, PageSize: pageSize,
		NextKey: nextKey,
	}
	if log.IsDebug() {
		log.Debug("badger", "response", res)
	}
	return &res, nil
}

func (b *BadgerStore) GetData(key []byte) ([]byte, error) {
	return b.GetRaw(dataKeyOf(key))
}

/* ==================================================================================
                            Additional implementations
================================================================================== */

// Get a value in StableStore.
func (b *BadgerStore) Get(k []byte) ([]byte, error) {
	return b.GetRaw(sstKeyOf(k))
}

// Set a key/value in StableStore.
func (b *BadgerStore) Set(k []byte, v []byte) error {
	return b.SetRaw(sstKeyOf(k), v)
}

// SetUint64 is like Set, but handles uint64 values
func (b *BadgerStore) SetUint64(key []byte, val uint64) error {
	return b.SetRaw(u64KeyOf(key), uint64ToBytes(val))
}

// GetUint64 is like Get, but handles uint64 values
func (b *BadgerStore) GetUint64(key []byte) (uint64, error) {
	val, err := b.GetRaw(u64KeyOf(key))
	if err != nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}

func (b *BadgerStore) Apply(rLog *raft.Log) interface{} {
	switch rLog.Type {
	case raft.LogCommand:
		var payload = VpLogCmd{}
		if err := json.Unmarshal(rLog.Data, &payload); err != nil {
			log.Error("error un-marshaling payload", "cause", err.Error())
			return nil
		}
		switch payload.Op {
		case CMDSET:
			return &VpRpcResponse{Error: b.SetRaw(dataKeyOf([]byte(payload.Key)), payload.Value), Data: payload.Value}
		case CMDDEL:
			return &VpRpcResponse{Error: b.DeleteRaw(dataKeyOf([]byte(payload.Key))), Data: nil}
		default:
			log.Warn("Invalid Raft log command", "payload", payload.Op)
		}
	}
	log.Info("Raft log command", "type", raft.LogCommand)
	return nil
}

func (b *BadgerStore) Close() error {
	return b.db.Close()
}

func (b *BadgerStore) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{}, nil
}

func (b *BadgerStore) Restore(rc io.ReadCloser) error {
	return nil
}

func (s Snapshot) Persist(_ raft.SnapshotSink) error {
	return nil
}

func (s Snapshot) Release() {}
