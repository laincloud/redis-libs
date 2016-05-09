# redis-libs
[![MIT license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://opensource.org/licenses/MIT)

operation library for redis cluster and redis sentinel

```
ListNodesInCluster : -list 
    list Clusternode of cluster
    params : Host,Port  {address of node}
    return : []ClusterNode, *ClusterNode    clusternode list and node it's slef
```

```
GetClusterInfo : -info --info
    get cluster map info
    params : Host,Port  {address of node}
    return : map[string]string   cluster state info and store in map
```

```
StartClusterSingle : -initsingle
    init a single master cluster and assigned all slots
    warns  : node must be empty
    params : Host,Port  {address of node}
    retrun : string   {info of the operation}
```

```
StartClusterMutilple : -initmultiple
    init with mutiple master nodes and assigned all slots balance to all nodes
    warns  : nodes must be empty
    params : addrs ...Address  {addresses of nodes to be master of the cluster}
    retrun : string   {info of the operation}
```

```
JoinToClusterAndBalanceSlots : -joinbalance
    make a new node join into the cluster and balance there slots
    warns  : node must be empty and single
    params : Host, Port, clsHost, clsPort     {source node's address and cluster address}
    retrun : string   {info of the operation}
```

```
DestoryCluster: -destroy
    destroy the whole cluster
    params : Host,Port  {one node's address in cluster}
    retrun : string   {info of the operation}
```

```
JoinToCluster:  -join
    join into a cluster and be an empty master with no Slots
    warns  : node must be empty when out of the cluster
    params : Host,Port  {one node's address out of cluster, in cluster is also ok but do nothing}
    retrun : string   {info of the operation}
```

```
ReplicateToCluster : -replicate
    make node to be a slave of one master in cluster
    warns  : node must be empty and single
    params : Host,Port,mstr_Host,mstr_Port     {address of slaver and master}
    retrun : string   {info of the operation}
```

```
QuitEmptyFromCluster: -quite
    make an empty node quit from cluster
    warns  : node must be empty and in cluster
    params : Host,Port     {address of node}
    retrun : string   {info of the operation}
```

```
QuitFromClusterAndBalaceItsSlots: -quitbalance
    make a master node assigned slots quit from cluster and balance its slots to others
    warns  : node must be in cluster
    params : Host,Port    {address of node}
    retrun : string   {info of the operation}
```

```
FixClusterSlotsMuiltAddr : -fixslots
    when some nodes in cluster got to be Chaos, such as some node quit from the cluster unexceptable
    warns  : nodes must be master in cluster and empty
    params : addrs ...Address    {addresses of node to be fixed}
    retrun : string   {info of the operation}
```

```
AddSlotInNode : -addslots
    add unassigned slot to master node
    warns  : node must be master and slots didn't be assigned
    params : Host,Port,slots []int   {address of node , slots to be added}
    retrun : string   {info of the operation}
```

```
DeleteSlotInNode : -deleteslots
    delete assigend slot in node
    warns  : node must be assigned with the slots
    params : Host,Port,slots []int   {address of node , slots to be deleted}
    retrun : string   {info of the operation}
```

```
MigrateSlotsBelongToCluster : -migrantein
    migrate slots inside cluster
    params : srcHost, srcPort, dstHost, dstPort string ,slots ...int    {from src node to dst node migrate slots}
    retrun : string   {info of the operation}
```

```
MigrateSlotsDataCorssCluster : -migrateout
    migrate slots cross cluster
    params : srcHost, srcPort, dstHost, dstPort string, slots ...int    {from src node to dst node cross cluster migrate slots}
    retrun : string   {info of the operation}
```

```
MigrateAllDataCorssCluster : -migarateallout
    migrate slots cross cluster
    params : srcHost, srcPort, dstHost, dstPort string,   {from src node to dst node cross cluster migrate all slots}
    retrun : string   {info of the operation}
```

```
MigratingNodeSlotsFix : -fixstate
    fix status of node's slots that in state of importing or migrating
    params : Host, Port string   {fix single node's migrate state}
    retrun : string   {info of the operation}
```

```
MigratingClusterSlotsFix : -fixclusterstate
    fix status of cluster's slots that in state of importing or migrating
    params : Host, Port string    {address of any one node in cluster}
    retrun : string   {info of the operation}
```
