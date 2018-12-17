package model

// Column represent a directory
type Column interface {
	Path() string
	Update()
	Refresh(string, []FileItem)

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

// LocalColumn use local file system
type LocalColumn struct {
	path string
	*BaseColumn
}

// Path the underground path of column
func (bc *LocalColumn) Path() string {
	return bc.path
}

// Refresh the specified path
func (bc *LocalColumn) Refresh(path string, items []FileItem) {
	bc.path = path

	fl := &BaseFileList{bc.Order(), items, items, false}
	se := &BaseSelector{0, fl}
	ma := &BaseMarker{nil, se}
	fi := &BaseFilter{bc.Filter(), bc.IsShowHidden(), fl}

	bc.FileList = fl
	bc.Selector = se
	bc.Marker = ma
	bc.Filterer = fi
	bc.Update()
}

// NewLocalColumn create column
func NewLocalColumn(path string, items []FileItem) Column {
	fl := &BaseFileList{OrderByName, items, items, false}
	se := &BaseSelector{0, fl}
	ma := &BaseMarker{nil, se}
	fi := &BaseFilter{"", false, fl}
	fi.DoFilter()
	fl.Sort(OrderByName)
	return &LocalColumn{path, &BaseColumn{fl, se, ma, fi}}
}
