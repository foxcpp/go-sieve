package interp

type Relational string

const (
	RelGreaterThan    Relational = "gt"
	RelGreaterOrEqual Relational = "ge"
	RelLessThan       Relational = "lt"
	RelLessOrEqual    Relational = "le"
	RelEqual          Relational = "eq"
	RelNotEqual       Relational = "ne"
)

func (r Relational) CompareString(lhs, rhs string) bool {
	switch r {
	case RelGreaterThan:
		return lhs > rhs
	case RelGreaterOrEqual:
		return lhs >= rhs
	case RelLessThan:
		return lhs < rhs
	case RelLessOrEqual:
		return lhs <= rhs
	case RelEqual:
		return lhs == rhs
	case RelNotEqual:
		return lhs != rhs
	}
	return false
}

func (r Relational) CompareUint64(lhs, rhs uint64) bool {
	switch r {
	case RelGreaterThan:
		return lhs > rhs
	case RelGreaterOrEqual:
		return lhs >= rhs
	case RelLessThan:
		return lhs < rhs
	case RelLessOrEqual:
		return lhs <= rhs
	case RelEqual:
		return lhs == rhs
	case RelNotEqual:
		return lhs != rhs
	}
	return false
}

func (r Relational) CompareNumericValue(lhs, rhs *uint64) bool {
	// https://www.rfc-editor.org/rfc/rfc4790.html#section-9.1
	// nil (string not starting with a digit)
	// represents positive infinity.  inf == inf. inf > any integer.

	switch r {
	case RelGreaterThan:
		if lhs == nil {
			if rhs == nil {
				return false
			}
			return true
		}
		if rhs == nil {
			return false
		}
		return *lhs > *rhs
	case RelGreaterOrEqual:
		return !RelLessThan.CompareNumericValue(lhs, rhs)
	case RelLessThan:
		if rhs == nil {
			if lhs == nil {
				return false
			}
			return true
		}
		if lhs == nil {
			return false
		}
		return *lhs < *rhs
	case RelLessOrEqual:
		return !RelGreaterThan.CompareNumericValue(lhs, rhs)
	case RelEqual:
		if lhs == nil && rhs == nil {
			return true
		}
		if lhs != nil && rhs != nil {
			return *lhs == *rhs
		}
		return false
	case RelNotEqual:
		return !RelEqual.CompareNumericValue(lhs, rhs)
	}
	return false
}
