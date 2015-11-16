package etcdaero

import (
	"fmt"
	. "gopkg.in/check.v1"
	"testing"
)

func TestEntry(t *testing.T) {
	TestingT(t)
}

type EntryTestsSuite struct{}

var _ = Suite(&EntryTestsSuite{})

func (s *EntryTestsSuite) Test_NewEtcdAeroEntry(c *C) {
	//c.Skip("Not now")

	data := map[string]interface{}{
		"super": 1,
		"puper": "asdsadsadasd",
	}

	entry, err := NewEtcdAeroEntry(data)
	c.Assert(err, IsNil)
	c.Assert(entry, NotNil)

	//c.Check(b.Len, Equals, 18)
}

func (s *EntryTestsSuite) Test_EmptyEtcdAeroEntry(c *C) {
	//c.Skip("Not now")

	entry := EmptyEtcdAeroEntry()
	c.Assert(entry, NotNil)
	c.Check(fmt.Sprintf("%s", entry.Body), Equals, fmt.Sprintf("%s", []byte{}))
}

func (s *EntryTestsSuite) Test_Export(c *C) {
	//c.Skip("Not now")

	data := map[string]interface{}{
		"super": 1,
		"puper": "asdsadsadasd",
	}

	entry, err := NewEtcdAeroEntry(data)
	res := entry.Export()
	c.Assert(err, IsNil)
	c.Assert(entry, NotNil)
	c.Assert(res, NotNil)
	c.Check(fmt.Sprintf("%s", res), Equals, `map[body:{"puper":"asdsadsadasd","super":1}]`)
}

func (s *EntryTestsSuite) Test_Import(c *C) {
	//c.Skip("Not now")

	data := map[string]interface{}{
		"body": []byte(`map[body:{"puper":"asdsadsadasd","super":1}]`),
	}

	entry, err := NewEtcdAeroEntry(data)
	c.Assert(err, IsNil)

	err = entry.Import(data)
	c.Assert(err, IsNil)
	c.Check(fmt.Sprintf("%s", entry.Body), Equals, `map[body:{"puper":"asdsadsadasd","super":1}]`)
}
