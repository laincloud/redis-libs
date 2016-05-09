package redislibs

type SentinelMaster struct {
	Name string
	Host string
	Port string
}

func GetSlavesInSentinel(sentinel_host, sentinel_port, master string) ([]*Address, error) {
	t, err := buildTalker(sentinel_host, sentinel_port)
	defer t.Close()
	if err != nil {
		return nil, err
	}
	return getSlavesInSentinel(t, master)
}

func GetMasterAddrByName(sentinel_host, sentinel_port, master string) (*Address, error) {
	t, err := buildTalker(sentinel_host, sentinel_port)
	defer t.Close()
	if err != nil {
		return nil, err
	}
	return getMasterAddrByName(t, master)
}

func MonitorSentinel(sentinel_host, sentinel_port, mHost, mPort, master, quorum string) (string, error) {
	t, err := buildTalker(sentinel_host, sentinel_port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	return monitorSentinel(t, mHost, mPort, master, quorum)
}

func FetchMastersInSentinel(sentinel_host, sentinel_port string) ([]*SentinelMaster, error) {
	t, err := buildTalker(sentinel_host, sentinel_port)
	defer t.Close()
	if err != nil {
		return nil, err
	}
	return fetchMastersInSentinel(t)
}

func ConfigSentinel(sentinel_host, sentinel_port string, configs ...string) (string, error) {
	t, err := buildTalker(sentinel_host, sentinel_port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	return configSentinel(t, configs...)
}

func monitorSentinel(t *Talker, mHost, mPort, master, quorum string) (string, error) {
	return t.Talk(Pack_command("SENTINEL", "MONITOR", master, mHost, mPort, quorum))
}

func getSlavesInSentinel(t *Talker, master string) ([]*Address, error) {
	resp, err := t.TalkForObject(Pack_command("SENTINEL", "slaves", master))
	if err != nil {
		return nil, err
	}
	slaves_statuses, _ := resp.([]interface{})
	slaves := make([]*Address, 0, 0)
	for _, slave := range slaves_statuses {
		slaves_info, _ := slave.([]interface{})
		s_addr, _ := slaves_info[1].(string)
		if addr, err := NewAddress(s_addr); err == nil {
			slaves = append(slaves, addr)
		}
	}
	return slaves, nil
}

func getMasterAddrByName(t *Talker, master string) (*Address, error) {
	resp, err := t.TalkForObject(Pack_command("SENTINEL", "get-master-addr-by-name", master))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	result := resp.([]interface{})
	addr := BuildAddress(result[0].(string), result[1].(string))
	return addr, nil
}

func sentinelReset(t *Talker, master string) (string, error) {
	return t.Talk(Pack_command("SENTINEL", "reset", master))
}

func configSentinel(t *Talker, configs ...string) (string, error) {
	params := []string{"SENTINEL", "set"}
	params = append(params, configs...)
	return t.Talk(Pack_command(params...))
}

func fetchMastersInSentinel(t *Talker) ([]*SentinelMaster, error) {
	resp, err := t.TalkForObject(Pack_command("sentinel", "masters"))
	if err != nil {
		return nil, err
	}

	masters := resp.([]interface{})
	masters_len := len(masters)
	sentinel_masters_ := make([]*SentinelMaster, masters_len, masters_len)

	for i, master := range masters {
		master_array_ := master.([]interface{})
		sentinel_masters_[i] = &SentinelMaster{
			Name: master_array_[1].(string),
			Host: master_array_[3].(string),
			Port: master_array_[5].(string),
		}
	}
	return sentinel_masters_, nil
}
