package model

// Marker content can be marked
type Marker interface {
	Mark(idx int)
	Unmark(idx int)
	ToggleMark()
	Marked() []FileItem
	IsMarked(idx int) bool
	ClearMark()
}

// BaseMarker base marker
type BaseMarker struct {
	marks []int
	*BaseSelector
}

// ToggleMark toggle mark current file
func (bm *BaseMarker) ToggleMark() {
	if len(bm.files) == 0 {
		return
	}

	if bm.IsMarked(bm.current) {
		bm.Unmark(bm.current)
	} else {
		bm.Mark(bm.current)
	}
}

// Marked get marked files
func (bm *BaseMarker) Marked() []FileItem {
	if len(bm.files) == 0 || len(bm.marks) == 0 {
		return nil
	}

	re := make([]FileItem, len(bm.marks))
	for i, v := range bm.marks {
		re[i] = bm.files[v]
	}
	return re
}

// IsMarked determine if the idx is selected
func (bm *BaseMarker) IsMarked(idx int) bool {
	for _, i := range bm.marks {
		if i == idx {
			return true
		}
	}
	return false
}

// Unmark index
func (bm *BaseMarker) Unmark(index int) {
	ii := -1
	for idx, i := range bm.marks {
		if i == index {
			ii = idx
			break
		}
	}
	if ii != -1 {
		bm.marks = append(bm.marks[:ii], bm.marks[ii+1:]...)
	}
}

// Mark add mark
func (bm *BaseMarker) Mark(idx int) {
	if !bm.IsMarked(idx) {
		bm.marks = append(bm.marks, idx)
	}
}

// ClearMark clear all markes
func (bm *BaseMarker) ClearMark() {
	bm.marks = nil
}
