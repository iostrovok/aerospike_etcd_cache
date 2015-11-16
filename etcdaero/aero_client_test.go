package etcdaero

import (
	"fmt"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func TestAeroClient(t *testing.T) {
	TestingT(t)
}

type AeroClientTetsSuite struct{}

var _ = Suite(&AeroClientTetsSuite{})

var TTL = time.Second * 10

var cfgAero *Config = &Config{
	AeroNamespace: "content_api",
	AeroPrefix:    "test_prefix",
	AeroHostsPots: []string{"127.0.0.1:3000"},
}

var key *AeroSpikeKey = &AeroSpikeKey{
	Set:  "sets",
	Pk:   "pkkey",
	Tags: []string{"tag1", "tag2", "tag3"},
}

func (s *AeroClientTetsSuite) Test_Init(c *C) {
	//c.Skip("Not now")
	as, err := NewAeroSpikeClient(cfgAero)
	c.Assert(err, IsNil)
	c.Assert(as, NotNil)
	c.Assert(as.getPolicy, NotNil)
	c.Assert(as.client, NotNil)
}

func (s *AeroClientTetsSuite) Test_putEntry_LoadEntry(c *C) {
	//c.Skip("Not now")

	data := map[string]interface{}{
		"super": 1,
		"puper": "asdsadsadasd",
	}

	entry, err := NewEtcdAeroEntry(data)
	c.Assert(err, IsNil)

	as, err := NewAeroSpikeClient(cfgAero)
	as.putEntry(key, entry, TTL)

	c.Assert(err, IsNil)
	c.Assert(as, NotNil)
	c.Assert(as.getPolicy, NotNil)
	c.Assert(as.client, NotNil)

	buf := EmptyEtcdAeroEntry()
	find := as.LoadEntry(key, buf)
	c.Check(find, Equals, true)
	c.Check(fmt.Sprintf("%s", buf.Body), Equals, `{"puper":"asdsadsadasd","super":1}`)
}
