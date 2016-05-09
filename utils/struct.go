package utils

type node struct {
	data interface{}
	next *node
}

func NewNode(data interface{}) *node {
	return &node{
		data: data,
	}
}
func (n *node) Data() interface{} {
	return n.data
}

type Queue struct {
	head   *node
	tail   *node
	length int
}

func NewQueue() *Queue {
	q := &Queue{}
	q.head = &node{}
	q.tail = q.head
	return q
}

func (q *Queue) Empty() bool {
	if q.head == q.tail {
		return true
	}
	return false
}

func (q *Queue) Push(data interface{}) {
	node := &node{
		data: data,
	}
	q.tail.next = node
	q.tail = node
	q.length++
}

func (q *Queue) Front() interface{} {
	if q.Empty() {
		return nil
	}
	return q.head.next.data
}

func (q *Queue) Back() interface{} {
	if q.Empty() {
		return nil
	}
	return q.tail.data
}

func (q *Queue) Pop() interface{} {
	if q.Empty() {
		return nil
	}
	node := q.head.next
	if node == q.tail {
		q.tail = q.head
	}
	q.head.next = node.next
	node.next = nil
	q.length--
	return node.data
}

func (q *Queue) Length() int {
	return q.length
}
