package integrations

import (
	"fmt"

	"github.com/zac-garby/magicalinternetpoints/lib/common"
)

type Reddit struct {
}

func (r *Reddit) GetRawPoints(account *common.Account) (map[string]int, error) {
	return nil, fmt.Errorf("not yet implemented")
}
