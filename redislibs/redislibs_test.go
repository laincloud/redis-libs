package redislibs

import (
	"fmt"
	// "github.com/laincloud/redis-libs/network"
	// "net"
	"testing"
	// "time"
)

func Test_ListNodes(t *testing.T) {

	// master_addr, err := GetMasterAddrByName("127.0.0.1", "5000", "mymaster")
	// if err == nil {
	// 	t.Log("测试通过了, 地址:", master_addr) //记录一些你期望记录的信息

	// }
	// SlaveOf("127.0.0.1", "6003", "127.0.0.1", "6001")

	// SlavesBySentinel("127.0.0.1", "5000", "mymaster")

	// if slaves, err := Slaves(master_addr.Host, master_addr.Port); err == nil {
	// 	t.Log("测试通过了, slave地址:", slaves)
	// }

	// sentinel_addrs := []*Address{BuildAddress("127.0.0.1", "5001"), BuildAddress("127.0.0.1", "5002"), BuildAddress("127.0.0.1", "5000")}
	// if err := RemoveSlaveFromSentinel("127.0.0.1", "6002", "master_service", sentinel_addrs...); err == nil {
	// 	t.Log("remove slave from sentinel pass")
	// }
	// if err := RemoveSlaveFromSentinel("127.0.0.1", "6002", "master_service", sentinel_addrs...); err == nil {
	// 	t.Log("remove slave from sentinel pass")
	// }
	// res, _ := GetMasterAddrByName("127.0.0.1", "26379", "mymaster")
	// fmt.Println(res)
	r, _ := GetMasterAddrByName("127.0.0.1", "26379", "mymaster")
	fmt.Println(r)

	// GetSlavesInSentinel("127.0.0.1", "5000", "master_service")
	GetSlavesInSentinel("127.0.0.1", "26379", "master_service2")

	// RoleStatus("127.0.0.1", "6001")
	// if role, status, err := RoleStatus("127.0.0.1", "6002"); err == nil {
	// 	fmt.Println(role, status)
	// }

	infos, err := RedisNodeInfo("127.0.0.1", "6001")
	if err != nil {
		fmt.Println("err:", err.Error())
	}
	fmt.Println(infos)

	// masters, err := FetchMastersInSentinel("127.0.0.1", "26379")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// for _, m := range masters {
	// 	fmt.Println(m)
	// }
	// addr, _ := NewAddress(":5000")
	// fmt.Println(addr)

}

// func TestRedisConn(t *testing.T) {
// 	c, _ := net.DialTimeout("tcp", "127.0.0.1:6379", time.Duration(10*time.Second))
// 	rc, err := network.NewRedisConn(c, network.NewConnectOption(10, 10, 1024))
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	// cmds := []string{Pack_command("info"), Pack_command("get", "100"), Pack_command("keys", "111*")}
// 	// for _, cmd := range cmds {
// 	// 	rc.Write([]byte(cmd))
// 	// 	if res, err := rc.ReadAll(); err == nil {
// 	// 		fmt.Println("res:", string(res))
// 	// 	} else {
// 	// 		fmt.Println("err:", err.Error())
// 	// 	}
// 	// }
// 	cmd := Pack_command("keys", "*")
// 	for i := 0; i < 1; i++ {
// 		rc.Write([]byte(cmd))
// 		if res, err := rc.ReadAll(); err == nil {
// 			fmt.Println("res:", string(res))
// 		} else {
// 			fmt.Println("err:", err)
// 		}
// 	}
// }

// func TestTransactionConn(t *testing.T) {
// 	c, _ := net.DialTimeout("tcp", "127.0.0.1:6379", time.Duration(10*time.Second))
// 	rc, err := network.NewRedisConn(c, network.NewConnectOption(10, 10, 1024))
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	// for {
// 	cmds := make([]byte, 0)
// 	cmd := Pack_command("multi")
// 	cmds = append(cmds, cmd...)
// 	cmd = Pack_command("get", "a")
// 	times := 10000
// 	for i := 0; i < times; i++ {
// 		cmds = append(cmds, cmd...)
// 	}
// 	cmd = Pack_command("exec")
// 	cmds = append(cmds, cmd...)
// 	rc.Write([]byte(cmds))
// 	for i := 0; i < times+1; i++ {
// 		if res, err := rc.ReadAll(); err == nil {
// 			fmt.Println("res0:", string(res))
// 		} else {
// 			fmt.Println("err:", err.Error())
// 		}
// 	}
// 	if res, err := rc.ReadAll(); err == nil {
// 		fmt.Println("res:", string(res))
// 	}
// }
