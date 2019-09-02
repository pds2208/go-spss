package r_times

// #cgo windows amd64 CFLAGS: -g -IC:/Users/pauld/scoop/apps/r/current/include
// #cgo windows LDFLAGS: -LC:/Users/pauld/scoop/apps/r/current/bin/x64 -lR -lRblas
// #cgo darwin amd64 CFLAGS: -g -I/Library/Frameworks/R.framework/Resources/include
// #cgo darwin LDFLAGS: -L/Library/Frameworks/R.framework/Resources/lib -lR -lRblas
// #cgo linux amd64 CFLAGS: -I/usr/share/R/include -g
// #cgo linux LDFLAGS: -L/usr/lib/R -lR
// #include <stdlib.h>
// #include "r_times.h"
import "C"
import "errors"

func Times(a, b float64) (float64, error) {
	var res = C.r_times(C.double(a), C.double(b))

	if res == -1 {
		return 0.0, errors.New("R call failed")
	}

	return float64(res), nil
}
