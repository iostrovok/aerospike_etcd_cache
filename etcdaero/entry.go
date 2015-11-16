package etcdaero

import (
	"encoding/json"
	"errors"
)

// EtcdAeroEntry app response
type EtcdAeroEntry struct {
	Body []byte
}

var ErrIncorrectDataFormat = errors.New("Incorrect data format error")

// NewEtcdAeroEntry
func EmptyEtcdAeroEntry() *EtcdAeroEntry {
	return &EtcdAeroEntry{
		Body: []byte(``),
	}
}

// NewEtcdAeroEntry
func NewEtcdAeroEntry(data map[string]interface{}) (*EtcdAeroEntry, error) {

	body, err := json.Marshal(data)

	return &EtcdAeroEntry{
		Body: body,
	}, err
}

// Import data from map[string]interface{} to EtcdAeroEntry
func (entry *EtcdAeroEntry) Import(b map[string]interface{}) error {

	if result, find := b["body"]; find {
		if result, ok := result.([]byte); ok {
			entry.Body = result
			return nil
		}
	}

	return ErrIncorrectDataFormat
}

// Export data from EtcdAeroEntry to map[string]interface{}
func (entry EtcdAeroEntry) Export() map[string]interface{} {
	return map[string]interface{}{
		"body": entry.Body,
	}
}
