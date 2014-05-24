package libcleo

import "testing"
import "proto"
import "fmt"

func TestRetrieval1(t *testing.T) {
	fmt.Println(rid2cleo(1))
	fmt.Println(rid2cleo(2))
	
	if rid2cleo(1) != proto.ChampionType_ANNIE {
		t.Fail()
	}
}
