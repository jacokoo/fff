package model

// Marker content can be marked
type Marker interface {
	ToggleMark()
	Marked() []FileItem
	IsMarked(idx int) bool
	ClearMark()
}

// BaseMarker base marker
type BaseMarker struct {
	markes []int
	*BaseSelector
}

// ToggleMark toggle mark current file
func (bm *BaseMarker) ToggleMark() {
	if len(bm.files) == 0 {
		return
	}

	ii := -1
	for idx, i := range bm.markes {
		if i == bm.current {
			ii = idx
			break
		}
	}
	if ii == -1 {
		bm.markes = append(bm.markes, bm.current)
		return
	}

	bm.markes = append(bm.markes[:ii], bm.markes[ii+1:]...)
}

// Marked get marked files
func (bm *BaseMarker) Marked() []FileItem {
	if len(bm.files) == 0 {
		return nil
	}

	if len(bm.markes) == 0 {
		file, err := bm.CurrentFile()
		if err != nil {
			return nil
		}
		return []FileItem{file}
	}

	re := make([]FileItem, len(bm.markes))
	for i, v := range bm.markes {
		re[i] = bm.files[v]
	}
	return re
}

// IsMarked determine if the idx is selected
func (bm *BaseMarker) IsMarked(idx int) bool {
	for _, i := range bm.markes {
		if i == idx {
			return true
		}
	}
	return false
}

// ClearMark clear all markes
func (bm *BaseMarker) ClearMark() {
	bm.markes = nil
}
