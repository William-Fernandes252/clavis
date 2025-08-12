package data

import (
	"encoding/json"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

const listDataType = Type("list")

type node struct {
	// next and prev are pointers to the next and previous nodes in the list
	// They are used to maintain the linked list structure.
	// To simplify the implementation, internally a list is implemented
	// as a ring, such that &l.root == &l.tail.next and &l.tail == &l.root.prev.
	// This allows for easy traversal in both directions.
	next, prev *node

	// value is the string value stored in this node
	value string

	// list is the List this node belongs to
	list *List
}

// Next returns the next node in the list.
// If the node is the tail of the list, it returns nil.
// If the list is nil, it returns the next node in the ring.
// This allows for easy traversal of the list.
func (n *node) Next() *node {
	if p := n.next; n.list != nil && p != &n.list.root {
		return p
	}
	return n.next
}

// Prev returns the previous node in the list.
// If the node is the head of the list, it returns nil.
// If the list is nil, it returns the previous node in the ring.
// This allows for easy traversal of the list.
func (n *node) Prev() *node {
	if p := n.prev; n.list != nil && p != &n.list.root {
		return p
	}
	return n.prev
}

// List represents a circular linked list.
//
// Clavis lists are linked lists of string values.
// They can be used to:
//   - Implement stacks and queues.
//   - Build queue management for background jobs.
type List struct {
	root   node
	length int
}

// NewList creates a new List instance.
// It initializes the root node and sets the length to 0.
func NewList() *List {
	return new(List).Init()
}

// Init initializes the List instance.
// It sets the root node's next and prev pointers to point to itself,
// effectively creating an empty circular linked list.
// It also sets the length of the list to 0.
func (l *List) Init() *List {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.length = 0
	return l
}

// Length returns the number of elements in the list.
// It returns 0 if the list is nil.
func (l *List) Length() int {
	if l == nil {
		return 0
	}
	return l.length
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List) Back() *node {
	if l.length == 0 {
		return nil
	}
	return l.root.prev
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List) Front() *node {
	if l.length == 0 {
		return nil
	}
	return l.root.next
}

// insert inserts n after at, increments l.len, and returns e.
func (l *List) insert(n, at *node) *node {
	n.prev = at
	n.next = at.next
	n.prev.next = n
	n.next.prev = n
	n.list = l
	l.length++
	return n
}

// insertValue is a convenience wrapper for insert(&node{Value: v}, at).
func (l *List) insertValue(v string, at *node) *node {
	return l.insert(&node{value: v}, at)
}

// remove removes n from its list, decrements l.len
func (l *List) remove(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
	n.next = nil // avoid memory leaks
	n.prev = nil // avoid memory leaks
	n.list = nil
	l.length--
}

// move moves n to next to at.
func (l *List) move(n, at *node) {
	if n == at {
		return
	}
	n.prev.next = n.next
	n.next.prev = n.prev

	n.prev = at
	n.next = at.next
	n.prev.next = n
	n.next.prev = n
}

// Remove removes e from l if e is an element of list l.
// It returns the element value n.Value.
// The element must not be nil.
func (l *List) Remove(n *node) string {
	if n.list == l {
		// if n.list == l, l must have been initialized when n was inserted
		// in l or l == nil (n is a zero node) and l.remove will crash
		l.remove(n)
	}
	return n.value
}

// PushFront inserts a new element n with value v at the front of list l and returns n.
func (l *List) PushFront(v string) *node {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element n with value v at the back of list l and returns n.
func (l *List) PushBack(v string) *node {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element n with value v immediately before mark and returns n.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List) InsertBefore(v string, mark *node) *node {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element n with value v immediately after mark and returns n.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List) InsertAfter(v string, mark *node) *node {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element n to the front of list l.
// If n is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List) MoveToFront(n *node) {
	if n.list != l || l.root.next == n {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(n, &l.root)
}

// MoveToBack moves element n to the back of list l.
// If n is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List) MoveToBack(n *node) {
	if n.list != l || l.root.prev == n {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(n, l.root.prev)
}

// MoveBefore moves element n to its new position before mark.
// If n or mark is not an element of l, or n == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List) MoveBefore(n, mark *node) {
	if n.list != l || n == mark || mark.list != l {
		return
	}
	l.move(n, mark.prev)
}

// MoveAfter moves element n to its new position after mark.
// If n or mark is not an element of l, or n == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List) MoveAfter(n, mark *node) {
	if n.list != l || n == mark || mark.list != l {
		return
	}
	l.move(n, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List) PushBackList(other *List) {
	l.lazyInit()
	for i, n := other.Length(), other.Front(); i > 0; i, n = i-1, n.Next() {
		l.insertValue(n.value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List) PushFrontList(other *List) {
	l.lazyInit()
	for i, n := other.Length(), other.Back(); i > 0; i, n = i-1, n.Prev() {
		l.insertValue(n.value, &l.root)
	}
}

func (l *List) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// toArray converts the list to a slice of strings.
// This is useful for serialization or when a simple representation of the list is needed.
func (l *List) toArray() []string {
	if l == nil || l.length == 0 {
		return nil
	}
	arr := make([]string, 0, l.length)
	for n := l.Front(); n != &l.root; n = n.Next() {
		arr = append(arr, n.value)
	}
	return arr
}

// fromArray creates a new List from a slice of strings.
// This is useful for deserialization or when initializing a list with predefined values.
func fromArray(arr []string) *List {
	l := NewList()
	for _, v := range arr {
		l.PushBack(v)
	}
	return l
}

// ListCodec implements the Data interface for lists of strings.
type ListCodec struct{}

// NewListCodec creates a new ListCodec instance.
func NewListCodec() *ListCodec {
	return &ListCodec{}
}

// Type returns the type name ("list")
func (l *ListCodec) Type() Type {
	return listDataType
}

// Serialize converts the list value into a byte slice for storage
func (l *ListCodec) Serialize(value *List) ([]byte, errors.Error) {
	data, err := json.Marshal(value.toArray())
	if err != nil {
		return nil, NewDataError("serialization-failed", "Failed to serialize list data", err)
	}
	return data, nil
}

// Deserialize reads the bytes back into the list data type
// It returns an empty list if the input is nil
func (l *ListCodec) Deserialize(raw []byte) (*List, errors.Error) {
	var value []string

	jsonErr := json.Unmarshal(raw, &value)
	if jsonErr != nil {
		return nil, NewDataError("deserialization-failed", "Failed to deserialize list data", jsonErr)
	}
	return fromArray(value), nil
}

var _ Codec[*List] = (*ListCodec)(nil)
