package file

import (
	"os/user"
	"strconv"
	"syscall"
)

var (
	uidMap = make(map[uint32]string)
	gidMap = make(map[uint32]string)
)

func gid2name(sys interface{}) string {
	if sys == nil {
		return ""
	}

	st, ok := sys.(*syscall.Stat_t)
	if !ok {
		return ""
	}

	gn, ok := gidMap[st.Gid]
	if ok {
		return gn
	}

	g, err := user.LookupGroupId(strconv.Itoa(int(st.Gid)))
	if err != nil {
		return ""
	}

	gidMap[st.Gid] = g.Name
	return g.Name
}

func uid2name(sys interface{}) string {
	if sys == nil {
		return ""
	}

	st, ok := sys.(*syscall.Stat_t)
	if !ok {
		return ""
	}

	un, ok := uidMap[st.Uid]
	if ok {
		return un
	}

	u, err := user.LookupId(strconv.Itoa(int(st.Uid)))
	if err != nil {
		return ""
	}

	uidMap[st.Gid] = u.Username
	return u.Username
}
