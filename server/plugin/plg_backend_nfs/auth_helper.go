package plg_backend_nfs

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	. "github.com/mickael-kerjean/filestash/server/common"
)

var (
	cacheForEtc   AppCache
	cacheForGroup AppCache
)

func init() {
	cacheForEtc = NewAppCache(120, 60)
	cacheForGroup = NewAppCache(120, 60)
}

func extractUserInfo(uidHint string, gidHint string, gidsHint string) (uint32, uint32, []groupLabel) {
	// case 1: everything is being sent as "uid=number, gid=number and gids=number,number,number"
	if _uid, err := strconv.Atoi(uidHint); err == nil {
		var (
			uid  uint32 = uint32(_uid)
			gid  uint32
			gids []groupLabel
		)
		if _gid, err := strconv.Atoi(gidHint); err == nil {
			gid = uint32(_gid)
		} else {
			gid = uid
		}
		for _, t := range strings.Split(gidsHint, ",") {
			tmp := strings.TrimSpace(t)
			if gid, err := strconv.Atoi(tmp); err == nil {
				gids = append(gids, groupLabel{uint32(gid), tmp, 0})
			}
		}
		return uid, gid, gids
	}
	// case 2: auto detect everything, aka "uid=www-data gid=www-data gids=..." based on uid=www-data
	if _uid, _gid, err := extractFromEtcPasswd(uidHint); err == nil {
		return _uid, _gid, extractFromEtcGroup(uidHint, _gid)
	}
	// case 3: base case
	return 0, 0, []groupLabel{}
}

func extractFromEtcPasswd(username string) (uint32, uint32, error) {
	if v := cacheForEtc.Get(map[string]string{"username": username}); v != nil {
		inCache := v.([]int)
		return uint32(inCache[0]), uint32(inCache[1]), nil
	}
	f, err := os.OpenFile("/etc/passwd", os.O_RDONLY, os.ModePerm)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	lines := bufio.NewReader(f)
	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}
		s := strings.Split(string(line), ":")
		if len(s) != 7 {
			continue
		} else if username != s[0] {
			continue
		}
		u, err := strconv.Atoi(s[2])
		if err != nil {
			continue
		}
		g, err := strconv.Atoi(s[3])
		if err != nil {
			continue
		}
		cacheForEtc.Set(map[string]string{"username": username}, []int{u, g})
		return uint32(u), uint32(g), nil
	}
	return 0, 0, ErrNotFound
}

type groupLabel struct {
	id       uint32
	label    string
	priority int
}

func extractFromEtcGroup(username string, primary uint32) []groupLabel {
	if v := cacheForGroup.Get(map[string]string{"username": username}); v != nil {
		return v.([]groupLabel)
	}
	f, err := os.OpenFile("/etc/group", os.O_RDONLY, os.ModePerm)
	if err != nil {
		return []groupLabel{}
	}
	defer f.Close()
	gids := []groupLabel{}
	lines := bufio.NewReader(f)
	for {
		line, _, err := lines.ReadLine()
		if err != nil {
			break
		}
		s := strings.Split(string(line), ":")
		if len(s) != 4 {
			continue
		}
		userInGroup := false
		for _, user := range strings.Split(s[3], ",") {
			if user == username {
				userInGroup = true
				break
			}
		}
		if userInGroup == false {
			continue
		}
		if gid, err := strconv.Atoi(s[2]); err == nil {
			ugid := uint32(gid)
			if ugid != primary {
				gids = append(gids, groupLabel{ugid, s[0], 0})
			}
		}
		cacheForGroup.Set(map[string]string{"username": username}, gids)
	}
	return gids
}
