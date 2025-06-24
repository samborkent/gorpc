package goc

import (
	"fmt"
	"reflect"
)

// Adapted from encoding/gob/type.go -> validUserType, ln55-77
func numIndirections(t reflect.Type) (int, error) {
	baseType := t
	// A type that is just a cycle of pointers (such as type T *T) cannot
	// be represented in gobs, which need some concrete data. We use a
	// cycle detection algorithm from Knuth, Vol 2, Section 3.1, Ex 6,
	// pp 539-540.  As we step through indirections, run another type at
	// half speed. If they meet up, there's a cycle.
	slowType := baseType // walks half as fast as ut.base

	indirections := 0

	for {
		pt := baseType
		if pt.Kind() != reflect.Pointer {
			break
		}

		baseType = pt.Elem()
		if baseType == slowType { // base type lapped slow type
			// recursive pointer type.
			return 0, fmt.Errorf("cannot represent recursive pointer type %s", baseType.String())
		}

		if indirections%2 == 0 {
			slowType = slowType.Elem()
		}

		indirections++
	}

	return indirections, nil
}
