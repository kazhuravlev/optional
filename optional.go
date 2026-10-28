// Package optional provides a generic container that distinguishes "no value"
// from a zero value, with JSON, YAML and database/sql support.
package optional

// Val contains value and flag that mark this value as empty.
//
// The zero value of Val is an empty container, identical to Empty.
type Val[T any] struct {
	hasVal bool
	value  T
}

// New create container with the given value.
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

// Get returns the value and whether it is presented.
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

// IsZero reports whether the container has no value.
//
// It is what makes `json:",omitzero"` (Go 1.24+) and `yaml:",omitempty"`
// skip empty containers: both encoders consult this method before falling
// back to reflection, and reflection cannot see the private fields of Val.
func (v Val[T]) IsZero() bool {
	return !v.hasVal
}

// ValDefault returns value, if presented or defaultVal in other case.
func (v Val[T]) ValDefault(defaultVal T) T { //nolint:ireturn
	if v.hasVal {
		return v.value
	}

	return defaultVal
}

// AsPointer adapt value to pointer. It will return nil when value not provided.
//
// The pointer refers to a copy of the value: writing through it does not
// modify the original container.
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
