package etcdaero

import (
	"log"
	"sync"
	"time"
)

const (
	// 10,05 mins
	aerospikeTTL = 10050 * time.Millisecond
)

// IEntryData data buffer for export and import
type IAeroBody interface {
	Get(data []interface{}) (interface{}, bool, error)
	ReNew(data []byte) error
}

type AeroChecker struct {
	sync.RWMutex
	Conn     *AeroSpikeClient
	List     map[string]IAeroBody
	SignalCh chan bool
	StopCh   chan bool
	SetTTLCh chan time.Duration
}

var singletonAero *AeroChecker

// FOR TEST ONLY?
func _cleanSingletonAero() {
	if singletonAero == nil {
		return
	}

	singletonAero.StopCh <- true
	singletonAero.Conn.Close()
	singletonAero = nil
}

func InitAeroChecker(conn *AeroSpikeClient) *AeroChecker {

	if singletonAero != nil {
		return singletonAero
	}

	singletonAero = _initAeroChecker(conn)

	return singletonAero
}

func _initAeroChecker(conn *AeroSpikeClient) *AeroChecker {

	if singletonAero != nil {
		return singletonAero
	}

	singletonAero = &AeroChecker{
		Conn:     conn,
		List:     map[string]IAeroBody{},
		SignalCh: make(chan bool, 100),
		StopCh:   make(chan bool, 2),
		SetTTLCh: make(chan time.Duration, 2),
	}

	go singletonAero._start()

	return singletonAero
}

func (aero *AeroChecker) _start() {
	ttl := aerospikeTTL
	for {
		select {
		case <-time.Tick(ttl):
			aero._load()
		case <-aero.SignalCh:
			aero._load()
		case <-aero.StopCh:
			return
		case ttl = <-aero.SetTTLCh:
			// nothing
		}
	}
}

func (aero *AeroChecker) _load() {

	list := aero._keys()

	for _, key := range list {

		cacheKey := &AeroSpikeKey{
			Set: key,
			Pk:  key,
		}
		data := &EtcdAeroEntry{}

		if ok := aero.Conn.LoadEntry(cacheKey, data); !ok {
			// Load from cache has mistake
			log.Printf("Not found in cache. Key: %s.%s", cacheKey.Set, cacheKey.Pk)
			continue
		}

		aero.reNew(key, data.Body)
	}

}

func (aero *AeroChecker) _keys() []string {

	out := []string{}

	aero.RLock()
	defer aero.RUnlock()

	for key := range aero.List {
		out = append(out, key)
	}

	return out
}

func (aero *AeroChecker) reNew(key string, data []byte) {
	aero.RLock()
	defer aero.RUnlock()

	obj, ok := aero.List[key]
	if !ok {
		return
	}

	obj.ReNew(data)
}

func SetTTLAero(ttl time.Duration) {
	singletonAero.SetTTL(ttl)
}

func (aero *AeroChecker) SetTTL(ttl time.Duration) {
	aero.SetTTLCh <- ttl
}

func ReLoadAero() {
	singletonAero.ReLoad()
}

func (aero *AeroChecker) ReLoad() {
	aero.SignalCh <- true
}

func StartAeroReader(key string, obj ...IAeroBody) {
	if len(obj) == 0 {
		singletonAero._add(key, NewlocalAeroStorage())
	} else {
		singletonAero._add(key, obj[0])
	}
}

func (aero *AeroChecker) _add(key string, obj IAeroBody) {
	aero.Lock()
	aero.List[key] = obj
	aero.Unlock()
}

func PutAero(key *AeroSpikeKey, data IEntryData, ttl time.Duration) {
	singletonAero.Put(key, data, ttl)
}

func (aero *AeroChecker) Put(key *AeroSpikeKey, data IEntryData, ttl time.Duration) {
	aero.Conn.PutEntry(key, data, ttl)
}

func GetAero(key string, params ...interface{}) (interface{}, bool) {
	return singletonAero.Get(key, params...)
}

func (aero *AeroChecker) Get(key string, params ...interface{}) (interface{}, bool) {
	aero.RLock()
	defer aero.RUnlock()

	obj, ok := aero.List[key]

	if !ok {
		return nil, false
	}

	out, ok, err := obj.Get(params)
	if err != nil {
		log.Printf("got for key %s error: %s", key, err)
		return out, false
	}

	return out, ok
}
