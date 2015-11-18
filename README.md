aerospike_etcd_cache/etcdaero is the Go library for etcd [github.com/coreos/etcd/client] and aerospike [github.com/aerospike/aerospike-client-go].

## Install

```bash
go get -u github.com/iostrovok/aerospike_etcd_cache/etcdaero
```

Goals: 

1) etcdaero gets data from source by one node only. When etcdaero selects node it uses locking with etcd/client.

2) etcdaero pushs data to aerospike.

3) Each node gets date from aerospike and prepares for using.


## Usage

```go
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
func _myFucnGetDataForCache(params []interface{}) (map[string]interface{}, error) {
	if len(params) != 2 {
		// params[0] => "my-add-param"
		// params[1] => "my-add-param-too"
		return nil, fmt.Errorf("LoadCategoriesFromMysqlToAeroSpike: Bad input len(params) != 2.\n")
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

	// keyETCD, cfgETCD, _myFucnGetDataForCache are required
	fmt.Printf("Start etcdaero.New\n")
	et, err := etcdaero.New(keyETCD, cfgETCD, _myFucnGetDataForCache, "my-add-param", "my-add-param-too")
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

```

