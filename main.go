package main

import (
	"flag"
	"fmt"
	. "github.com/laincloud/redis-libs/redislibs"
	"strconv"
)

const (
	CMD_LIST              = "list"
	CMD_GETINFO           = "info"
	CMD_STARTSINGLE       = "initsingle"
	CMD_DESTROY           = "destroy"
	CMD_QUITEMPTY         = "quite"
	CMD_QUITBALANCE       = "quitbalance"
	CMD_FIXSTATESINGLE    = "fixstate"
	CMD_FIXSTATECLUSTR    = "fixclusterstate"
	CMD_JOINEMPTY         = "join"
	CMD_REPLICATE         = "replicate"
	CMD_JOINBALANCE       = "joinbalance"
	CMD_MIGRATEALLCROSS   = "migrateallout"
	CMD_FIXSLOTS          = "fixslots"
	CMD_SARTMULTIPLE      = "initmultiple"
	CMD_ADDSLOTS          = "addslots"
	CMD_DELSLOTS          = "deleteslots"
	CMD_MIGRATESLOTSIN    = "migratein"
	CMD_MIGRATESLOTSCROSS = "migrateout"

	FUNC_LISTNODES      = 1 << 0
	FUNC_GETINFO        = 1 << 1
	FUNC_STARTSINGLE    = 1 << 2
	FUNC_DESTROY        = 1 << 3
	FUNC_QUITEMPTY      = 1 << 4
	FUNC_QUITBALANCE    = 1 << 5
	FUNC_FIXSTATESINGLE = 1 << 6
	FUNC_FIXSTATECLUSTR = 1 << 7

	FUNC_JOINEMPTY       = 1 << 8
	FUNC_REPLICATE       = 1 << 9
	FUNC_JOINBALANCE     = 1 << 10
	FUNC_MIGRATEALLCROSS = 1 << 11

	FUNC_FIXSLOTS     = 1 << 12
	FUNC_SARTMULTIPLE = 1 << 13

	FUNC_ADDSLOTS = 1 << 14
	FUNC_DELSLOTS = 1 << 15

	FUNC_MIGRATESLOTSIN    = 1 << 16
	FUNC_MIGRATESLOTSCROSS = 1 << 17
)

var (
	OneNodeMap = map[int]OpOneNode{ //      address
		FUNC_STARTSINGLE:    StartClusterSingle,
		FUNC_DESTROY:        DestoryCluster,
		FUNC_QUITEMPTY:      QuitEmptyFromCluster,
		FUNC_QUITBALANCE:    QuitFromClusterAndBalaceItsSlots,
		FUNC_FIXSTATESINGLE: MigratingNodeSlotsFix,
		FUNC_FIXSTATECLUSTR: MigratingClusterSlotsFix,
	}

	TwoNodeMap = map[int]OpTwoNode{ //    (address *address)
		FUNC_JOINEMPTY:       JoinToCluster,
		FUNC_REPLICATE:       ReplicateToClusterNode,
		FUNC_JOINBALANCE:     JoinToClusterAndBalanceSlots,
		FUNC_MIGRATEALLCROSS: MigrateAllDataCorssCluster,
	}

	MultiNodeMap = map[int]OpMultiNode{ //    (...*address)
		FUNC_FIXSLOTS:     FixClusterSlotsMuiltAddr,
		FUNC_SARTMULTIPLE: StartClusterMutilple,
	}

	OneIntsMap = map[int]OpOneIntsNode{ //    (*address ,...slots)
		FUNC_ADDSLOTS: AddSlotInNode,
		FUNC_DELSLOTS: DeleteSlotInNode,
	}
	TwoIntsMap = map[int]OpTwoIntsNode{ //    (*address ,*address ,...slots)
		FUNC_MIGRATESLOTSIN:    MigrateSlotsBelongToCluster,
		FUNC_MIGRATESLOTSCROSS: MigrateSlotsDataCorssCluster,
	}

	CMD = []string{
		CMD_LIST,
		CMD_GETINFO,
		CMD_STARTSINGLE,
		CMD_DESTROY,
		CMD_QUITEMPTY,
		CMD_QUITBALANCE,
		CMD_FIXSTATESINGLE,
		CMD_FIXSTATECLUSTR,
		CMD_JOINEMPTY,
		CMD_REPLICATE,
		CMD_JOINBALANCE,
		CMD_MIGRATEALLCROSS,
		CMD_FIXSLOTS,
		CMD_SARTMULTIPLE,
		CMD_ADDSLOTS,
		CMD_DELSLOTS,
		CMD_MIGRATESLOTSIN,
		CMD_MIGRATESLOTSCROSS,
	}

	HINTS = []string{
		"<address> show list nodes",
		"<address> get cluster info",
		"<address> start cluster with single node",
		"<address> destroy cluster absolutely",
		"<address> empty node quit from cluster",
		"<address> non empty master quit from cluster",
		"<address> fix slots' state with single node",
		"<address> fix cluster's slots state",
		"<address> <address> an empty node join into cluster",
		"<address> <address> replicate an empty master node slave to a master node",
		"<address> <address> an empty master node join to be a master and balance their slots with it",
		"<address> <address> migrate all slots' data cross cluster",
		"[address...] fix slots with nodes",
		"[address...] start cluster with multiple master",
		"<address> [slots...] add slots to master node",
		"<address> [slots...] delete sltos from master node",
		"<address> <address> [slots...] migrate slots between cluster's two node",
		"<address> <address> [slots...] migrate slots' data cross cluster",
	}
	argsvalue = [18]bool{}
)

type OpOneNode func(string, string) (string, error)
type OpTwoNode func(string, string, string, string) (string, error)
type OpMultiNode func(...*Address) (string, error)
type OpOneIntsNode func(string, string, ...int) (string, error)
type OpTwoIntsNode func(string, string, string, string, ...int) (string, error)

type Args struct {
	value *bool
	name  string
	hint  string
}

func initVar(args ...Args) {
	for _, arg := range args {
		flag.BoolVar(arg.value, arg.name, false, arg.hint)
	}
}
func main() {
	var addr *Address
	var err error
	var res interface{}
	op := 0
	args := make([]Args, len(argsvalue), len(argsvalue))
	for i := 0; i < len(argsvalue); i++ {
		args[i] = Args{&(argsvalue[i]), CMD[i], HINTS[i]}
	}
	initVar(args...)
	flag.Parse()
	v := 1
	for i := 0; i < len(argsvalue); i++ {
		if argsvalue[i] {
			op |= v
		}
		v = v << 1
	}
	if op&(op-1) > 0 {
		ShowWrongArgsError()
		return
	}
	if op == FUNC_LISTNODES || op == FUNC_GETINFO {
		if flag.NArg() != 1 {
			ShowWrongArgsError()
			return
		}
		addr, err = NewAddress(flag.Args()[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		if op == FUNC_LISTNODES {
			res, _, err = ListNodesInCluster(addr.Host, addr.Port)
		} else {
			res, err = GetClusterInfo(addr.Host, addr.Port)
		}
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(res)
		return
	}
	if op <= FUNC_FIXSTATECLUSTR {
		if flag.NArg() != 1 {
			ShowWrongArgsError()
			return
		}
		addr, err = NewAddress(flag.Args()[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		opfunc := OneNodeMap[op]
		res, err = opfunc(addr.Host, addr.Port)
	} else if op <= FUNC_MIGRATEALLCROSS {
		if flag.NArg() != 2 {
			ShowWrongArgsError()
			return
		}
		addrs, err := getAddresses(flag.Args()...)
		if err != nil {
			fmt.Println(err)
			return
		}
		opfunc := TwoNodeMap[op]
		res, err = opfunc(addrs[0].Host, addrs[0].Port, addrs[1].Host, addrs[1].Port)
	} else if op <= FUNC_SARTMULTIPLE {
		addrs, err := getAddresses(flag.Args()...)
		if err != nil {
			fmt.Println(err)
			return
		}
		opfunc := MultiNodeMap[op]
		res, err = opfunc(addrs...)
	} else if op <= FUNC_DELSLOTS {
		if flag.NArg() < 2 {
			ShowWrongArgsError()
			return
		}
		addr, err = NewAddress(flag.Args()[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		slots, err := getSlots(flag.Args()[1:]...)
		if err != nil {
			fmt.Println(err)
			return
		}
		opfunc := OneIntsMap[op]
		res, err = opfunc(addr.Host, addr.Port, slots...)
	} else if op <= FUNC_MIGRATESLOTSCROSS {
		if flag.NArg() < 3 {
			ShowWrongArgsError()
			return
		}
		addrs, err := getAddresses(flag.Args()[:2]...)
		if err != nil {
			fmt.Println(err)
			return
		}
		slots, err := getSlots(flag.Args()[2:]...)
		if err != nil {
			fmt.Println(err)
			return
		}
		opfunc := TwoIntsMap[op]
		res, err = opfunc(addrs[0].Host, addrs[0].Port, addrs[1].Host, addrs[1].Port, slots...)
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
}

func addressFunc(i int) func(host, port string) (string, error) {
	return StartClusterSingle
}

func ShowWrongArgsError() {
	fmt.Println("wrong arguments")
}
func getAddresses(values ...string) ([]*Address, error) {
	addrs := make([]*Address, len(values), len(values))
	for i, v := range values {
		addr, err := NewAddress(v)
		if err != nil {
			return nil, err
		}
		addrs[i] = addr
	}
	return addrs, nil
}

func getSlots(values ...string) ([]int, error) {
	slots := make([]int, len(values), len(values))
	var slot int
	var err error
	for i, v := range values {
		slot, err = strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		slots[i] = slot
	}
	return slots, nil
}
