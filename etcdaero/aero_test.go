package etcdaero

import (
	"fmt"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestAero(t *testing.T) {
	TestingT(t)
}

type AeroTetsSuite struct{}

var _ = Suite(&AeroTetsSuite{})

//

var keyAero string = "my_simple_key"

type aeroBodyTest struct {
	IAeroBody

	list []string
}

func (aob aeroBodyTest) Get(data []interface{}) (interface{}, bool, error) {

	fmt.Printf("---> (aob aeroBodyTest) Get(data []interface{}) (interface{}, bool, error)\n")

	return "my_string", true, nil
}

func (aob aeroBodyTest) ReNew(data []byte) error {

	fmt.Printf("---> (aob aeroBodyTest) ReNew(data []byte) error: %s\n", data)

	aob.list = append(aob.list, string(data))
	return nil
}

//

func (s *AeroTetsSuite) Test_InitAeroChecker(c *C) {
	c.Skip("Not now")

	_cleanSingletonAero()

	conn, _ := NewAeroSpikeClient(cfgAero)

	as := InitAeroChecker(conn)
	c.Assert(as, NotNil)
}

func (s *AeroTetsSuite) Test_AddAero_ReLoadAero(c *C) {
	c.Skip("Not now")

	_cleanSingletonAero()

	conn, _ := NewAeroSpikeClient(cfgAero)
	as := InitAeroChecker(conn)

	obj := aeroBodyTest{
		list: []string{},
	}

	as._add(keyAero, obj)
	// as.ReLoad()

	as.SetTTL(1 * time.Second)

	fmt.Printf("as: %+v\n", as)
	fmt.Printf("obj: %+v\n", obj)

	time.Sleep(4 * time.Second)

	//c.Assert(err, IsNil)
	c.Assert(as, NotNil)

	//_cleanSingletonAero()
	c.Assert(nil, NotNil)
}
