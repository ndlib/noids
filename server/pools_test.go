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

func TestMint(t *testing.T) {
    _, err := AddPool("mint", ".sd")
    if err != nil {
        t.Fatalf("%v\n", err)
        return
    }
    table := []struct {
        pool string
        count int
        result []string
        err error
    } { {"mint", 5, []string{"0", "1", "2", "3", "4"}, nil},
        {"mint", 1, []string{"5"}, nil},
        {"mint", 20, []string{"6", "7", "8", "9"}, nil},
        {"mint", 20, []string{}, PoolClosed},
    }
    for _, z := range table {
        result, err := PoolMint(z.pool, z.count)
        if err != z.err {
            t.Errorf("%v\n", err)
        }
        if len(result) != len(z.result) {
            t.Errorf("%v != %v\n", result, z.result)
        }
        for i := range result {
            if result[i] != z.result[i] {
                t.Errorf("%v != %v\n", result, z.result)
                break
            }
        }
    }
}
