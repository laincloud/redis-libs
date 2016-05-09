package redislibs

import (
	"errors"
	"github.com/mijia/sweb/log"
	"strconv"
	"strings"
)

const (
	CONNECTED_SLAVES = "connected_slaves"
	ROLE             = "role"
	HOST             = "ip"
	PORT             = "port"
)

func SlaveOf(sHost, sPort, mHost, mPort string) (string, error) {
	t, err := buildTalker(sHost, sPort)
	defer t.Close()
	if err != nil {
		return "", err
	}
	return slaveOf(t, mHost, mPort)
}

func Slaves(mHost, mPort string) ([]*Address, error) {
	t, err := buildTalker(mHost, mPort)
	defer t.Close()
	if err != nil {
		return nil, err
	}
	return slaves(t)
}

func Role(Host, Port string) (string, error) {
	t, err := buildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	r, _, err := role_status(t)
	return r, err
}

func RoleStatus(Host, Port string) (string, string, error) {
	t, err := buildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", "", err
	}
	return role_status(t)
}

// 1. slave no one
// 2. sentinel reset *(master name in sentinel)
func RemoveSlaveFromSentinel(sHost, sPort, master string, sentinel_addrs ...*Address) error {
	t, err := buildTalker(sHost, sPort)
	defer t.Close()
	if err != nil {
		return err
	}
	m, err := slaveOf(t, "NO", "ONE")
	if err != nil {
		return err
	}
	if m != STATUS_OK {
		return errors.New(m)
	}
	for _, addr := range sentinel_addrs {
		t, err := buildTalker(addr.Host, addr.Port)
		defer t.Close()
		if err != nil {
			return err
		}
		sentinelReset(t, master)
	}
	return nil
}

func slaves(t *Talker) ([]*Address, error) {
	resp, err := t.Talk(Pack_command("INFO", "Replication"))
	if err != nil {
		return nil, err
	}
	result := UnPackResponse(resp)
	infos := make(map[string]string)
	for _, r := range result {
		kv := strings.Split(r, ":")
		if len(kv) != 2 {
			log.Info("r:", r)
			continue
		}
		infos[kv[0]] = kv[1]
	}
	if infos[ROLE] != ROLE_MASTER {
		return nil, errors.New(STATUS_ERR + " Node Must Be Master")
	}
	addres := make([]*Address, 0, 0)
	slave_count, err := strconv.Atoi(infos[CONNECTED_SLAVES])
	if err != nil {
		return nil, err
	}
	var host, port, info string
	for i := 0; i < slave_count; i++ {
		info = infos["slave"+strconv.Itoa(i)] //slave0 slave1 ... in info Replication
		sub_infos := strings.Split(info, ",")
		for _, sub_info := range sub_infos {
			addr_status := strings.Split(sub_info, "=")
			if addr_status[0] == HOST {
				host = addr_status[1]
			} else if addr_status[0] == PORT {
				port = addr_status[1]
			}
		}
		addres = append(addres, BuildAddress(host, port))
	}
	return addres, nil
}

func slaveOf(t *Talker, mHost, mPort string) (string, error) {
	return t.Talk(Pack_command("slaveof", mHost, mPort))
}

func role_status(t *Talker) (string, string, error) {
	resp, err := t.TalkForObject(Pack_command("role")) // role reback with array
	if err != nil {
		return "", "", err
	}
	infos := resp.([]interface{})
	if len(infos) == 0 {
		return "", "", errors.New("Fatal Error")
	}
	role, _ := infos[0].(string)
	if ROLE_MASTER == role || ROLE_SENTINEL == role {
		return role, "", nil
	}
	status, _ := infos[3].(string)
	return role, status, nil

}
