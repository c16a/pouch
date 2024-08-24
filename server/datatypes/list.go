package datatypes

import (
	"encoding/json"
	"errors"
)

type List struct {
	head   *Element
	tail   *Element
	length int
	Values []string `json:"values"`
	Name   string   `json:"name"`
}

// Element represents a node in the linked list
type Element struct {
	data string
	next *Element
}

func (list *List) MarshalJSON() ([]byte, error) {
	return json.Marshal(list)
}

func NewList() *List {
	return &List{Values: make([]string, 0), Name: "list"}
}

func (list *List) GetName() string {
	return list.Name
}

func (list *List) GetValues() []string {
	return list.Values
}

func (list *List) LPushAll(values []string) {
	for _, value := range values {
		list.LPush(value)
	}
}

// LPush adds a new node with the given data to the start of the list
func (list *List) LPush(value string) {
	newNode := &Element{data: value}
	if list.head == nil {
		list.head = newNode
		list.tail = newNode
		return
	}
	newNode.next = list.head
	list.head = newNode
}

func (list *List) RPushAll(values []string) {
	for _, value := range values {
		list.RPush(value)
	}
}

// RPush adds a new node with the given data to the end of the list
func (list *List) RPush(data string) {
	newNode := &Element{data: data}
	if list.tail == nil {
		list.head = newNode
		list.tail = newNode
		return
	}
	list.tail.next = newNode
	list.tail = newNode
}

func (list *List) LPopN(n int) ([]string, error) {
	var result []string
	for i := 0; i < n; i++ {
		popped, ok := list.LPop()
		if !ok {
			return result, nil
		}
		result = append(result, popped)
	}
	return result, nil
}

// LPop removes and returns the first element of the list
func (list *List) LPop() (string, bool) {
	if list.head == nil {
		return "", false
	}
	data := list.head.data
	list.head = list.head.next
	if list.head == nil {
		list.tail = nil
	}
	return data, true
}

func (list *List) RPopN(n int) ([]string, error) {
	var result []string
	for i := 0; i < n; i++ {
		popped, ok := list.RPop()
		if !ok {
			return result, nil
		}
		result = append(result, popped)
	}
	return result, nil
}

// RPop removes and returns the last element of the list
func (list *List) RPop() (string, bool) {
	if list.tail == nil {
		return "", false
	}

	// If there's only one element
	if list.head == list.tail {
		data := list.head.data
		list.head = nil
		list.tail = nil
		return data, true
	}

	current := list.head
	for current.next != list.tail {
		current = current.next
	}
	data := list.tail.data
	list.tail = current
	list.tail.next = nil

	return data, true
}

// LRange returns a slice of elements from start to end (inclusive)
// If end is -1, it returns elements from start to the end of the list
func (list *List) LRange(start, end int) ([]string, error) {
	if start < 0 {
		return nil, errors.New("start index cannot be negative")
	}

	var result []string
	current := list.head
	index := 0

	for current != nil {
		if index >= start && (end == -1 || index <= end) {
			result = append(result, current.data)
		}
		if end != -1 && index > end {
			break
		}
		index++
		current = current.next
	}

	if start >= index {
		return nil, errors.New("start index out of range")
	}

	return result, nil
}

// LLen returns the number of elements in the list
func (list *List) LLen() int {
	count := 0
	current := list.head
	for current != nil {
		count++
		current = current.next
	}
	return count
}
