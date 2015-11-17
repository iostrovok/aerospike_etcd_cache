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

## Error Handling

etcd client might return three types of errors.

- context error

Each API call has its first parameter as `context`. A context can be canceled or have an attached deadline. If the context is canceled or reaches its deadline, the responding context error will be returned no matter what internal errors the API call has already encountered.

- cluster error

Each API call tries to send request to the cluster endpoints one by one until it successfully gets a response. If a requests to an endpoint fails, due to exceeding per request timeout or connection issues, the error will be added into a list of errors. If all possible endpoints fail, a cluster error that includes all encountered errors will be returned.

- response error

If the response gets from the cluster is invalid, a plain string error will be returned. For example, it might be a invalid JSON error.

Here is the example code to handle client errors:

```go
cfg := client.Config{Endpoints: []string{"http://etcd1:2379,http://etcd2:2379,http://etcd3:2379"}}
c, err := client.New(cfg)
if err != nil {
	log.Fatal(err)
}

kapi := client.NewKeysAPI(c)
resp, err := kapi.Set(ctx, "test", "bar", nil)
if err != nil {
	if err == context.Canceled {
		// ctx is canceled by another routine
	} else if err == context.DeadlineExceeded {
		// ctx is attached with a deadline and it exceeded
	} else if cerr, ok := err.(*client.ClusterError); ok {
		// process (cerr.Errors)
	} else {
		// bad cluster endpoints, which are not etcd servers
	}
}
```
