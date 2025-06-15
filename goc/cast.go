package goc

func castGeneric[B, A any](a A) (B, error) {
	var zero B

	b, ok := any(a).(B)
	if !ok {
		return zero, ErrTypeAssertion
	}

	return b, nil
}
