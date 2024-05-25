package optional

// Val contains value and flag that mark this value as empty.
type Val[T any] struct {
	hasVal bool
	value  T
}

func New[T any](val T) Val[T] {
	return Val[T]{
		hasVal: true,
		value:  val,
	}
}

// NewFromPointer create Val from pointer to some type. Nil pointer means that value not provided.
func NewFromPointer[T any](val *T) Val[T] {
	if val == nil {
		return Empty[T]()
	}

	return New(*val)
}

// Empty create container without value.
func Empty[T any]() Val[T] {
	return Val[T]{
		hasVal: false,
		value:  *new(T),
	}
}

func (v Val[T]) Get() (T, bool) { //nolint:ireturn,nolintlint // this is a concrete type. Bug in linter?
	return v.value, v.hasVal
}

// Val return internal value.
func (v Val[T]) Val() T { //nolint:ireturn
	return v.value
}

// HasVal return value of flag hasVal.
func (v Val[T]) HasVal() bool {
	return v.hasVal
}

// ValDefault returns value, if presented or defaultVal in other case.
func (v Val[T]) ValDefault(defaultVal T) T { //nolint:ireturn
	if v.hasVal {
		return v.value
	}

	return defaultVal
}

// AsPointer adapt value to pointer. It will return nil when value not provided.
func (v Val[T]) AsPointer() *T {
	if v.hasVal {
		return &v.value
	}

	return nil
}

// Set will set the value and mark that value is presented.
func (v *Val[T]) Set(val T) {
	v.hasVal = true
	v.value = val
}

// Reset will clear value and mark this as empty.
func (v *Val[T]) Reset() {
	v.hasVal = false
	v.value = *new(T)
}
