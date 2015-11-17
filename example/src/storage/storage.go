package storage

import (
	"encoding/json"
	"errors"
	"etcdaero"
	"sync"

	"fmt"
)

type localStorage struct {
	sync.RWMutex
	etcdaero.IAeroBody
	Data map[string]interface{}
}

// We want to single data for each goroutine.
var Singleton *localStorage

func NewlocalStorage() localStorage {
	if Singleton == nil {
		Singleton = &localStorage{
			Data: map[string]interface{}{},
		}
	}
	return *Singleton
}

func (puk localStorage) ReNew(data []byte) error {
	return Singleton._ReNew(data)
}

// We can prepare out data for local storaging
func (puk *localStorage) _ReNew(data []byte) error {
	d := map[string]interface{}{}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	fmt.Printf("\n----\ndata from aerospike\n")
	fmt.Printf("Now we prepare that to using: %+v\n", d)

	// Preparing data for current node's requirements.
	s := map[string]interface{}{
		"ru": d["1"],
		"vi": d["2"],
		"en": d["3"],
	}

	fmt.Printf("Prepared data: %+v\n", s)

	puk.Lock()
	puk.Data = s
	puk.Unlock()

	return nil
}

func (puk localStorage) Get(data []interface{}) (interface{}, bool, error) {
	return Singleton._Get(data)
}

func (puk *localStorage) _Get(data []interface{}) (interface{}, bool, error) {

	if len(data) == 0 {
		return nil, false, errors.New("no key")
	}

	key, ok := data[0].(string)
	if !ok {
		return nil, false, errors.New("bad key type")
	}

	puk.RLock()
	res, found := puk.Data[key]
	puk.RUnlock()

	return res, found, nil
}
