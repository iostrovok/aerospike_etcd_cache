package main

import (
	"etcdaero"
	"fmt"
	"log"
	"storage"
	"time"
)

var cfgETCD *etcdaero.Config = &etcdaero.Config{
	AeroNamespace: "content_api",
	AeroPrefix:    "test_prefix",
	AeroHostsPots: []string{"127.0.0.1:3000"},
	EtcdPort:      4001,
	EtcdHost:      "127.0.0.1",
	EtcdEndpoints: []string{"http://127.0.0.1:4001"},
}

var keyETCD string = "my_simple_key_etcd"

func _myFuncGetDataForCache(params []interface{}) (map[string]interface{}, error) {
	if len(params) != 3 { //
		return nil, fmt.Errorf("_myFuncGetDataForCache: Bad input len(params) != 3.\n")
	}
	return map[string]interface{}{
		"1": params[0],
		"2": params[1],
		"3": params[2],
	}, nil
}

func main() {

	fmt.Println("Start")
	et, err := etcdaero.New(keyETCD, cfgETCD, _myFuncGetDataForCache, "Winnie", "Pooh", "Honey")
	if err != nil {
		log.Fatal(err)
	}

	etcdaero.StartAeroReader(keyETCD, storage.NewlocalStorage())

	et.SetTTL(4 * time.Second)

	time.Sleep(25 * time.Second)

	key := "ru"
	obj, find := etcdaero.GetAero(keyETCD, key)
	fmt.Printf("result from aerospike. FIND: %t, Data for key %s: %+v\n", find, key, obj)
}
