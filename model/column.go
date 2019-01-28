package model

// Column represent a directory
type Column interface {
	File() FileItem
	Path() string
	Update()
	Refresh(FileItem) error
	MarkedOrSelected() []FileItem
	ToggleMarkAll()

	FileList
	Filterer
	Marker
	Selector
}

// BaseColumn is a column in local file systems
type BaseColumn struct {
	FileList
	Selector
	Marker
	Filterer
}

// Update column state
func (bc *BaseColumn) Update() {
	bc.DoFilter()
	bc.SelectFirst()
	bc.ClearMark()
	bc.Sort(bc.Order())
}

// MarkedOrSelected if have marks return marked else return selected
func (bc *BaseColumn) MarkedOrSelected() []FileItem {
	mks := bc.Marked()
	if len(mks) == 0 {
		file, err := bc.CurrentFile()
		if err == nil {
			mks = append(mks, file)
		}
	}
	return mks
}

// ToggleMarkAll if have marked files clear them, else mark all
func (bc *BaseColumn) ToggleMarkAll() {
	mks := bc.Marked()
	if len(mks) == 0 {
		for i := range bc.Files() {
			bc.Mark(i)
		}
		return
	}

	bc.ClearMark()
}

// LocalColumn use local file system
type LocalColumn struct {
	item FileItem
	*BaseColumn
}

// File the underground file
func (bc *LocalColumn) File() FileItem {
	return bc.item
}

// Path the underground path of column
func (bc *LocalColumn) Path() string {
	return bc.item.Path()
}

// Refresh the specified path
func (bc *LocalColumn) Refresh(item FileItem) error {
	if item != nil {
		err := bc.item.(DirOp).Close()
		if err != nil {
			return err
		}
		bc.item = item
	}

	items, err := bc.item.(DirOp).Read()
	if err != nil {
		return err
	}

	fl := &BaseFileList{bc.Order(), items, items, false}
	se := &BaseSelector{0, fl}
	ma := &BaseMarker{nil, se}
	fi := &BaseFilter{bc.Filter(), bc.IsShowHidden(), fl}

	bc.FileList = fl
	bc.Selector = se
	bc.Marker = ma
	bc.Filterer = fi
	bc.Update()
	return nil
}

// NewLocalColumn create column
func NewLocalColumn(item FileItem) (Column, error) {
	items, err := item.(DirOp).Read()
	if err != nil {
		return nil, err
	}
	fl := &BaseFileList{OrderByName, items, items, false}
	se := &BaseSelector{0, fl}
	ma := &BaseMarker{nil, se}
	fi := &BaseFilter{"", false, fl}
	fi.DoFilter()
	fl.Sort(OrderByName)
	return &LocalColumn{item, &BaseColumn{fl, se, ma, fi}}, nil
}
