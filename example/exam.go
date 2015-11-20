package main

import (
	"etcdaero"
	"fmt"
	"log"
	"time"
)

/*
	Gets data from source (ex - database request) and prepare for aerospake storing.
	Function type is "type LoadFunc func([]interface{}) (map[string]interface{}, error)"
	We can pass db connection and other date with "params []interface{}".
*/
func _myFuncGetDataForCache(params []interface{}) (map[string]interface{}, error) {
	if len(params) != 2 {
		// params[0] => "my-add-param"
		// params[1] => "my-add-param-too"
		return nil, fmt.Errorf("_myFuncGetDataForCache: Bad input len(params) != 2.\n")
	}
	// Our complex data...
	return map[string]interface{}{
		"1": "Winnie - 1!",
		"2": "Pooh - 2!",
		"3": "Honey - 3!",
	}, nil
}

var cfgETCD *etcdaero.Config = &etcdaero.Config{
	AeroNamespace: "content_api",
	AeroPrefix:    "test_prefix",
	AeroHostsPots: []string{"127.0.0.1:3000"},
	EtcdPort:      4001,
	EtcdHost:      "127.0.0.1",
	EtcdEndpoints: []string{"http://127.0.0.1:4001"},
}

var keyETCD string = "my_simple_key_etcd_aero"

func main() {

	// keyETCD, cfgETCD, _myFuncGetDataForCache are required
	fmt.Printf("Start etcdaero.New\n")
	et, err := etcdaero.New(keyETCD, cfgETCD, _myFuncGetDataForCache, "my-add-param", "my-add-param-too")
	if err != nil {
		log.Fatal(err)
	}

	/*
		Starts aerospike reader.
		Default result is map[string]interface{}.
		If we need a post-processing after aerospike we
		have to use etcdaero.StartAeroReader( keyETCD, etcdaero.IAeroBody ).
		See example/exam2.go & example/src/storage/storage.go for more details.
	*/
	fmt.Printf("Starts aerospike reader.\n")
	etcdaero.StartAeroReader(keyETCD)

	// just for test
	et.SetTTL(4 * time.Second)

	// here we're making something
	time.Sleep(25 * time.Second)

	/*
		Get data from local cache.
		If we use etcdaero.IAeroBody we can get data with params
		ex: etcdaero.GetAero(keyETCD, "key")
		See example/exam2.go & example/src/storage/storage.go for more details.
	*/
	obj, find := etcdaero.GetAero(keyETCD)
	fmt.Printf("result from aerospike. FIND: %t, Data: %+v\n", find, obj)
}
