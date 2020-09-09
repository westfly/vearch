// Copyright 2019 The Vearch Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package master

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/vearch/vearch/util/errutil"
	"math"
	"sort"
	"time"

	"github.com/tiglabs/raft/proto"
	"github.com/vearch/vearch/proto/vearchpb"
	"github.com/vearch/vearch/ps/engine/mapping"
	"github.com/vearch/vearch/util/cbjson"

	"github.com/spf13/cast"
	"github.com/vearch/vearch/client"
	"github.com/vearch/vearch/proto/entity"
	"github.com/vearch/vearch/util"
	"github.com/vearch/vearch/util/log"
	"go.etcd.io/etcd/clientv3/concurrency"
)

//masterService is used for master administrator purpose.It should not be used by router and partition server program
type masterService struct {
	*client.Client
}

func newMasterService(client *client.Client) (*masterService, error) {
	return &masterService{client}, nil
}

// registerServerService find nodeId partitions
func (ms *masterService) registerServerService(ctx context.Context, ip string, nodeID entity.NodeID) (*entity.Server, error) {
	server := &entity.Server{Ip: ip}

	spaces, err := ms.Master().QuerySpacesByKey(ctx, entity.PrefixSpace)
	if err != nil {
		return nil, err
	}

	for _, s := range spaces {
		for _, p := range s.Partitions {
			for _, id := range p.Replicas {
				if nodeID == id {
					server.PartitionIds = append(server.PartitionIds, p.Id)
					break
				}
			}
		}
	}

	return server, nil

}

// registerPartitionService partition/[id]:[body]
func (ms *masterService) registerPartitionService(ctx context.Context, partition *entity.Partition) error {
	log.Info("rigister parttion:[%d] ", partition.Id)
	marshal, err := json.Marshal(partition)
	if err != nil {
		return err
	}
	return ms.Master().Put(ctx, entity.PartitionKey(partition.Id), marshal)
}

// createDBService three keys "db/id/[dbId]:[dbName]" ,"db/name/[dbName]:[dbId]" ,"db/body/[dbId]:[dbBody]"
func (ms *masterService) createDBService(ctx context.Context, db *entity.DB) (err error) {
	//validate name has in db is in return err
	if err = db.Validate(); err != nil {
		return err
	}

	if err = ms.validatePS(ctx, db.Ps); err != nil {
		return err
	}

	//generate a new db id
	db.Id, err = ms.Master().NewIDGenerate(ctx, entity.DBIdSequence, 1, 5*time.Second)
	if err != nil {
		return err
	}

	err = ms.Master().STM(context.Background(), func(stm concurrency.STM) error {
		idKey, nameKey, bodyKey := ms.Master().DBKeys(db.Id, db.Name)

		if stm.Get(nameKey) != "" {
			return fmt.Errorf("dbname %s is exists", db.Name)
		}

		if stm.Get(idKey) != "" {
			return fmt.Errorf("dbID %d is exists", db.Id)
		}
		value, err := json.Marshal(db)
		if err != nil {
			return err
		}
		stm.Put(nameKey, cast.ToString(db.Id))
		stm.Put(idKey, db.Name)
		stm.Put(bodyKey, string(value))
		return nil
	})
	return err
}

func (ms *masterService) validatePS(ctx context.Context, psList []string) error {
	if len(psList) < 1 {
		return nil
	}
	servers, err := ms.Master().QueryServers(ctx)
	if err != nil {
		return err
	}
	for _, ps := range psList {
		flag := false
		for _, server := range servers {
			if server.Ip == ps {
				flag = true
				if !client.IsLive(server.RpcAddr()) {
					return fmt.Errorf("server:[%s] can not connection", ps)
				}
				break
			}
		}
		if !flag {
			return fmt.Errorf("server:[%s] not found in cluster", ps)
		}
	}
	return nil
}

func (ms *masterService) deleteDBService(ctx context.Context, dbstr string) (err error) {
	db, err := ms.queryDBService(ctx, dbstr)
	if err != nil {
		return err
	}
	//it will local cluster ,to create space
	mutex := ms.Master().NewLock(ctx, entity.PrefixLockCluster, time.Second*300)
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock space err ")
		}
	}()

	spaces, err := ms.Master().QuerySpaces(ctx, db.Id)
	if err != nil {
		return err
	}

	if len(spaces) > 0 {
		return vearchpb.NewError(vearchpb.ErrorEnum_DB_Not_Empty, nil)
	}

	err = ms.Master().STM(context.Background(),
		func(stm concurrency.STM) error {
			idKey, nameKey, bodyKey := ms.Master().DBKeys(db.Id, db.Name)
			stm.Del(idKey)
			stm.Del(nameKey)
			stm.Del(bodyKey)
			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

func (this *masterService) updateDBIpList(ctx context.Context, dbModify *entity.DBModify) (db *entity.DB, err error) {
	// process painc
	defer errutil.CatchError(err)
	var id int64
	db = &entity.DB{}
	if util.IsNum(dbModify.DbName) {
		id = cast.ToInt64(dbModify.DbName)
	} else if id, err = this.Master().QueryDBName2Id(ctx, dbModify.DbName); err != nil {
		return nil, err
	}
	bs, err := this.Master().Get(ctx, entity.DBKeyBody(id))
	errutil.ThrowError(err)
	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_DB_NOTEXISTS, nil)
	}
	err = json.Unmarshal(bs, db)
	errutil.ThrowError(err)
	if dbModify.Method == proto.ConfRemoveNode {
		ps := make([]string, 0, len(db.Ps))
		for _, ip := range db.Ps {
			if ip != dbModify.IpAddr {
				ps = append(ps, ip)
			}
		}
		db.Ps = ps
	} else if dbModify.Method == proto.ConfAddNode {
		exist := false
		for _, ip := range db.Ps {
			if ip == dbModify.IpAddr {
				exist = true
				break
			}
		}
		if !exist {
			db.Ps = append(db.Ps, dbModify.IpAddr)
		}
	} else {
		err = fmt.Errorf("method support addIP:[%d] removeIP:[%d] not support:[%d]",
			proto.ConfAddNode, proto.ConfRemoveNode, dbModify.Method)
		errutil.ThrowError(err)
	}
	log.Debug("db info is %v", db)
	_, _, bodyKey := this.Master().DBKeys(db.Id, db.Name)
	value, err := json.Marshal(db)
	errutil.ThrowError(err)
	err = this.Client.Master().Put(ctx, bodyKey, value)
	return db, err
}

func (ms *masterService) queryDBService(ctx context.Context, dbstr string) (db *entity.DB, err error) {
	var id int64
	db = &entity.DB{}
	if util.IsNum(dbstr) {
		id = cast.ToInt64(dbstr)
	} else if id, err = ms.Master().QueryDBName2Id(ctx, dbstr); err != nil {
		return nil, err
	}

	bs, err := ms.Master().Get(ctx, entity.DBKeyBody(id))

	if err != nil {
		return nil, err
	}

	if bs == nil {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_DB_NOTEXISTS, nil)
	}

	if err := json.Unmarshal(bs, db); err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

// server/[serverAddr]:[serverBody]
// spaceKeys "space/[dbId]/[spaceId]:[spaceBody]"
func (ms *masterService) createSpaceService(ctx context.Context, dbName string, space *entity.Space) (err error) {
	if space.DBId, err = ms.Master().QueryDBName2Id(ctx, dbName); err != nil {
		log.Error("Failed When finding DbId according DbName:%v,And the Error is:%v", dbName, err)
		return err
	}

	//to validate schema
	_, err = mapping.SchemaMap(space.Properties)
	if err != nil {
		log.Error("master service createSpaceService error: %v", err)
		return err
	}

	//it will lock cluster ,to create space
	mutex := ms.Master().NewLock(ctx, "space", time.Second*300)
	if err = mutex.Lock(); err != nil {
		return err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("unlock space err:[%s]", err.Error())
		}
	}()

	//spaces is existed
	if _, err := ms.Master().QuerySpaceByName(ctx, space.DBId, space.Name); err != nil {
		vErr := vearchpb.NewError(0, err)
		if vErr.GetError().Code != vearchpb.ErrorEnum_SPACE_NOTEXISTS {
			return vErr
		}
	} else {
		return vearchpb.NewError(vearchpb.ErrorEnum_DUP_SPACE, nil)
	}

	log.Info("create space, db: %s, spaceName: %s ,space :[%s]", dbName, space.Name, cbjson.ToJsonString(space))

	//find all servers for create space
	servers, err := ms.Master().QueryServers(ctx)
	if err != nil {
		return err
	}

	//generate space id
	spaceID, err := ms.Master().NewIDGenerate(ctx, entity.SpaceIdSequence, 1, 5*time.Second)
	if err != nil {
		return err
	}
	space.Id = spaceID

	width := math.MaxUint32 / space.PartitionNum
	for i := 0; i < space.PartitionNum; i++ {
		partitionID, err := ms.Master().NewIDGenerate(ctx, entity.PartitionIdSequence, 1, 5*time.Second)

		if err != nil {
			return err
		}

		space.Partitions = append(space.Partitions, &entity.Partition{
			Id:      entity.PartitionID(partitionID),
			SpaceId: space.Id,
			DBId:    space.DBId,
			Slot:    entity.SlotID(i * width),
		})
	}

	serverPartitions, err := ms.filterAndSortServer(ctx, space, servers)
	if err != nil {
		return err
	}

	if int(space.ReplicaNum) > len(serverPartitions) {
		return fmt.Errorf("not enough PS , need replica %d but only has %d", int(space.ReplicaNum), len(servers))
	}

	space.Enabled = util.PBool(false)
	defer func() {
		if !(*space.Enabled) { // if space is still not enabled , so to remove it
			if e := ms.Master().Delete(context.Background(), entity.SpaceKey(space.DBId, space.Id)); e != nil {
				log.Error("to delete space err :%s", e.Error())
			}
		}
	}()

	spaceProperties, err := entity.UnmarshalPropertyJSON(space.Properties)
	if err != nil {
		return err
	}

	space.SpaceProperties = spaceProperties

	marshal, err := json.Marshal(space)
	if err != nil {
		return err
	}
	err = ms.Master().Create(ctx, entity.SpaceKey(space.DBId, space.Id), marshal)
	if err != nil {
		return err
	}

	//peak servers for space
	var paddrs [][]string
	for i := 0; i < len(space.Partitions); i++ {
		if addrs, err := ms.generatePartitionsInfo(servers, serverPartitions, space.ReplicaNum, space.Partitions[i]); err != nil {
			return err
		} else {
			paddrs = append(paddrs, addrs)
		}
	}

	var errChain = make(chan error, 1)
	//send create space to every space
	for i := 0; i < len(space.Partitions); i++ {
		go func(addrs []string, partition *entity.Partition) {
			//send request for all server
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("create partition err: %v ", r)
					errChain <- err
					log.Error(err.Error())
				}
			}()
			for _, addr := range addrs {
				if err := client.CreatePartition(addr, space, partition.Id); err != nil {
					err := fmt.Errorf("create partition err: %s ", err.Error())
					errChain <- err
					log.Error(err.Error())
				}
			}
		}(paddrs[i], space.Partitions[i])
	}

	//check all partition is ok
	for i := 0; i < len(space.Partitions); i++ {
		v := 0
		for {
			v++
			select {
			case err := <-errChain:
				return err
			case <-ctx.Done():
				return fmt.Errorf("create space has error")
			default:

			}

			partition, err := ms.Master().QueryPartition(ctx, space.Partitions[i].Id)
			if v%5 == 0 {
				log.Debug("check the partition:%d status ", space.Partitions[i].Id)
			}
			if err != nil && vearchpb.NewError(0, err).GetError().Code != vearchpb.ErrorEnum_PARTITION_NOT_EXIST {
				return err
			}
			if partition == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			break
		}
	}

	space.Enabled = util.PBool(true)

	//update version
	err = ms.updateSpace(ctx, space)
	if err != nil {
		space.Enabled = util.PBool(false)
		return err
	}

	return nil
}

//create partitions for space create
func (ms *masterService) generatePartitionsInfo(servers []*entity.Server, serverPartitions map[int]int, replicaNum uint8, partition *entity.Partition) ([]string, error) {
	addres := make([]string, 0, replicaNum)
	partition.Replicas = make([]entity.NodeID, 0, replicaNum)

	kvList := make([]struct {
		index  int
		length int
	}, 0, len(serverPartitions))

	for k, v := range serverPartitions {
		kvList = append(kvList, struct {
			index  int
			length int
		}{index: k, length: v})
	}

	sort.Slice(kvList, func(i, j int) bool {
		return kvList[i].length < kvList[j].length
	})

	//find addr for all servers
	for _, kv := range kvList {
		addr := servers[kv.index].RpcAddr()
		ID := servers[kv.index].ID

		if !client.IsLive(addr) {
			serverPartitions[kv.index] = kv.length
			continue
		}
		serverPartitions[kv.index] = serverPartitions[kv.index] + 1
		addres = append(addres, addr)
		partition.Replicas = append(partition.Replicas, ID)

		replicaNum--
		if replicaNum <= 0 {
			break
		}
	}

	if replicaNum > 0 {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_MASTER_PS_NOT_ENOUGH_SELECT, fmt.Errorf("need %d but got %d", partition.Replicas, len(addres)))
	}

	return addres, nil
}

func (ms *masterService) filterAndSortServer(ctx context.Context, space *entity.Space, servers []*entity.Server) (map[int]int, error) {

	db, err := ms.queryDBService(ctx, cast.ToString(space.DBId))
	if err != nil {
		return nil, err
	}

	var psMap map[string]bool
	if len(db.Ps) > 0 {
		psMap = make(map[string]bool)
		for _, ps := range db.Ps {
			psMap[ps] = true
		}
	}

	serverPartitions := make(map[int]int)

	spaces, err := ms.Master().QuerySpacesByKey(ctx, entity.PrefixSpace)

	serverIndex := make(map[entity.NodeID]int)

	if psMap == nil { //means only use public
		for i, s := range servers {
			if !s.Private {
				serverPartitions[i] = 0
				serverIndex[s.ID] = i
			}
		}
	} else { // only use define
		for i, s := range servers {
			if psMap[s.Ip] {
				serverPartitions[i] = 0
				serverIndex[s.ID] = i
			}
		}
	}

	for _, space := range spaces {
		for _, partition := range space.Partitions {
			for _, nodeID := range partition.Replicas {
				if index, ok := serverIndex[nodeID]; ok {
					serverPartitions[index] = serverPartitions[index] + 1
				}
			}
		}
	}

	return serverPartitions, nil
}

func (ms *masterService) deleteSpaceService(ctx context.Context, dbName string, spaceName string) error {
	//send delete to partition
	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		return err
	}

	space, err := ms.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		return err
	}
	if space == nil { //if zero it not exists
		return nil
	}

	//delete key
	err = ms.Master().Delete(ctx, entity.SpaceKey(dbId, space.Id))
	if err != nil {
		return err
	}

	for _, p := range space.Partitions {
		for _, replica := range p.Replicas {
			if server, err := ms.Master().QueryServer(ctx, replica); err != nil {
				log.Error("query partition:[%d] for replica:[%s] has err:[%s]", p.Id, replica, err.Error())
			} else {
				if err := client.DeletePartition(server.RpcAddr(), p.Id); err != nil {
					log.Error("delete partition:[%d] for server:[%s] has err:[%s]", p.Id, server.RpcAddr(), err.Error())
				}
			}
		}
	}

	return nil
}

func (ms *masterService) querySpacesService(ctx context.Context, dbName string, spaceName string) ([]*entity.Space, error) {
	//send delete to partition
	dbId, err := ms.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		return nil, err
	}
	return ms.Master().QuerySpaces(ctx, dbId)
}

func (ms *masterService) queryDBs(ctx context.Context) ([]*entity.DB, error) {

	_, bytesArr, err := ms.Master().PrefixScan(ctx, entity.PrefixDataBaseBody)
	if err != nil {
		return nil, err
	}
	dbs := make([]*entity.DB, 0, len(bytesArr))
	for _, bs := range bytesArr {
		db := &entity.DB{}
		if err := json.Unmarshal(bs, db); err != nil {
			log.Error("decode db err: %s,and the bs is:%s", err.Error(), string(bs))
			continue
		}
		dbs = append(dbs, db)
	}

	return dbs, err
}

func (this *masterService) updateSpaceService(ctx context.Context, dbName, spaceName string, temp *entity.Space) (*entity.Space, error) {

	//it will lock cluster ,to create space
	mutex := this.Master().NewLock(ctx, entity.LockSpaceKey(dbName, spaceName), time.Second*300)
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	defer func() {
		if err := mutex.Unlock(); err != nil {
			log.Error("failed to unlock space,the Error is:%v ", err)
		}
	}()

	dbId, err := this.Master().QueryDBName2Id(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to find database id according database name:%v,the Error is:%v ", dbName, err)
	}

	space, err := this.Master().QuerySpaceByName(ctx, dbId, spaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to find space according space name:%v,the Error is:%v ", spaceName, err)
	}

	if space == nil {
		return nil, fmt.Errorf("can not found space by name : %s", spaceName)
	}

	buff := bytes.Buffer{}
	if temp.DBId != 0 && temp.DBId != space.DBId {
		buff.WriteString("db_id not same ")
	}

	if temp.PartitionNum != 0 && temp.PartitionNum != space.PartitionNum {
		buff.WriteString("partition_num  can not change ")
	}
	if temp.ReplicaNum != 0 && temp.ReplicaNum != space.ReplicaNum {
		buff.WriteString("replica_num  can not change ")
	}
	if buff.String() != "" {
		return nil, fmt.Errorf(buff.String())
	}

	if temp.Name != "" {
		space.Name = temp.Name
	}

	if temp.Enabled != nil {
		space.Enabled = temp.Enabled
	}

	if err := space.Validate(); err != nil {
		return nil, err
	}

	space.Version++
	space.Partitions = temp.Partitions

	if temp.Properties != nil && len(temp.Properties) > 0 {

		//parse old space
		oldFieldMap, err := mapping.SchemaMap(space.Properties)
		if err != nil {
			return nil, err
		}

		//parse new space
		newFieldMap, err := mapping.SchemaMap(temp.Properties)
		if err != nil {
			return nil, err
		}

		for k, v := range oldFieldMap {
			if fm, ok := newFieldMap[k]; ok {
				if !mapping.Equals(v, fm) {
					return nil, fmt.Errorf("not equals by field:[%s] old[%v] new[%v]", k, v, fm)
				}
				delete(newFieldMap, k)
			}
		}

		if len(newFieldMap) > 0 {
			log.Info("change schema for space: %s , change fields : %d , value is :[%s]", space.Name, len(newFieldMap), string(temp.Properties))

			schema, err := mapping.MergeSchema(space.Properties, temp.Properties)
			if err != nil {
				return nil, err
			}

			space.Properties = schema
		}

	}

	// notify all partitions
	for _, p := range space.Partitions {
		partition, err := this.Master().QueryPartition(ctx, p.Id)
		if err != nil {
			return nil, err
		}

		server, err := this.Master().QueryServer(ctx, partition.LeaderID)
		if err != nil {
			return nil, err
		}

		if !client.IsLive(server.RpcAddr()) {
			return nil, fmt.Errorf("partition %s is shutdown", server.RpcAddr())
		}
	}

	for _, p := range space.Partitions {
		partition, err := this.Master().QueryPartition(ctx, p.Id)
		if err != nil {
			return nil, err
		}

		server, err := this.Master().QueryServer(ctx, partition.LeaderID)
		if err != nil {
			return nil, err
		}

		if err := client.UpdatePartition(server.RpcAddr(), space, p.Id); err != nil {
			return nil, err
		}
	}

	space.Version--
	if err := this.updateSpace(ctx, space); err != nil {
		return nil, err
	} else {
		return space, nil
	}
}

func (ms *masterService) updateSpace(ctx context.Context, space *entity.Space) error {
	space.Version++
	space.PartitionNum = len(space.Partitions)
	marshal, err := json.Marshal(space)
	if err != nil {
		return err
	}
	if err = ms.Master().Update(ctx, entity.SpaceKey(space.DBId, space.Id), marshal); err != nil {
		return err
	}

	return nil
}

func (ms *masterService) ChangeMember(ctx context.Context, cm *entity.ChangeMember) error {
	partition, err := ms.Master().QueryPartition(ctx, cm.PartitionID)
	if err != nil {
		return err
	}

	space, err := ms.Master().QuerySpaceByID(ctx, partition.DBId, partition.SpaceId)
	if err != nil {
		return err
	}

	dbName, err := ms.Master().QueryDBId2Name(ctx, space.DBId)
	if err != nil {
		return err
	}

	spacePartition := space.GetPartition(cm.PartitionID)

	if cm.Method != 1 {
		for _, nodeID := range spacePartition.Replicas {
			if nodeID == cm.NodeID {
				return fmt.Errorf("partition:[%d] already has ms server:[%d] in replicas:[%v]", cm.PartitionID, cm.NodeID, spacePartition.Replicas)
			}
		}
		spacePartition.Replicas = append(spacePartition.Replicas, cm.NodeID)
	} else {
		tempIDs := make([]entity.NodeID, 0, len(spacePartition.Replicas)-1)

		for _, id := range spacePartition.Replicas {
			if id != cm.NodeID {
				tempIDs = append(tempIDs, id)
			}
		}
		spacePartition.Replicas = tempIDs
	}

	masterNode, err := ms.Master().QueryServer(ctx, partition.LeaderID)
	if err != nil {
		return err
	}

	var targetNode *entity.Server
	targetNode, err = ms.Master().QueryServer(ctx, cm.NodeID)
	if err != nil {
		if cm.Method == proto.ConfRemoveNode {
			// maybe node is crush
			targetNode = nil
		} else {
			return err
		}
	}

	if !client.IsLive(masterNode.RpcAddr()) {
		return fmt.Errorf("server:[%d] addr:[%s] can not connect ", cm.NodeID, masterNode.RpcAddr())
	}

	if _, err := ms.updateSpaceService(ctx, dbName, space.Name, space); err != nil {
		return err
	}

	if cm.Method == proto.ConfAddNode && targetNode != nil {
		if err := client.CreatePartition(targetNode.RpcAddr(), space, cm.PartitionID); err != nil {
			return fmt.Errorf("create partiiton has err:[%s] addr:[%s]", err.Error(), targetNode.RpcAddr())
		}
	} else if cm.Method == proto.ConfRemoveNode {

	} else {
		return fmt.Errorf("change member only support addNode:[%d] removeNode:[%d] not support:[%d]", proto.ConfAddNode, proto.ConfRemoveNode, cm.Method)
	}

	if err := client.ChangeMember(masterNode.RpcAddr(), cm); err != nil {
		return err
	}
	if cm.Method == proto.ConfRemoveNode && targetNode != nil && client.IsLive(targetNode.RpcAddr()) {
		if err := client.DeleteReplica(targetNode.RpcAddr(), cm.PartitionID); err != nil {
			return fmt.Errorf("create partiiton has err:[%s] addr:[%s]", err.Error(), targetNode.RpcAddr())
		}
	}
	return nil
}

// recover fail node
func (ms *masterService) RecoverFailServer(ctx context.Context, rs *entity.RecoverFailServer) (e error) {
	// painc process
	defer func() {
		if r := recover(); r != nil {
			e = fmt.Errorf("panic is %v", r)
		}
	}()
	// get all fail server info
	failServer, err := ms.Master().QueryAllFailServer(ctx)
	if err != nil {
		panic(err)
	}
	targetFailServer := &entity.FailServer{}
	// get target fail server
	for _, fs := range failServer {
		if fs.Node.Ip == rs.FailNodeAddr {
			targetFailServer = fs
			break
		}
	}
	log.Debug("targetFailServer is %s ", targetFailServer)
	//get target new server
	servers, err := ms.Master().QueryServers(ctx)
	newServer := &entity.Server{}
	for _, s := range servers {
		if s.Ip == rs.NewNodeAddr {
			newServer = s
		}
	}
	log.Debug("newServer is %s ", newServer)
	if newServer.ID > 0 && targetFailServer.ID > 0 {
		for _, pid := range targetFailServer.Node.PartitionIds {
			cm := &entity.ChangeMember{}
			cm.Method = proto.ConfAddNode
			cm.NodeID = newServer.ID
			cm.PartitionID = pid
			if err = ms.ChangeMember(ctx, cm); err != nil {
				info := fmt.Sprintf("ChangePartitionMember failed,partitionId is %d ,nodeId is %d,err is %s ",
					cm.PartitionID, cm.NodeID, err.Error())
				log.Error(info)
				panic(fmt.Errorf(info))
			}
			log.Info("ChangePartitionMember success,partitionId is %d ,nodeId is %d,", cm.PartitionID, cm.NodeID)
		}
		//if success,remove from failServer
		key := entity.FailServerKey(targetFailServer.ID)
		//delete failserver record
		e = ms.Master().Delete(ctx.(context.Context), key)
		log.Debug("Delete failserver key %s .", key)
	} else {
		return fmt.Errorf("newServer or targetFailServer is nil ")
	}
	return e
}
