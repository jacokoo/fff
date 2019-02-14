package model

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

/*
S_IFMT     0170000   bit mask for the file type bit fields
S_IFSOCK   0140000   socket
S_IFLNK    0120000   symbolic link
S_IFREG    0100000   regular file
S_IFBLK    0060000   block device
S_IFDIR    0040000   directory
S_IFCHR    0020000   character device
S_IFIFO    0010000   FIFO
S_ISUID    0004000   set UID bit
S_ISGID    0002000   set-group-ID bit (see below)
S_ISVTX    0001000   sticky bit (see below)
S_IRWXU    00700     mask for file owner permissions
S_IRUSR    00400     owner has read permission
S_IWUSR    00200     owner has write permission
S_IXUSR    00100     owner has execute permission
S_IRWXG    00070     mask for group permissions
S_IRGRP    00040     group has read permission
S_IWGRP    00020     group has write permission
S_IXGRP    00010     group has execute permission
S_IRWXO    00007     mask for permissions for others (not in group)
S_IROTH    00004     others have read permission
S_IWOTH    00002     others have write permission
S_IXOTH    00001     others have execute permission
*/
const (
	rawDir  = 0040000
	rawLink = 0120000
)

var (
	_ = Op(new(sshItem))
	_ = FileOp(new(sshfile))
	_ = DirOp(new(sshdir))
)

var (
	readMap = map[string]func(*sshc, string, bool) (io.Reader, error){
		// stat -c "%G // %U // %f // %Y // %s // %N" .* *
		// root // root // 4168 // 1548990154 // 4096 // '.'
		// root // root // a1ff // 1548989494 // 12 // 'systemd' -> '/etc/systemd'
		"Linux": func(sc *sshc, path string, dir bool) (io.Reader, error) {
			if !dir {
				return sc.exec(`stat -c "%G // %U // %f // %Y // %s // %N" ` + path)
			}
			return sc.exec("cd " + path + `; stat -c "%G // %U // %f // %Y // %s // %N" .* *`)
		},
		// stat -f "%Sg // %Su // %Xp // %m // %z // '%N' -> '%Y'" .* *
		// wheel // root // 41ed // 1532542394 // 960 // '.' -> ''
		// wheel // root // a1ed // 1512168297 // 11 // 'var' -> 'private/var'
		"Darwin": func(sc *sshc, path string, dir bool) (io.Reader, error) {
			if !dir {
				return sc.exec(`stat -f "%Sg // %Su // %Xp // %m // %z // ’%N’ -> ’%Y’" ` + path)
			}
			return sc.exec("cd " + path + `; stat -f "%Sg // %Su // %Xp // %m // %z // ’%N’ -> ’%Y’" .* *`)
		},
	}
)

type sshFileItem struct {
	name, group, user string
	mtime             time.Time
	size              int64
	mode              os.FileMode
	dir               bool
	link              *fileLink
}

func (sf *sshFileItem) IsDir() bool        { return sf.dir }
func (sf *sshFileItem) ModTime() time.Time { return sf.mtime }
func (sf *sshFileItem) Mode() os.FileMode  { return sf.mode }
func (sf *sshFileItem) Name() string       { return sf.name }
func (sf *sshFileItem) Size() int64        { return sf.size }
func (sf *sshFileItem) Sys() interface{}   { return nil }
func (sf *sshFileItem) Link() (Link, bool) { return sf.link, sf.link != nil }

func (sc *sshc) readDir(pp string) ([]FileItem, error) {
	fn, ok := readMap[sc.os]
	if !ok {
		return nil, sc.error("target os is not supported")
	}
	buf, _ := fn(sc, pp, true)
	scan := bufio.NewScanner(buf)
	its := make([]FileItem, 0)
	for scan.Scan() {
		file, err := parseSSHFile(scan.Text())
		if err != nil {
			continue
		}
		if file.name == "." || file.name == ".." {
			continue
		}
		si := &sshItem{sc, path.Join(pp, file.name), file}
		if file.IsDir() {
			its = append(its, &sshdir{si})
		} else {
			its = append(its, &sshfile{si})
		}
	}
	return its, nil
}

func (sc *sshc) readFile(pp string) (FileItem, error) {
	fn, ok := readMap[sc.os]
	if !ok {
		return nil, sc.error("target os is not supported")
	}
	buf, _ := fn(sc, pp, false)
	scan := bufio.NewScanner(buf)
	its := make([]FileItem, 0)
	for scan.Scan() {
		file, err := parseSSHFile(scan.Text())
		if err != nil {
			continue
		}
		si := &sshItem{sc, file.name, file}
		file.name = path.Base(file.name)
		if file.IsDir() {
			its = append(its, &sshdir{si})
		} else {
			its = append(its, &sshfile{si})
		}
	}
	if len(its) != 1 {
		return nil, sc.error("no such file: " + pp)
	}
	return its[0], nil
}

func parseSSHFile(str string) (*sshFileItem, error) {
	ts := strings.Split(str, " // ")
	if len(ts) != 6 {
		return nil, errors.New("incorrect file string")
	}
	file := new(sshFileItem)

	// group, user
	file.group = ts[0]
	file.user = ts[1]

	// mode
	s, err := strconv.ParseInt(ts[2], 16, 32)
	if err != nil {
		return nil, err
	}
	file.dir = false
	perm := os.FileMode(s & 0777)
	s -= int64(perm)
	if s&rawDir == rawDir {
		perm = perm | os.ModeDir
		file.dir = true
	} else if s&rawLink == rawLink {
		perm = perm | os.ModeSymlink
		file.link = &fileLink{false, "", false}
	}
	file.mode = perm

	// mtime
	itime, err := strconv.ParseInt(ts[3], 10, 64)
	if err != nil {
		return nil, err
	}
	file.mtime = time.Unix(itime, 0)

	// size
	size, err := strconv.ParseInt(ts[4], 10, 64)
	if err != nil {
		return nil, err
	}
	file.size = size

	// name
	ns := strings.Split(ts[5], " -> ")
	file.name = strings.TrimRight(strings.Trim(ns[0], "‘’ "), "/")

	// link
	if file.link != nil {
		file.link.target = strings.Trim(ns[1], "' ")
	}

	return file, nil
}

type sshItem struct {
	sshc  *sshc
	ipath string
	*sshFileItem
}

func (so *sshItem) Path() string {
	return so.sshc.origin.Path() + LoaderString(so.sshc.loader) + so.ipath
}

func (so *sshItem) Dir() (FileItem, error) {
	pp := path.Dir(so.ipath)
	if pp == so.ipath {
		return so.sshc.origin.(Op).Dir()
	}
	if pp == "/" {
		return so.sshc.root, nil
	}
	return so.sshc.root.(DirOp).To(pp)
}

func (so *sshItem) Open() error { return so.sshc.origin.(Op).Open() }
func (so *sshItem) Delete() error {
	_, err := so.sshc.execf("rm -rf %s", so.ipath)
	return err
}
func (so *sshItem) Rename(name string) error {
	pp := path.Dir(so.ipath)
	np := path.Join(pp, name)
	_, err := so.sshc.execf("mv %s %s", so.ipath, np)
	return err
}

type sshfile struct {
	*sshItem
}

func (sf *sshfile) readerFromCache() (io.ReadCloser, error) {
	cc, ok := sf.sshc.cache[sf.ipath]
	if !ok {
		return nil, errors.New("no cache")
	}
	if cc.modTime != sf.ModTime() {
		return nil, errors.New("file changed")
	}

	return os.Open(cc.path)
}

func (sf *sshfile) Reader() (io.ReadCloser, error) {
	rc, err := sf.readerFromCache()
	if err == nil {
		return rc, nil
	}

	tmp, err := ioutil.TempFile(sf.sshc.tmpDir, "*")
	if err != nil {
		return nil, err
	}
	sf.sshc.cache[sf.ipath] = &sshfileCache{tmp.Name(), sf.ModTime()}

	session, err := sf.sshc.conn.NewSession()
	if err != nil {
		return nil, err
	}

	in, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, err
	}
	err = session.Start("dd if=" + sf.ipath)
	if err != nil {
		session.Close()
		return nil, err
	}

	return newReadCloser(io.TeeReader(in, tmp), session, tmp), nil
}

func (sf *sshfile) Writer(int) (io.WriteCloser, error) {
	session, err := sf.sshc.conn.NewSession()
	if err != nil {
		return nil, err
	}

	out, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, err
	}
	err = session.Start("dd of=" + sf.ipath)
	if err != nil {
		session.Close()
		return nil, err
	}

	return newWriteCloser(out, out, session), nil
}

func (sf *sshfile) View() error {
	se, fn, err := sf.sshc.term()
	if err != nil {
		return err
	}
	defer fn()
	defer se.Close()

	return se.Run(sf.sshc.config.pager + " " + sf.ipath)
}

func (sf *sshfile) Edit() error {
	se, fn, err := sf.sshc.term()
	if err != nil {
		return err
	}
	defer fn()
	defer se.Close()

	return se.Run(sf.sshc.config.editor + " " + sf.ipath)
}

type sshdir struct {
	*sshItem
}

func (sd *sshdir) Read() ([]FileItem, error) {
	return sd.sshc.readDir(sd.ipath)
}

func (sd *sshdir) Write([]FileItem) (Task, error) {
	return nil, nil
}

func (sd *sshdir) Move([]FileItem) error {
	return sd.sshc.error("move is not supported")
}
func (sd *sshdir) NewFile(name string) error {
	_, err := sd.sshc.execf("touch %s", path.Join(sd.ipath, name))
	return err
}
func (sd *sshdir) NewDir(name string) error {
	_, err := sd.sshc.execf("mkdir -p %s", path.Join(sd.ipath, name))
	return err
}
func (sd *sshdir) To(pp string) (FileItem, error) {
	p := path.Join(sd.ipath, pp)
	return sd.sshc.readFile(p)
}
func (sd *sshdir) Shell() error {
	se, fn, err := sd.sshc.term()
	if err != nil {
		return err
	}
	defer fn()
	defer se.Close()
	return se.Run(fmt.Sprintf("cd %s; %s", sd.ipath, sd.sshc.config.shell))
}

type sshroot struct {
	*sshdir
}
