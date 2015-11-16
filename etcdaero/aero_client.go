package etcdaero

import (
	"log"
	"net"
	"strconv"
	"time"

	aerospike "github.com/aerospike/aerospike-client-go"
)

const (
	// 10,05 mins
	maxRetries          = 2
	ConnectionQueueSize = 2
)

// IEntryData data buffer for export and import
type IEntryData interface {
	Export() map[string]interface{}
	Import(data map[string]interface{}) error
}

type AeroSpikeKey struct {
	Set, Pk string
	Tags    []string
}

type AeroSpikeClient struct {
	prefix    string
	namespace string
	client    *aerospike.Client
	getPolicy *aerospike.BasePolicy
}

func NewAeroSpikeClient(cfg *Config) (*AeroSpikeClient, error) {

	clientPolicy := aerospike.NewClientPolicy()
	clientPolicy.ConnectionQueueSize = ConnectionQueueSize
	clientPolicy.LimitConnectionsToQueueSize = true
	clientPolicy.Timeout = 50 * time.Millisecond

	hosts := []*aerospike.Host{}
	for _, connStr := range cfg.AeroHostsPots {
		hostStr, portStr, err := net.SplitHostPort(connStr)
		if err != nil {
			return nil, err
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, aerospike.NewHost(hostStr, port))
	}

	client, err := aerospike.NewClientWithPolicyAndHost(clientPolicy, hosts...)
	if err != nil {
		return nil, err
	}

	getPolicy := aerospike.NewPolicy()
	getPolicy.Timeout = time.Millisecond * 50
	getPolicy.MaxRetries = maxRetries

	out := &AeroSpikeClient{
		namespace: cfg.AeroNamespace,
		prefix:    cfg.AeroPrefix,
		getPolicy: getPolicy,
		client:    client,
	}

	return out, nil
}

func (as *AeroSpikeClient) createKey(key *AeroSpikeKey) (aeroKey *aerospike.Key, err error) {
	return aerospike.NewKey(as.namespace, key.Set, as.prefix+key.Pk)
}

// PutEntry store data to cache by goroutines
func (a *AeroSpikeClient) PutEntry(key *AeroSpikeKey, data IEntryData, ttl time.Duration) {
	go a.putEntry(key, data, ttl)
}

func (as *AeroSpikeClient) Close() {
	as.client.Close()
}

// putEntry store data to cache
func (as *AeroSpikeClient) putEntry(key *AeroSpikeKey, data IEntryData, ttl time.Duration) {

	aKey, _ := as.createKey(key)

	policy := aerospike.NewWritePolicy(0, int32(ttl.Seconds()))
	policy.MaxRetries = maxRetries

	bins := data.Export()
	bins["tags"] = key.Tags
	bins["id"] = key.Pk

	if err := as.client.Put(policy, aKey, bins); err != nil {
		log.Printf("PutExternal error: %s", err)
	}

	return
}

// LoadEntry load data with bins
func (as *AeroSpikeClient) LoadEntry(key *AeroSpikeKey, buf IEntryData) bool {

	aKey, err := as.createKey(key)
	if err != nil {
		log.Println(err)
		return false
	}

	bins := buf.Export()
	binSlice := make([]string, 0, len(bins))
	for k := range bins {
		binSlice = append(binSlice, k)
	}

	rec, err := as.client.Get(as.getPolicy, aKey, binSlice...)

	if err != nil {
		return false
	}

	if rec == nil {
		// Cache not found
		return false
	}

	if err = buf.Import(rec.Bins); err != nil {
		return false
	}

	return true
}
