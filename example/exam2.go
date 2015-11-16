package main

import (
	"etcdaero"
	"fmt"
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

func _myFucnGetDataForCache(params []interface{}) (map[string]interface{}, error) {
	if len(params) != 1 { // params[0] => "eeee"
		return nil, fmt.Errorf("LoadCategoriesFromMysqlToAeroSpike: Bad input len(params) != 4.\n")
	}
	return map[string]interface{}{
		"1": "Super for 1!",
		"2": "Super for 2!",
		"3": "Super for 3!",
	}, nil
}

func main() {

	fmt.Println("Start")
	et, err := etcdaero.New(keyETCD, cfgETCD, _myFucnGetDataForCache, "eeee")

	fmt.Printf("et: %T\n", et)
	fmt.Printf("err: %s\n", err)
	fmt.Printf("et: %+v\n", et)

	etcdaero.StartAeroReader(keyETCD, storage.NewlocalStorage())

	et.SetTTL(4 * time.Second)

	time.Sleep(25 * time.Second)

	obj, find := etcdaero.GetAero(keyETCD, "ru")

	fmt.Printf("find: %t\n", find)
	fmt.Printf("obj: %T\n", obj)
	fmt.Printf("obj: %+v\n", obj)

	fmt.Println("Finish")

}
