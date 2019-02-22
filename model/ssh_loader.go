package model

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	_ = Loader(new(sshLoader))
)

var (
	sshCache = make(map[string]*sshc)
)

type sshconfig struct {
	host, port           string
	user, key            string
	editor, pager, shell string
	timeout              time.Duration
}

type sshc struct {
	config *sshconfig
	conn   *ssh.Client
	os     string
	origin FileItem
	root   FileItem
	loader Loader
	tmpDir string
	cache  map[string]*sshfileCache
}

type sshfileCache struct {
	path    string
	modTime time.Time
}

func (ss *sshc) error(msg string) error {
	return fmt.Errorf("%s: %s", ss.loader.Name(), msg)
}

func (ss *sshc) term() (*ssh.Session, func(), error) {
	se, err := ss.conn.NewSession()
	if err != nil {
		return nil, nil, err
	}

	se.Stderr = os.Stderr
	se.Stdin = os.Stdin
	se.Stdout = os.Stdout

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 115200,
		ssh.TTY_OP_OSPEED: 115200,
	}

	fd := int(os.Stdin.Fd())
	w, h, err := terminal.GetSize(fd)
	if err != nil {
		se.Close()
		return nil, nil, err
	}
	ts, err := terminal.MakeRaw(fd)
	if err != nil {
		se.Close()
		return nil, nil, err
	}
	se.RequestPty("xterm-256color", h, w, modes)

	return se, func() { terminal.Restore(fd, ts) }, nil
}

func (ss *sshc) execf(format string, args ...interface{}) (*bytes.Buffer, error) {
	return ss.exec(fmt.Sprintf(format, args...))
}

func (ss *sshc) exec(cmd string) (*bytes.Buffer, error) {
	session, err := ss.conn.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	err = session.Run(cmd)
	return &out, err
}

type sshLoader struct{}

func (*sshLoader) Name() string               { return "ssh" }
func (*sshLoader) Seperator() string          { return "/" }
func (*sshLoader) Support(item FileItem) bool { return strings.HasSuffix(item.Name(), ".ssh.fff") }
func (sl *sshLoader) Create(origin FileItem) (FileItem, error) {
	sc, err := sl.loadConfig(origin)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s@%s", sc.user, sc.host)
	ssc, ok := sshCache[key]
	if !ok {
		conn, err := sl.login(sc)
		if err != nil {
			return nil, err
		}

		td, err := ioutil.TempDir("", "fff")
		if err != nil {
			return nil, err
		}
		ssc = &sshc{sc, conn, "", origin, nil, sl, td, make(map[string]*sshfileCache)}
		buf, err := ssc.exec("uname")
		if err != nil {
			return nil, err
		}
		ssc.os = strings.Trim(buf.String(), " \n")

		sfi := &sshFileItem{"/", "root", "root", time.Now(), 0, 0755, true, nil}
		root := &sshroot{&sshdir{&sshItem{ssc, "/", sfi}}}
		ssc.root = root
		sshCache[key] = ssc
	}

	return ssc.root, nil
}

func (*sshLoader) login(sc *sshconfig) (*ssh.Client, error) {
	cfg := &ssh.ClientConfig{
		User:    sc.user,
		Timeout: sc.timeout,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	host := fmt.Sprintf("%s:%s", sc.host, sc.port)

	if sc.key != "" {
		buf, err := ioutil.ReadFile(sc.key)
		if err != nil {
			return nil, err
		}

		key, err := ssh.ParsePrivateKey(buf)
		if err != nil {
			return nil, err
		}

		cfg.Auth = []ssh.AuthMethod{ssh.PublicKeys(key)}
		return ssh.Dial("tcp", host, cfg)
	}

	if ag, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		cfg.Auth = []ssh.AuthMethod{ssh.PublicKeysCallback(agent.NewClient(ag).Signers)}
		conn, err := ssh.Dial("tcp", host, cfg)
		if err == nil {
			return conn, nil
		}
	}

	RequestCh <- &Request{fmt.Sprintf("Enter password for %s@%s", sc.user, sc.host), true}
	pw := <-ResponseCh
	if pw != "" {
		cfg.Auth = []ssh.AuthMethod{ssh.Password(pw)}
		conn, err := ssh.Dial("tcp", host, cfg)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	return nil, errors.New("login fail")
}

func (*sshLoader) loadConfig(file FileItem) (*sshconfig, error) {
	re, err := file.(FileOp).Reader()
	if err != nil {
		return nil, err
	}
	defer re.Close()

	bt, err := ioutil.ReadAll(re)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bt), "\n")
	sc := new(sshconfig)
	for _, v := range lines {
		if len(v) == 0 {
			continue
		}
		ts := strings.Split(v, ":")
		if len(ts) != 2 {
			return nil, fmt.Errorf("%s: illegle config line: %s", file.Path(), v)
		}
		name := strings.Trim(ts[0], " ")
		value := strings.Trim(ts[1], " ")
		switch name {
		case "host":
			sc.host = value
		case "port":
			sc.port = value
		case "user":
			sc.user = value
		case "key":
			sc.key = value
		case "editor":
			sc.editor = value
		case "pager":
			sc.pager = value
		case "shell":
			sc.shell = value
		case "timeout":
			ti, err := time.ParseDuration(value)
			if err != nil {
				return nil, err
			}
			sc.timeout = ti
		}
	}

	if sc.user == "" {
		sc.user = "root"
	}
	if sc.port == "" {
		sc.port = "22"
	}
	if sc.editor == "" {
		sc.editor = "vi"
	}
	if sc.pager == "" {
		sc.pager = "less"
	}
	if sc.shell == "" {
		sc.shell = "bash"
	}
	if sc.timeout == 0 {
		sc.timeout = 3 * time.Second
	}
	return sc, nil
}
