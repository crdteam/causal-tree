# API

## Class diagram

```mermaid
classDiagram
    class Container
    <<interface>> Container
    Container: Len() int
    Container: Cursor() Cursor
    
    class Value
    <<interface>> Value
    Value: Snapshot()

    class String
    Container <|.. String
    Value <|.. String
    String: Snapshot() str
    
    class List
    Container <|.. List
    Value <|.. List
    List: Snapshot() []any
    
    class CausalTree
    Register <|.. CausalTree

    class Cursor
    <<interface>> Cursor
    Cursor: Index(int i)
    Cursor: Delete()
    
    class Register
    <<interface>> Register
    Register: SetString() String
    Register: SetCounter() Counter
    Register: SetList() List
    Register: Clear()
    Register: Value() Value

    class Char
    Char: rune ch
    
    class StringCursor
    Cursor <|.. StringCursor
    StringCursor: Insert(rune ch) Char
    StringCursor: Value() Char
    
    class ListCursor
    Cursor <|.. ListCursor
    ListCursor: Insert() ListElement
    ListCursor: Value() ListElement
    
    class ListElement
    Register <|.. ListElement
    
    class Counter
    Value <|.. Counter
    Counter: Snapshot() int32
    Counter: Increment(int32 x)
    Counter: Decrement(int32 x)
```

## Relationship diagrams

```mermaid
classDiagram
    class CausalTree
    CausalTree .. String
    CausalTree .. Counter
    CausalTree .. List
    CausalTree: SetString() String
    CausalTree: SetCounter() Counter
    CausalTree: SetList() List
    CausalTree: Clear()
    CausalTree: Value() Value
    
    class ListElement
    ListCursor --> ListElement
    ListElement .. String
    ListElement .. Counter
    ListElement .. List
    ListElement: SetString() String
    ListElement: SetCounter() Counter
    ListElement: SetList() List
    ListElement: Clear()
    ListElement: Value() Value

    class String
    String: Snapshot() str
    String: Len() int
    String: Cursor() StringCursor

    class List
    List: Snapshot() []any
    List: Len() int
    List: Cursor() ListCursor

    class Counter
    Counter: Snapshot() int32
    Counter: Increment(int32 x)
    Counter: Decrement(int32 x)

    class StringCursor 
    String --> StringCursor
    StringCursor: Insert(rune ch) Char
    StringCursor: Value() Char
    
    class ListCursor
    List --> ListCursor
    ListCursor: Insert() ListElement
    ListCursor: Value() ListElement

    
    class Char
    StringCursor --> Char
    Char: rune ch
```

## Definitions

A _Value_ is a persistent structure containing data. They may be read and converted
into the data they represent. 

- _String_: sequence of Unicode codepoints (represents `string`)
- _Counter_: mutable integer (represents `int32`)
- _List_: sequence of values (represents `[]interface{}`)

A _Container_ is a persistent data structure representing a collection of other values.
_String_ and _List_ are containers. They are composed of _cells_, representing each
element of the container.

- _Char_: Unicode codepoint (cell of _String_)
- _Element_: definite position on a list (cell of _List_)

A _Register_ is a persistent structure containing a single value. The last written value, for
some definition of "last", is the accepted value.

- _CausalTree_: the whole tree contains a single value, or none.
- _Element_: each position on a list may be set or empty.

A _Cursor_ is a mutable volatile pointer to a cell within a container. It may be
used to move across the container, read the pointed value, insert a new value, or delete
an existing value. The cursor moves automatically depending on the executed operation.

## Example

```go
func main() {
    // Creates a new tree and sets it to a list.
    t := crdt.NewCausalTree()
    list0 := t.SetList()
    cur0 := list0.ListCursor()
    {
        // Insert a String in first position.
        elem1 := cur0.Insert()
        s1 := elem1.SetString()

        // Inserts the characters 'c', 'r', 't' in string.
        cur1 := s1.StringCursor()
        cur1.Insert('c')
        cur1.Insert('r')
        cur1.Insert('t')

        // Move to index 1 and insert 'd', forming "crdt".
        cur1.Index(1)
        cur1.Insert('d')
    }
    {
        // Inserts a Counter in the second list position.
        elem2 := cur0.Insert()
        c2 := elem2.SetCounter()

        // Mutate the counter.
        c2.Increment(10)
        c2.Decrement(4)

        // Delete the counter, but keep the element.
        elem2.Clear()
    }
    {
        // Inserts a List in the third position.
        elem3 := cur0.Insert()
        c3 := elem3.SetList()

        // Inserts the numbers 1, 2, 3, 4 in the list.
        c3.Insert().SetCounter().Increment(1)
        c3.Insert().SetCounter().Increment(2)
        c3.Insert().SetCounter().Increment(3)
        c3.Insert().SetCounter().Increment(4)

        // Delete the element in index 2.
        c3.Index(2)
        c3.Delete()
    }

    // Returns a snapshot of the whole tree.
    t.Snapshot() // ["crdt", null, [1, 2, 4]]
}
```

## Interfaces

```go
// Value represents a structure that may be converted to concrete data.
// Each one has a method "Snapshot()" with appropriate return type.
type Value interface {
    isValue()
}

// Register contains a single value or none at all.
type Register interface {
    // SetString sets the register to an empty string.
    SetString() *String
    // SetString sets the register to a zeroed counter.
    SetCounter() *Counter
    // SetString sets the register to an empty list.
    SetList() *List
    // Clear resets the register to an empty state.
    Clear()
    // Value returns the underlying value.
    Value() Value
}

// Container represents a collection of values.
type Container interface {
    // Len walks the container and returns the number of elements.
    Len() int
    // Cursor returns the container's cursor initialized to its starting position.
    Cursor() Cursor
}

// Cursor represents a pointer to a container's element.
// Concrete cursors have an Insert() method with appropriate parameters and return type.
// Concrete cursors have a Value() method with appropriate return type.
type Cursor interface {
    // Index moves the cursor to the i-th element. It panics if i is out of bounds.
    Index(i int)
    // Delete removes the pointed element from the collection. The cursor is moved to the
    // previous element, or the container's head.
    Delete()
}
```

## Concrete types

### `CausalTree`

```go
// CausalTree represents a CRDT structure in the causal tree model.
type CausalTree struct {
    atoms []Atom
}

// CausalTree can be set and reset, implementing a Register.
func (*CausalTree) SetString() *String   { ... }
func (*CausalTree) SetCounter() *Counter { ... }
func (*CausalTree) SetList() *List       { ... }
func (*CausalTree) Clear()               { ... }
func (*CausalTree) Value() Value         { ... }

// treePosition represents a given atom's position in a causal tree.
type treePosition struct {
    tree   *CausalTree
    atomID AtomID

    // treePosition stores the last known position within the atoms slice to speed-up searching:
    // since atoms can only be inserted, its actual position may only be at or to the right
    // of the latest known position.
    lastKnownPos int
}
```

### `String`

```go
// String and StringCursor are just pointers to a location within the string.
type String struct { treePosition }
func (*String) isValue() {}

// StringCursor returns a cursor to the string's initial position.
func (s *String) StringCursor() *StringCursor {
    return &StringCursor{s.tree, s.atomID, s.lastKnownPos}
}

// Cursor returns an underlying StringCursor.
func (s *String) Cursor() Cursor {
    return s.StringCursor()
}

// Snapshot returns the structure's representation as a string.
func (s *String) Snapshot() string { ... }

// Len returns the String's number of codepoints.
func (s *String) Len() int { ... }

// StringCursor is a mutable tree location.
type StringCursor struct { treePosition }

// StringCursor implements Cursor, with generic mutating operations.
func (cur *StringCursor) Index(i int) { ... }
func (cur *StringCursor) Delete()     { ... }

// Value retrieves the Char at the cursor's position.
func (cur *StringCursor) Value() *Char {
    return &Char{cur.tree, cur.atomID, cur.lastKnownPos}
}

// Insert inserts a character after the cursor, and moves the cursor to the inserted char.
func (cur *StringCursor) Insert(ch rune) { ... }


// Char represents a Unicode codepoint within a String.
// It is not a Value because it can't be set within registers, always existing within a
// String.
type Char struct { treePosition }

func (ch *Char) Snapshot() rune { ... }
```

### `Counter`

```go
// Counter is a Value representing a mutable integer.
type Counter struct { treePosition }
func (*Counter) isValue() {}

// Snapshot returns the current value of the counter.
func (c *Counter) Snapshot() int32 { ... }

// Increment and decrement add mutations to the counter.
func (c *Counter) Increment(x int32) { ... }
func (c *Counter) Decrement(x int32) { ... }
```

### `List`

```go
// List and ListCursor are just pointers to a location within the list.
type List struct { treePosition }
func (*List) isValue() {}

// ListCursor returns a cursor to the list's initial position.
func (l *List) ListCursor() *ListCursor {
    return &ListCursor{s.tree, s.atomID, s.lastKnownPos}
}

// Cursor returns an underlying ListCursor.
func (l *List) Cursor() Cursor {
    return s.ListCursor()
}

// Snapshot returns the structure's representation as a slice.
func (l *List) Snapshot() []interface{} { ... }

// Len returns the number of elements in the list.
func (l *List) Len() int { ... }

// ListCursor is a mutable tree location.
type ListCursor struct { treePosition }

// ListCursor implements Cursor, with generic mutating operations.
func (cur *ListCursor) Index(i int) { ... }
func (cur *ListCursor) Delete()     { ... }

// Value retrieves the Element at the cursor's position.
func (cur *ListCursor) Value() *Element {
    return &Element{cur.tree, cur.atomID, cur.lastKnownPos}
}

// Insert inserts an empty element after the cursor, and moves the cursor to the inserted position.
func (cur *ListCursor) Insert() *Element { ... }


// Element represents a position within a List.
// It is not a Value because it can't be set within registers, always existing within a
// List.
type Element struct { treePosition }

// Element can be set and reset, implementing a Register.
func (*Element) SetString() *String   { ... }
func (*Element) SetCounter() *Counter { ... }
func (*Element) SetList() *List       { ... }
func (*Element) Clear()               { ... }
func (*Element) Value() Value         { ... }
```

