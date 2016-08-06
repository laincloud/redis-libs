package redislibs

import (
	"errors"
	"fmt"
	. "github.com/laincloud/redis-libs/utils"
	"github.com/mijia/sweb/log"
	"strconv"
	"strings"
	"time"
)

func ListNodesInCluster(Host, Port string) ([]*ClusterNode, *ClusterNode, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return nil, nil, err
	}
	return listNodesInCluster(t)
}

func ListMasterNodes(cs_nodes []*ClusterNode) []*ClusterNode {
	return listMasterNodes(cs_nodes)
}

func GetClusterInfo(Host, Port string) (map[string]string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return nil, err
	}
	return getClusterInfo(t)
}

func StartClusterSingle(Host, Port string) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	_, myself, err := listNodesInCluster(t)
	if len(myself.Assigned_slots) > 0 {
		return fmt.Sprintf(STATUS_ERR+" node %s:%s cant' init to be a cluster", Host, Port), nil
	}

	slots := make([]int, SLOT_COUNT, SLOT_COUNT)
	for i := 0; i < SLOT_COUNT; i++ {
		slots[i] = i
	}

	m, err := addSlotNode(t, slots)
	log.Infof("Ask `cluster addslots` Rsp %s", m)
	if err != nil || m != STATUS_OK {
		return m, err
	}

	log.Infof("nodes %s %s started as a standalone cluster", Host, Port)
	return STATUS_OK, nil
}

func DestoryCluster(Host string, Port string) (string, error) {
	cs_nodes, _, err := ListNodesInCluster(Host, Port)
	if err != nil {
		return "", err
	}
	res := 1
	for _, node := range cs_nodes {
		log.Infof("start destroy node %s %s ", node.Host, node.Port)
		_, err := DeleteSlotInNode(node.Host, node.Port, node.Assigned_slots...)
		if err != nil {
			log.Warnf("Node %s %s can not be destroied", node.Host, node.Port)
			res = -1
			continue
		}
		node_t, err := BuildTalker(node.Host, node.Port)
		defer node_t.Close()
		if err != nil {
			res = -1
			continue
		}
		m, err := clusterReset(node_t)
		if err != nil || m != STATUS_OK {
			res = -1
		}
	}
	if res == -1 {
		log.Warnf("Faltal Error when destroy cluster")
		return "", errors.New("Some nodes can not be destroied when destroying cluster")
	}
	log.Infof("cluster was destroied: %s:%s", Host, Port)
	return STATUS_OK, nil
}

func JoinToCluster(Host, Port, cls_Host, cls_Port string) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	_, myself, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	if len(myself.Assigned_slots) > 0 ||
		strings.Contains(myself.Role_in_cluster, ROLE_SLAVE) {
		return STATUS_ERR + " node can't join the cluster or in cluster alerady", nil
	}
	log.Infof("node %s %s start join into the cluster", Host, Port)
	return joinToCluster(t, cls_Host, cls_Port)
}

func ReplicateToClusterNode(Host, Port, mstr_Host, mstr_Port string) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}

	_, myself, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	if len(myself.Assigned_slots) > 0 {
		return STATUS_ERR + " node has slots can't be relicated", nil
	}
	m, err := joinToCluster(t, mstr_Host, mstr_Port)
	if err != nil || m != STATUS_OK {
		return m, err
	}

	m, err = flushAllData(t)
	if err != nil || m != STATUS_OK {
		return m, err
	}

	cs_nodes, myself, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}

	var master_node *ClusterNode
	for _, node := range cs_nodes {
		if node.Host == mstr_Host && node.Port == mstr_Port {
			master_node = node
			break
		}
	}
	if master_node == nil {
		return STATUS_ERR + " master didn't in cluster", nil
	}

	if err != nil || m != STATUS_OK {
		return m, err
	}
	log.Infof("node %s:%s start to be slave of %s:%s", Host, Port, mstr_Host, mstr_Port)
	return replicateToNode(t, master_node.Node_id)
}

func QuitEmptyFromCluster(Host, Port string) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, myself, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	if len(myself.Assigned_slots) > 0 {
		log.Warnf(STATUS_ERR + " Node still has slots and can't quit from cluster")
		return STATUS_ERR + " Node still has slots and can't quit from cluster", nil
	}
	log.Infof("node %s:%s start to quit from cluster", Host, Port)
	return quitEmptyNodeFromCluster(t, cs_nodes, myself)
}

func QuitFromClusterAndBalaceItsSlots(Host, Port string) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, myself, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}

	if len(myself.Assigned_slots) == 0 {
		return quitEmptyNodeFromCluster(t, cs_nodes, myself)
	}
	master := listMasterNodes(cs_nodes)
	if len(master) == 1 {
		log.Warnf(STATUS_ERR + " no more master for slots balance, only if you destroy the cluster")
		return STATUS_ERR + " no more master for slots balance, only if you destroy the cluster", nil
	}
	log.Infof("node %s:%s start to quit from cluster", Host, Port)
	log.Infof("calculating slots balance!")
	slots_to := balanceSlotsToOthers(len(master)-1, myself.Assigned_slots)
	log.Infof("migrating data!")
	i := 0
	for _, node := range master {
		if node.Node_id == myself.Node_id {
			continue
		}
		dst_t, err := BuildTalker(node.Host, node.Port)
		defer dst_t.Close()
		if err != nil {
			return "", err
		}
		for _, slot := range slots_to[i] {
			m, err := migrateSlot(t, dst_t, myself, node, slot)
			if err != nil || m != STATUS_OK {
				return m, err
			}
		}
		i++
	}
	return quitEmptyNodeFromCluster(t, cs_nodes, myself)

}

func FixClusterSlotsSingle(Host, Port string) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, myself, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	if !strings.Contains(myself.Role_in_cluster, ROLE_MASTER) {
		return STATUS_ERR + " Node must be master", nil
	}
	if len(myself.Assigned_slots) > 0 {
		return STATUS_ERR + " Node must be empty", nil
	}
	Assigned_slots := getAssignedSlots(cs_nodes)
	assigned_size := len(Assigned_slots)
	if assigned_size == SLOT_COUNT {
		return STATUS_ERR + " Cluster is alerady healthy", nil
	}
	total_Assigned_slots := MergerSortedSlices(Assigned_slots...)
	unAssigned_slots := unAssignedSlots(total_Assigned_slots)
	log.Infof("node %s:%s start to fix slots in cluster", Host, Port)
	return addSlotNode(t, unAssigned_slots)
}

func MigratingNodeSlotsFix(Host, Port string) (string, error) {

	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, _, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	log.Infof("node %s:%s start to fix slots' migrating or importing state of node", Host, Port)
	return fixMigratingSlots(t, cs_nodes...)
}

func MigratingClusterSlotsFix(Host, Port string) (string, error) {
	log.Infof("node %s:%s start to fix slots' migrating or importing state of cluster", Host, Port)
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, _, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	failed := make([]*ClusterNode, 0, 0)
	var selfNode *ClusterNode
	var T *Talker
	for _, node := range cs_nodes {
		T, err = BuildTalker(node.Host, node.Port)
		defer T.Close()
		if err != nil {
			failed = append(failed, node)
			continue
		}
		_, selfNode, err = listNodesInCluster(T)
		if err != nil {
			failed = append(failed, node)
			continue
		}
		_, err = fixMigratingSlots(T, selfNode)
		if err != nil {
			failed = append(failed, node)
		}
	}
	if len(failed) > 0 {
		fail_text := STATUS_ERR
		for _, fn := range failed {
			fail_text += fmt.Sprintf(" node %s:%s ,", fn.Host, fn.Port)
		}
		fail_text += "fix failed"
		return fail_text, nil
	}
	return STATUS_OK, nil
}

func FixClusterSlotsMuiltAddr(addrs ...*Address) (string, error) {

	if len(addrs) == 0 {
		return "", nil
	}

	t, err := BuildTalker(addrs[0].Host, addrs[0].Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, _, err := listNodesInCluster(t)
	if err != nil {
		return "", err
	}
	nodes := make([]*ClusterNode, len(addrs), len(addrs))
	for i, addr := range addrs {
		pos := 0
		for _, node := range cs_nodes {
			if node.Host == addr.Host && node.Port == addr.Port {
				if len(node.Assigned_slots) > 0 {
					log.Warnf(STATUS_ERR + " Node must be empty")
					return STATUS_ERR + " Node must be empty", nil
				}
				nodes[i] = node
				break
			}
			pos++
		}
		if pos == len(cs_nodes) {
			log.Warnf(STATUS_ERR + "addrs are not in the same cluster")
			return STATUS_ERR + "addrs are not in the same cluster", nil
		}
	}
	Assigned_slots := getAssignedSlots(cs_nodes)
	assigned_size := len(Assigned_slots)
	if assigned_size == SLOT_COUNT {
		return STATUS_OK, nil
	}
	total_Assigned_slots := MergerSortedSlices(Assigned_slots...)
	unAssigned_slots := unAssignedSlots(total_Assigned_slots)

	slots := balanceSlotsToOthers(len(addrs), unAssigned_slots)
	for i, node := range nodes {
		node_t, err := BuildTalker(node.Host, node.Port)
		defer node_t.Close()
		if err != nil {
			return "", err
		}
		m, err := addSlotNode(node_t, slots[i])
		if err != nil || m != STATUS_OK {
			return m, err
		}
	}
	return STATUS_OK, nil
}

func StartClusterMutilple(addrs ...*Address) (string, error) {
	if len(addrs) <= 0 {
		return STATUS_ERR + " No nodes", nil
	}
	if len(addrs) == 1 {
		return StartClusterSingle(addrs[0].Host, addrs[0].Port)
	}
	mstr_size := len(addrs)
	node_slots := initAndBalanceSlots(mstr_size)
	t, err := BuildTalker(addrs[0].Host, addrs[0].Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	m, err := addSlotNode(t, node_slots[0])
	log.Infof("node %s:%s join cluster %s", addrs[0].Host, addrs[0].Port, m)
	if err != nil || m != STATUS_OK {
		return m, err
	}
	var addr *Address
	var node_t *Talker
	for i := 1; i < mstr_size; i++ {
		addr = addrs[i]
		node_t, err = BuildTalker(addr.Host, addr.Port)
		defer node_t.Close()
		if err != nil {
			return "", err
		}

		m, err := joinToCluster(node_t, addrs[0].Host, addrs[0].Port)
		if err != nil {
			log.Info(err.Error())
		}
		if err != nil || m != STATUS_OK {
			log.Infof("node %s:%s join cluster failed", addr.Host, addr.Port)
			return m, err
		}

		m, err = addSlotNode(node_t, node_slots[i])
		log.Infof("node %s:%s join cluster %s", addr.Host, addr.Port, m)
		if err != nil || m != STATUS_OK {
			return m, err
		}
	}
	return "", nil
}

func JoinToClusterAndBalanceSlots(Host, Port, clsHost, clsPort string) (string, error) {
	src_t, err := BuildTalker(Host, Port)
	defer src_t.Close()
	if err != nil {
		return "", err
	}
	_, src_myself, err := listNodesInCluster(src_t)
	if err != nil {
		return "", err
	}
	if len(src_myself.Assigned_slots) > 0 {
		info := STATUS_ERR + " node still has slots and can't join others"
		log.Warnf(info)
		return info, nil
	}
	cls_t, err := BuildTalker(clsHost, clsPort)
	defer cls_t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, _, err := listNodesInCluster(cls_t)
	if err != nil {
		return "", err
	}
	m, err := joinToCluster(src_t, clsHost, clsPort)
	if err != nil || m != STATUS_OK {
		return m, err
	}
	wait_time := 0
	for {
		log.Infof("wait %d second make sure the node is joined into the cluster absolately", wait_time+5)
		time.Sleep(5 * time.Second)
		if wait_time >= INFO_TIME_OUT {
			log.Warnf(STATUS_ERR + "GET CLUSTER INFO TIMEOUT")
			return STATUS_ERR + "GET CLUSTER INFO TIMEOUT", nil
		}
		wait_time += 5
		infos, err := getClusterInfo(src_t)
		if err != nil {
			return "fail", err
		} else if infos[INFO_CLUSTER_STATE] != "ok" {
			log.Warn("Cluster Info Get Failed: ", infos[INFO_CLUSTER_STATE])
		} else {
			break
		}
	}

	master_nodes := listMasterNodes(cs_nodes)
	cls_mstr_size := len(master_nodes)
	sizes := balanceSlotsSize(cls_mstr_size+1, SLOT_COUNT)
	var node *ClusterNode
	for i, size := range sizes[:cls_mstr_size] {
		node = master_nodes[i]
		log.Infof("migrate slots from %s:%s to %s:%s", node.Host, node.Port, Host, Port)
		slts_t, err := BuildTalker(node.Host, node.Port)
		defer slts_t.Close()
		if err != nil {
			return "", err
		}
		for _, slot := range node.Assigned_slots[size:] {
			m, err := migrateSlot(slts_t, src_t, node, src_myself, slot)
			if err != nil || m != STATUS_OK {
				return m, err
			}
		}
	}
	return STATUS_OK, nil
}

func MigrateAllDataCorssCluster(srcHost, srcPort, dstHost, dstPort string) (string, error) {
	src_t, err := BuildTalker(srcHost, srcPort)
	defer src_t.Close()
	if err != nil {
		return "", err
	}

	dst_t, err := BuildTalker(dstHost, dstPort)
	defer dst_t.Close()
	if err != nil {
		return "", err
	}

	src_cs_nodes, _, err := listNodesInCluster(src_t)
	if err != nil {
		return "", err
	}
	src_masters := listMasterNodes(src_cs_nodes)

	dst_cs_nodes, _, err := listNodesInCluster(dst_t)
	if err != nil {
		return "", err
	}
	dst_masters := listMasterNodes(dst_cs_nodes)

	dst_slots_map := [SLOT_COUNT]Address{}
	for _, dst_master := range dst_masters {
		for _, slot := range dst_master.Assigned_slots {
			dst_slots_map[slot] = Address{dst_master.Host, dst_master.Port}
		}
	}
	err_occrd := false
	var dst_addr Address
	for _, node := range src_masters {
		src_mt, err := BuildTalker(node.Host, node.Port)
		defer src_mt.Close()
		if err != nil {
			return "", err
		}
		for _, slot := range node.Assigned_slots {
			dst_addr = dst_slots_map[slot]
			_, err := migrateSlotData(src_mt, dst_addr.Host, dst_addr.Port, slot)
			if err != nil {
				err_occrd = true
				continue
			}
		}
	}
	if err_occrd {
		return STATUS_ERR + "error occured", nil
	}
	return STATUS_OK, nil
}

func AddSlotInNode(Host, Port string, slots ...int) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	_, myself, err := listNodesInCluster(t)
	if !strings.Contains(myself.Role_in_cluster, ROLE_MASTER) {
		return STATUS_ERR + " node must be master", nil
	}
	return addSlotNode(t, slots)
}

func DeleteSlotInNode(Host, Port string, slots ...int) (string, error) {
	t, err := BuildTalker(Host, Port)
	defer t.Close()
	if err != nil {
		return "", err
	}
	_, myself, err := listNodesInCluster(t)
	if !strings.Contains(myself.Role_in_cluster, ROLE_MASTER) {
		return STATUS_ERR + " node must be master", nil
	}
	if !slotsBelongsToNode(myself, slots) {
		return STATUS_ERR + " node must contain the slots", nil
	}
	m, err := deleteSlotNode(t, slots)
	return m, nil
}

func MigrateSlotsBelongToCluster(srcHost, srcPort, dstHost, dstPort string, slots ...int) (string, error) {
	src_t, err := BuildTalker(srcHost, srcPort)
	defer src_t.Close()
	if err != nil {
		return "", err
	}
	cs_nodes, src, err := listNodesInCluster(src_t)
	if err != nil {
		return "", err
	}
	innode := slotsBelongsToNode(src, slots)
	if !innode {
		return "", errors.New(fmt.Sprintf("some or all slots are not belongs to node %s %s\n", src.Host, src.Port))
	}

	dst_t, err := BuildTalker(dstHost, dstPort)
	defer dst_t.Close()
	if err != nil {
		return "", err
	}
	var dst *ClusterNode
	for _, node := range cs_nodes {
		if node.Host == dstHost && node.Port == dstPort {
			dst = node
			break
		}
	}
	if dst == nil {
		return "", errors.New("Destination node is not in the cluster")
	}

	for _, slot := range slots {
		_, err := migrateSlot(src_t, dst_t, src, dst, slot)
		if err != nil {
			return "", err
		}
	}
	return STATUS_OK, nil
}

func MigrateSlotsDataCorssCluster(srcHost, srcPort, dstHost, dstPort string, slots ...int) (string, error) {
	src_t, err := BuildTalker(srcHost, srcPort)
	defer src_t.Close()
	if err != nil {
		return "", err
	}
	_, srcNode, err := listNodesInCluster(src_t)
	if err != nil {
		return "", err
	}
	innode := slotsBelongsToNode(srcNode, slots)
	if !innode {
		return "", errors.New(fmt.Sprintf("some or all slots are not belongs to node %s %s\n", srcNode.Host, srcNode.Port))
	}
	for _, slot := range slots {
		_, err := migrateSlotData(src_t, dstHost, dstPort, slot)
		if err != nil {
			// return "", err
			continue
		}
	}

	return STATUS_OK, nil
}

/*
*--------------------------------------------------------------------------
 */

func listNodesInCluster(t *Talker) ([]*ClusterNode, *ClusterNode, error) {
	r, err := t.TalkRaw(COMMAND_CLUSTER_NODES)
	if err != nil {
		return nil, nil, err
	}
	var myself *ClusterNode
	nodes := strings.Split(r, SYM_LF)
	cs_nodes := make([]*ClusterNode, 0, len(nodes))
	for _, v := range nodes {
		args := strings.Split(v, " ")
		node, err := NewClusterNode(args)
		if node == nil {
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		if node.Host == "" {
			node.Host = "127.0.0.1"
		}
		if strings.Contains(node.Role_in_cluster, ROLE_MYSELF) {
			myself = node
		}
		cs_nodes = append(cs_nodes, node)
	}
	return cs_nodes, myself, nil
}

func listMasterNodes(cs_nodes []*ClusterNode) []*ClusterNode {
	master := make([]*ClusterNode, 0, 0)
	for _, node := range cs_nodes {
		if strings.Contains(node.Role_in_cluster, ROLE_MASTER) &&
			len(node.Assigned_slots) > 0 {
			master = append(master, node)
		}
	}
	return master
}

func getClusterInfo(t *Talker) (map[string]string, error) {
	resp, err := t.TalkRaw(Pack_command("CLUSTER", "INFO"))
	if err != nil {
		return nil, err
	}
	res := stringToArray(resp)
	infos := make(map[string]string)
	for _, r := range res {
		kv := strings.Split(r, ":")
		infos[kv[0]] = kv[1]
	}
	return infos, nil
}

func quitEmptyNodeFromCluster(t *Talker, cs_nodes []*ClusterNode, myself *ClusterNode) (string, error) {
	var node_t *Talker
	var err error
	failed := make([]*ClusterNode, 0, 0)
	for _, node := range cs_nodes {
		if myself.Node_id == node.Node_id {
			continue
		}
		//Others forget me
		node_t, err = BuildTalker(node.Host, node.Port)
		defer node_t.Close()
		if err != nil {
			failed = append(failed, node)
			continue
		}
		m, err := quitFromCluster(node_t, myself.Node_id)
		if err != nil || m != STATUS_OK {
			failed = append(failed, node)
			continue
		}

	}
	//I forget others
	m, err := clusterReset(t)
	if err != nil || m != STATUS_OK {
		return m, err
	}
	if len(failed) > 0 {
		fail_text := STATUS_ERR
		for _, fn := range failed {
			fail_text += fmt.Sprintf(" node %s:%s ,", fn.Host, fn.Port)
		}
		fail_text += "forget failed"

		return fail_text, nil
	}
	log.Infof("node %s:%s quit cluster successfully", myself.Host, myself.Port)
	return STATUS_OK, nil
}

func migrateSlotData(t *Talker, dstHost string, dstPort string, slot int) (string, error) {
	key_count := 0
	keys_err := make([]string, 0, 0)
	for {
		keysObj, err := t.TalkForObject(Pack_command("CLUSTER", "GETKEYSINSLOT", strconv.Itoa(slot), strconv.Itoa(WIND_MIGRATE)))
		if err != nil {
			return "", err
		}
		keys := keysObj.([]interface{})
		if len(keys) <= len(keys_err) {
			if len(keys_err) > 0 {
				res := fmt.Sprintf(STATUS_ERR+" %d nums of keys as %v migrate failed", len(keys_err), keys_err)
				log.Infof(res)
				return res, nil
			}
			return STATUS_OK, nil
		}
		key_count += len(keys)
		for _, key := range keys {
			k := key.(string)
			m, err := t.TalkRaw(Pack_command("MIGRATE", dstHost, dstPort, k, strconv.Itoa(REDIS_DB_NO), strconv.Itoa(TIMEOUT_MIGRATE)))
			if err != nil {
				log.Infof("key %s migrate Resp:%s", k, m)
				keys_err = append(keys_err, k)
			}
		}

	}
	return STATUS_OK, nil
}

func migrateSlot(src_t *Talker, dst_t *Talker, srcNode *ClusterNode, dstNode *ClusterNode, slot int) (string, error) {
	_, err := setSlotInMigrating(src_t, dstNode.Node_id, slot)
	if err != nil {
		return "", err
	}
	_, err = setSlotInImpoerting(dst_t, srcNode.Node_id, slot)
	if err != nil {
		return "", err
	}
	_, err = migrateSlotData(src_t, dstNode.Host, dstNode.Port, slot)
	if err != nil {
		return "", err
	}
	_, err = setSlotToNode(src_t, dstNode.Node_id, slot)
	if err != nil {
		return "", err
	}
	_, err = setSlotToNode(dst_t, dstNode.Node_id, slot)
	if err != nil {
		return "", err
	}
	return STATUS_OK, nil

}

func fixMigratingSlots(t *Talker, cs_nodes ...*ClusterNode) (string, error) {
	for _, node := range cs_nodes {
		for _, mgrtingSlot := range node.migrating_slots {
			_, err := setSlotInStable(t, mgrtingSlot)
			if err != nil {
				return "", err
			}
		}
	}
	return STATUS_OK, nil
}

func joinToCluster(t *Talker, cls_Host, cls_Port string) (string, error) {
	return t.TalkRaw(Pack_command("cluster", "meet", cls_Host, cls_Port))
}

func quitFromCluster(t *Talker, nodeid string) (string, error) {
	return t.TalkRaw(Pack_command("cluster", "forget", nodeid))
}
func replicateToNode(t *Talker, master_nodeid string) (string, error) {
	return t.TalkRaw(Pack_command("cluster", "replicate", master_nodeid))
}

func setSlotInMigrating(t *Talker, dstNodeId string, slot int) (string, error) {
	return t.TalkRaw(Pack_command("CLUSTER", "SETSLOT", strconv.Itoa(slot), "MIGRATING", dstNodeId))
}

func setSlotInImpoerting(t *Talker, srcNodeId string, slot int) (string, error) {
	return t.TalkRaw(Pack_command("CLUSTER", "SETSLOT", strconv.Itoa(slot), "importING", srcNodeId))
}

func setSlotInStable(t *Talker, slot int) (string, error) {
	return t.TalkRaw(Pack_command("CLUSTER", "SETSLOT", strconv.Itoa(slot), "STABLE"))
}

func setSlotToNode(t *Talker, dstNodeId string, slot int) (string, error) {
	return t.TalkRaw(Pack_command("CLUSTER", "SETSLOT", strconv.Itoa(slot), "NODE", dstNodeId))
}

func flushAllData(t *Talker) (string, error) {
	return t.TalkRaw(Pack_command("FLUSHALL"))
}

func clusterReset(t *Talker) (string, error) {
	m, err := flushAllData(t)
	if err != nil {
		return m, err
	}
	return t.TalkRaw(Pack_command("CLUSTER", "RESET"))
}

func deleteSlotNode(t *Talker, slots []int) (string, error) {
	if len(slots) == 0 {
		return "", nil
	}
	args := make([]string, 2, len(slots)+2)
	args[0] = "CLUSTER"
	args[1] = "DELSLOTS"
	for _, v := range slots {
		args = append(args, strconv.Itoa(v))
	}
	m, err := t.TalkRaw(Pack_command(args...))
	if err != nil {
		return m, err
	}
	return m, nil
}

func addSlotNode(t *Talker, slots []int) (string, error) {
	if len(slots) == 0 {
		return "", nil
	}
	args := make([]string, 2, len(slots)+2)
	args[0] = "CLUSTER"
	args[1] = "ADDSLOTS"
	for _, v := range slots {
		args = append(args, strconv.Itoa(v))
	}
	m, err := t.TalkRaw(Pack_command(args...))
	if err != nil {
		return m, err
	}
	return m, nil
}
