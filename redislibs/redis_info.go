package redislibs

import "strings"

func RedisNodeInfo(host, port string) (map[string]map[string]string, error) {
	t, err := buildTalker(host, port)
	defer t.Close()
	if err != nil {
		return nil, err
	}
	return redisNodeInfo(t)
}

func redisNodeInfo(t *Talker) (map[string]map[string]string, error) {
	resp, err := t.Talk(Pack_command("info"))
	if err != nil {
		return nil, err
	}
	infos := strings.Split(resp, SYM_CRLF)
	res := make(map[string]map[string]string)
	var sub map[string]string
	for _, info := range infos {
		if info == "" || strings.HasPrefix(info, SYM_DOLLAR) || strings.HasPrefix(info, SYM_STAR) {
			continue
		}
		if strings.HasPrefix(info, "#") {
			sub = make(map[string]string)
			res[info[2:]] = sub
			continue
		}
		entry := strings.Split(info, ":")
		sub[entry[0]] = entry[1]
	}
	return res, nil
}
