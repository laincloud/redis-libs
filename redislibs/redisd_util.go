package redislibs

import "strings"

func unAssignedSlots(Assigned_slots []int) []int {
	size := SLOT_COUNT - len(Assigned_slots)
	unAssigned_slots := make([]int, size, size)
	start := 0
	p := 0
	for _, slot := range Assigned_slots {
		if slot != start {
			for start < slot {
				unAssigned_slots[p] = start
				start++
				p++
			}
		}
		start = slot + 1
	}
	for start < SLOT_COUNT {
		unAssigned_slots[p] = start
		p++
		start++
	}
	return unAssigned_slots
}

/*
 * slots must be sorted ascend
 */
func slotsBelongsToNode(cnode *ClusterNode, slots []int) bool {
	cpos, spos := 0, 0
	for cpos < len(cnode.Assigned_slots) && spos < len(slots) {
		if cnode.Assigned_slots[cpos] > slots[spos] {
			return false
		} else if cnode.Assigned_slots[cpos] < slots[spos] {
			cpos++
		} else {
			cpos++
			spos++
		}
	}
	return spos == len(slots)
}

func balanceSlotsSize(mstr_size, total int) []int {
	size := make([]int, mstr_size, mstr_size)
	avg := total / mstr_size
	pos := total - avg*mstr_size
	for i := 0; i < pos; i++ {
		size[i] = avg + 1
	}
	for i := pos; i < mstr_size; i++ {
		size[i] = avg
	}
	return size
}

func initAndBalanceSlots(mstr_size int) [][]int {
	if mstr_size == 0 {
		return nil
	}
	size := balanceSlotsSize(mstr_size, SLOT_COUNT)
	node_slots := make([][]int, 0, 0)
	p := 0
	for i := 0; i < mstr_size; i++ {
		node_slots = append(node_slots, make([]int, 0, 0))
		for j := 0; j < size[i]; j++ {
			node_slots[i] = append(node_slots[i], p)
			p++
		}
	}
	return node_slots
}

func balanceSlotsToOthers(mstr_size int, slots []int) [][]int {
	if mstr_size == 0 {
		return nil
	}
	total := len(slots)
	size := balanceSlotsSize(mstr_size, total)
	node_slots := make([][]int, 0, 0)
	start := 0
	end := 0
	for i := 0; i < mstr_size; i++ {
		end += size[i]
		node_slots = append(node_slots, slots[start:end])
		start = end
	}
	return node_slots
}

func getAssignedSlots(cs_nodes []*ClusterNode) [][]int {
	mstr := listMasterNodes(cs_nodes)
	Assigned_slots := make([][]int, len(mstr), len(mstr))
	for i, node := range mstr {
		Assigned_slots[i] = node.Assigned_slots
	}
	return Assigned_slots
}

func stringToArray(resp string) []string {
	values := strings.Split(resp, SYM_CRLF)
	args := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" || strings.HasPrefix(v, SYM_STAR) || strings.HasPrefix(v, SYM_DOLLAR) {
			continue
		}
		args = append(args, v)
	}
	return args
}
