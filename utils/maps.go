package utils

type OrderedMap[K comparable, V any] struct {
	m     map[K]V
	order []K
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{m: map[K]V{}}
}

func (q *OrderedMap[K, V]) Keys() []K {
	return q.order
}

func (q *OrderedMap[K, V]) Values() []V {
	result := []V{}
	for _, key := range q.Keys() {
		result = append(result, q.m[key])
	}
	return result
}

func (q *OrderedMap[K, V]) Insert(k K, v V) {
	q.order = append(q.order, k)
	q.m[k] = v
}

func (q *OrderedMap[K, V]) Get(k K) V {
	return q.m[k]
}
