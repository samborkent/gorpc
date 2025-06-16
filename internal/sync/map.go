package sync

import (
	"cmp"
	"iter"
	"sync"
)

type Map[K comparable, V any] struct {
	m sync.Map
}

func (m *Map[K, V]) Clear() {
	m.m.Clear()
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.m.CompareAndDelete(key, old)
}

func (m *Map[K, V]) CompareAndSwap(key K, old V, new V) (swapped bool) {
	return m.m.CompareAndSwap(key, old, new)
}

func (m *Map[K, V]) Delete(key K) {
	m.m.Delete(key)
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	val, ok := m.m.Load(key)
	if !ok {
		return value, false
	}

	value, ok = val.(V)
	if !ok {
		return value, false
	}

	return value, true
}

// TODO: fix
func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	val, ok := m.m.LoadAndDelete(key)
	if !ok {
		return value, false
	}

	value, ok = val.(V)
	if !ok {
		return value, false
	}

	return value, true
}

// TODO: fix
func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	val, ok := m.m.LoadOrStore(key, value)
	if !ok {
		return actual, false
	}

	actual, ok = val.(V)
	if !ok {
		return actual, false
	}

	return actual, true
}

// TODO: fix
func (m *Map[K, V]) Range() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.m.Range(func(key, value any) bool {
			k, okKey := key.(K)
			v, okVal := value.(V)

			if !cmp.Or(okKey, okVal) {
				return false
			}

			return yield(k, v)
		})
	}
}

func (m *Map[K, V]) Store(key K, value V) {
	m.m.Store(key, value)
}

// TODO: fix
func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	val, ok := m.m.Swap(key, value)
	if !ok {
		return previous, false
	}

	previous, ok = val.(V)
	if !ok {
		return previous, false
	}

	return previous, true
}
