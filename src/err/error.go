package err

import (
	"fmt"
)

type HammerError struct {
	Type int
	Err  error
}

func (he *HammerError) Error() string {
	return fmt.Sprintf("%d : %s", he.Type, he.Err)
}
