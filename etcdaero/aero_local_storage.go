package etcdaero

import (
	"encoding/json"
	"sync"
)

type localAeroStorage struct {
	sync.RWMutex
	IAeroBody
	Data interface{}
}

var Singleton *localAeroStorage

func NewlocalAeroStorage() localAeroStorage {
	if Singleton == nil {
		Singleton = &localAeroStorage{
			Data: nil,
		}
	}
	return *Singleton
}

func (puk localAeroStorage) ReNew(data []byte) error {
	return Singleton._ReNew(data)
}

func (puk *localAeroStorage) _ReNew(data []byte) error {
	d := map[string]interface{}{}
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	puk.Lock()
	puk.Data = d
	puk.Unlock()

	return nil
}

func (puk localAeroStorage) Get(data []interface{}) (interface{}, bool, error) {
	return Singleton._Get(data)
}

func (puk *localAeroStorage) _Get(data []interface{}) (interface{}, bool, error) {

	puk.RLock()
	res := puk.Data
	puk.RUnlock()

	found := false

	if res != nil {
		found = true
	}

	return res, found, nil
}
