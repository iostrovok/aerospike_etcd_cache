package etcdaero

import (
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/etcd/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)

type LoadFunc func([]interface{}) (map[string]interface{}, error)

type Config struct {
	AeroNamespace string
	AeroPrefix    string
	AeroHostsPots []string

	EtcdPort      int
	EtcdHost      string
	EtcdEndpoints []string
}

// 17 min 17 sec
const defTimerTTL = 17 * 61 * time.Second

type EtcdAero struct {
	etcdLockTTL time.Duration
	timerTTL    time.Duration
	sleepTTL    time.Duration
	AeroTTL     time.Duration
	client      client.Client
	clientKey   client.KeysAPI
	cfg         *Config
	key         string
	value       string
	stopC       chan os.Signal
	delta       float64
	//Aero        *AeroSpikeClient
	Aero      *AeroChecker
	prevIndex uint64
}

// New - creates new object
func New(key string, cfg *Config, f LoadFunc, faces ...interface{}) (*EtcdAero, error) {

	aeroClient, err := NewAeroSpikeClient(cfg)
	if err != nil {
		return nil, err
	}

	aero := InitAeroChecker(aeroClient)

	ea := &EtcdAero{
		key:  key,
		cfg:  cfg,
		Aero: aero,
	}

	ea.SetTTL(defTimerTTL)
	ea.stopC = make(chan os.Signal, 1)
	signal.Notify(ea.stopC, os.Interrupt, os.Kill)

	if err := ea._init(); err != nil {
		aeroClient.Close()
		return nil, err
	}

	ea._initCaching(f, faces...)

	return ea, err
}

/*
	>>>>>>>>>>>>>>>>>>>>> CONFIG FUNCTION
*/

func (ea *EtcdAero) SetTTL(timerTTL time.Duration) {

	ea.timerTTL = timerTTL
	ea.etcdLockTTL = timerTTL * 3 / 2
	ea.AeroTTL = timerTTL * 5

	// randomize sleep ttl form 0.8 to 1.2 of original timerTTL / 2
	crc32q := crc32.MakeTable(0xD5828281)
	d := 1.0 + float64(int64(crc32.Checksum([]byte(ea.value), crc32q))%1000-500)/2500.0
	ea.sleepTTL = time.Duration(float64(timerTTL) * d / 2)
}

func (ea *EtcdAero) Key(key string) {
	ea.key = key
}

func (ea *EtcdAero) _init() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	ea.value = fmt.Sprintf("%s%d", hostname, ea.cfg.EtcdPort)

	initCfg := client.Config{
		Endpoints: ea.cfg.EtcdEndpoints,
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: 2 * time.Second,
	}

	ea.client, err = client.New(initCfg)
	if err != nil {
		log.Fatal(err)
	}
	ea.clientKey = client.NewKeysAPI(ea.client)

	return nil
}

// _initCaching starts nexy job
func (ea *EtcdAero) _initCaching(f LoadFunc, faces ...interface{}) error {
	go func(ea *EtcdAero, f LoadFunc) {
		for {
			ea._make(f, faces...)
			time.Sleep(ea.sleepTTL)
		}
	}(ea, f)

	return nil
}

/*
	CONFIG FUNCTION <<<<<<<<<<<<<<<<<<<<<
*/

func (ea *EtcdAero) _updateResponse(resp *client.Response, err error) bool {
	if err == nil {
		ea.prevIndex = resp.Index
		return true
	}
	ea.prevIndex = 0
	return false
}

func (ea *EtcdAero) getLock() bool {
	resp, err := ea.clientKey.Set(context.Background(), ea.key, ea.value, _setOptions("", 0, ea.etcdLockTTL))
	return ea._updateResponse(resp, err)
}

func (ea *EtcdAero) renewLock() bool {
	resp, err := ea.clientKey.Set(context.Background(), ea.key, ea.value, _setOptions(ea.value, ea.prevIndex, ea.etcdLockTTL))
	return ea._updateResponse(resp, err)
}

func (ea *EtcdAero) releaseLock() {
	// Ignore any errors
	ea.clientKey.Delete(context.Background(), ea.key, _deleteOptions(ea.value))
}

func (ea *EtcdAero) _make(f LoadFunc, faces ...interface{}) {
	if !ea.getLock() {
		return
	}

	data, err := f(faces)
	if err == nil {
		err = ea._putAero(data)
	}

	if err != nil {
		log.Printf("Error while pages caching %v", err)
		ea.releaseLock()
		return
	}

	for {
		select {
		case <-ea.stopC:
			ea.releaseLock()
		case <-time.Tick(ea.timerTTL):
			data, err := f(faces)

			if err == nil {
				err = ea._putAero(data)
			}

			if err != nil {
				ea.releaseLock()
				return
			}

			if !ea.renewLock() {
				return
			}
		}
	}
}

func (ea *EtcdAero) _putAero(data map[string]interface{}) error {

	pass, err := NewEtcdAeroEntry(data)
	if err != nil {
		return err
	}

	cacheKey := &AeroSpikeKey{
		Set: ea.key,
		Pk:  ea.key,
	}

	ea.Aero.Put(cacheKey, pass, ea.AeroTTL)
	ea.Aero.ReLoad()

	return nil
}

func _setOptions(prevVal string, prevIndex uint64, ttl time.Duration) *client.SetOptions {

	out := &client.SetOptions{
		TTL:       ttl,
		Dir:       false,
		PrevExist: client.PrevNoExist,
	}

	if prevVal != "" {
		out.PrevValue = prevVal
		out.PrevExist = client.PrevExist
	}

	if prevIndex > 0 {
		out.PrevIndex = prevIndex
		out.PrevExist = client.PrevExist
	}

	return out
}

func _deleteOptions(prevVal string) *client.DeleteOptions {
	out := &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	}
	if prevVal != "" {
		out.PrevValue = prevVal
	}

	return out
}
