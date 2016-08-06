package redislibs

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/laincloud/redis-libs/network"
	"net"
	"strconv"
	"strings"
	"time"
)

type Address struct {
	Host string
	Port string
}

func NewAddress(adr string) (*Address, error) {
	addr := new(Address)
	hostport := strings.Split(adr, ":")
	if len(hostport) != 2 {
		return nil, errors.New("address must format as host:port")
	}
	addr.Host = hostport[0]
	if addr.Host == "" {
		addr.Host = "127.0.0.1"
	}
	addr.Port = hostport[1]
	return addr, nil
}

func BuildAddress(host, port string) *Address {
	addr := new(Address)
	addr.Host = host
	addr.Port = port
	return addr
}

func (addr *Address) String() string {
	return addr.Host + ":" + addr.Port
}

type ITalker interface {
	Connect()
	TalkRaw(commands string) string
	Close()
}

type Talker struct {
	*network.Conn
	br *bufio.Reader
}

func (t *Talker) Close() {
	if t == nil {
		return
	}
	t.Conn.Close()
}

func BuildTalker(h, p string) (*Talker, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(h, p), 5*time.Second)
	if err != nil {
		return nil, err
	}
	co := network.NewConnectOption(5, 5, 1024)
	cn, _ := network.NewConnect(conn, co)
	t := &Talker{Conn: cn}
	if conn != nil {
		t.br = bufio.NewReader(t.GetConn())
	}
	return t, nil
}

type ClusterNode struct {
	Host            string
	Port            string
	Node_id         string
	Role_in_cluster string
	Masterid        string
	Assigned_slots  []int
	migrating_slots []int
	info            []string
}

func (node *ClusterNode) String() string {
	res := ""
	for _, v := range node.info {
		res += v + " "
	}
	res += "\n"
	return res
}

func NewClusterNode(args []string) (*ClusterNode, error) {
	//<id> <ip:Port> <flags> <master> <ping-sent> <pong-recv> <config-epoch> <link-state> <slot> <slot> ... <slot>
	// last_ping_sent_time := args[4]
	// last_pong_received_time := args[5]
	// config_epoch := args[6]
	// link_status := args[7]
	cnode := new(ClusterNode)
	cnode.info = args
	if len(args) < 8 {
		return nil, nil
	}
	Node_id := args[0]
	latest_know_ip_address_and_Port := args[1]
	Role_in_cluster := args[2]
	Node_id_of_master_if_it_is_a_slave := args[3]

	Assigned_slots := args[8:]

	cnode.Node_id = Node_id
	addrs := strings.Split(latest_know_ip_address_and_Port, NETADDRSPLITSMB)
	cnode.Host = addrs[0]
	cnode.Port = addrs[1]
	cnode.Role_in_cluster = Role_in_cluster
	if Node_id_of_master_if_it_is_a_slave != "-" {
		cnode.Masterid = Node_id_of_master_if_it_is_a_slave
	}
	cnode.Assigned_slots = make([]int, 0, 10)
	cnode.migrating_slots = make([]int, 0, 0)
	for _, v := range Assigned_slots {
		if strings.HasPrefix(v, "[") {
			idx := strings.Index(v, SYM_MINUS)
			slotInMigrating, err := strconv.Atoi(v[1:idx])
			if err != nil {
				fmt.Printf("parse err %s\n", err)
			}
			cnode.migrating_slots = append(cnode.migrating_slots, slotInMigrating)
			continue
		}
		nums := strings.Split(v, SYM_MINUS)
		if len(nums) > 1 {
			start, err := strconv.Atoi(nums[0])
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(nums[1])
			if err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				cnode.Assigned_slots = append(cnode.Assigned_slots, i)
			}
		} else {
			num, err := strconv.Atoi(nums[0])
			cnode.Assigned_slots = append(cnode.Assigned_slots, num)
			if err != nil {
				return nil, err
			}
		}
	}
	return cnode, nil
}
