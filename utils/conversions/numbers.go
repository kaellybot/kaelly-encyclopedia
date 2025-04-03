package conversions

import (
	"fmt"
	"math"
)

func Int64ToInt32(value int64) (int32, error) {
	if value > math.MaxInt32 || value < math.MinInt32 {
		return 0, fmt.Errorf("value %d overflows int32 range", value)
	}

	return int32(value), nil
}
