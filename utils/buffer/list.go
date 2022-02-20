package buffer

const (
	unknownPosition = iota
	frontPosition
	middlePosition
	backPosition
)

type IList interface {
	Len() int
	Front() *listItem
	Back() *listItem
	PushFront(v interface{}) *listItem
	PushBack(v interface{}) *listItem
	Remove(i *listItem)
	MoveToFront(i *listItem)
}

type listItem struct {
	Value interface{}
	Next  *listItem
	Prev  *listItem
}

type list struct {
	len   int
	front *listItem
	back  *listItem
}

func newList() IList {
	return new(list)
}

func (l list) Len() int {
	return l.len
}

func (l *list) Front() *listItem {
	return l.front
}

func (l *list) Back() *listItem {
	return l.back
}

func (l *list) PushFront(v interface{}) *listItem {
	i := &listItem{}
	i.Value = v
	i.Prev = nil
	if l.len == 0 {
		i.Next = nil
		l.back = i
	} else {
		i.Next = l.front
		l.front.Prev = i
	}
	l.front = i
	l.len++
	return i
}

func (l *list) PushBack(v interface{}) *listItem {
	i := &listItem{}
	i.Value = v
	i.Next = nil
	if l.len == 0 {
		i.Prev = nil
		l.front = i
	} else {
		i.Prev = l.back
		l.back.Next = i
	}
	l.back = i
	l.len++
	return i
}

func (l *list) Remove(i *listItem) {
	if l.len != 0 && i != nil {
		switch getItemPosition(i) {
		case middlePosition:
			i.Next.Prev = i.Prev
			i.Prev.Next = i.Next
		case frontPosition:
			i.Next.Prev = nil
			l.front = i.Next
		case backPosition:
			i.Prev.Next = nil
			l.back = i.Prev
		default:
			return
		}
		i.Value = nil
		l.len--
	}
}

func (l *list) MoveToFront(i *listItem) {
	if l.len > 1 && i != nil {
		if i.Prev != nil {
			if i.Next != nil {
				i.Next.Prev = i.Prev
				i.Prev.Next = i.Next
			} else {
				i.Prev.Next = nil
				l.back = i.Prev
			}
			i.Prev = nil
			i.Next = l.front
			l.front.Prev = i
			l.front = i
		}
	}
}

func getItemPosition(i *listItem) int {
	if i.Next != nil && i.Prev != nil {
		return middlePosition
	}
	if i.Next != nil {
		return frontPosition
	}
	if i.Prev != nil {
		return backPosition
	}
	return unknownPosition
}
