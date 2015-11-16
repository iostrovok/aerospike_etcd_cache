package main

import (
	"etcdaero"
	"fmt"
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

func _myFucnGetDataForCache(params []interface{}) (map[string]interface{}, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("LoadCategoriesFromMysqlToAeroSpike: Bad input len(params) != 4.\n")
	}
	return map[string]interface{}{"string": "Super!"}, nil
}

func main() {

	fmt.Println("Start")
	et, err := etcdaero.New(keyETCD, cfgETCD, _myFucnGetDataForCache, "eeee")

	fmt.Printf("et: %T\n", et)
	fmt.Printf("err: %s\n", err)
	fmt.Printf("et: %+v\n", et)

	etcdaero.StartAeroReader(keyETCD)

	et.SetTTL(1 * time.Second)

	time.Sleep(5 * time.Second)

	obj, find := etcdaero.GetAero(keyETCD)

	fmt.Printf("find: %t\n", find)
	fmt.Printf("obj: %T\n", obj)
	fmt.Printf("obj: %+v\n", obj)

	fmt.Println("Finish")

}
