package server

import (
	"testing"
)

func TestEverything(t *testing.T) {
	pi, err := AddPool("a", "something.seeddee")
	if err != nil {
		t.Errorf("Got error %v\n", err)
	}
	if pi.Name != "a" ||
		pi.Template != "something.seeddee+0" ||
		pi.Used != 0 ||
		pi.Max != 70728100 ||
		pi.Closed != false {
		t.Errorf("%v does not match expected\n")
	}

	pools := AllPools()
	if len(pools) != 1 || pools[0] != "a" {
		t.Errorf("Wrong pool list %v\n", pools)
	}

}
