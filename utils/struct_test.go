package utils

import (
	"fmt"
	"testing"
)

func Test_Queue(t *testing.T) {
	Queue := NewQueue()
	fmt.Println(Queue.Front())
	Queue.Push(NewNode(1))
	fmt.Println(Queue.Front(), ":", Queue.Length())
	Queue.Pop()
	Queue.Push(NewNode(2))
	fmt.Println(Queue.Pop(), ":", Queue.Length())
	Queue.Push(NewNode(3))
	fmt.Println(Queue.Pop(), ":", Queue.Length())
	Queue.Push(NewNode(4))
	fmt.Println(Queue.Pop(), ":", Queue.Length())

}
