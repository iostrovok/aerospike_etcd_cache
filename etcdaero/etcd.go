package etcdaero

import (
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/coreos/go-etcd/etcd"
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
	etcdLockTTL uint64
	timerTTL    time.Duration
	sleepTTL    time.Duration
	AeroTTL     time.Duration
	client      *etcd.Client
	cfg         *Config
	key         string
	value       string
	stopC       chan os.Signal
	delta       float64
	//Aero        *AeroSpikeClient
	Aero *AeroChecker
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

	ea._initCaching(f, faces)

	return ea, err
}

/*
	>>>>>>>>>>>>>>>>>>>>> CONFIG FUNCTION
*/

func (ea *EtcdAero) SetTTL(timerTTL time.Duration) {

	ea.timerTTL = timerTTL
	ea.etcdLockTTL = uint64(timerTTL * 3 / 2 / time.Second)
	ea.AeroTTL = timerTTL * 10

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
		log.Printf("Could not get hostname: %v", err)
		return err
	}

	ea.value = fmt.Sprintf("%s%d", hostname, ea.cfg.EtcdPort)
	ea.client = etcd.NewClient(ea.cfg.EtcdEndpoints)

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

func (ea *EtcdAero) getLock() bool {
	res, err := ea.client.RawCreate(ea.key, ea.value, ea.etcdLockTTL)
	if err != nil {
		log.Printf("Error while trying to get lock for Pages Caching: %v", err)
		return false
	}
	return res.StatusCode == http.StatusCreated
}

func (ea *EtcdAero) releaseLock() {
	_, err := ea.client.CompareAndDelete(ea.key, ea.value, 0)
	if err != nil {
		log.Printf("Error while releasing lock for Pages Caching: %v", err)
	}
}

func (ea *EtcdAero) renewLock() bool {
	_, err := ea.client.RawCompareAndSwap(ea.key, ea.value, ea.etcdLockTTL, ea.value, 0)
	if err != nil {
		log.Printf("Error while trying to renew lock for Pages Caching: %v", err)
		return false
	}
	return true
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

	//ea.Aero.PutEntry(cacheKey, pass, ea.AeroTTL)
	ea.Aero.Put(cacheKey, pass, ea.AeroTTL)
	ea.Aero.ReLoad()

	return nil
}
