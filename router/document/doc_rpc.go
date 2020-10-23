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

package document

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/vearch/vearch/client"
	"github.com/vearch/vearch/config"
	"github.com/vearch/vearch/proto/vearchpb"
	"github.com/vearch/vearch/util/log"
	"google.golang.org/grpc"
)

const defaultTimeOutMs = 10 * 1000

type Request interface {
	GetHead() *vearchpb.RequestHead
}

type RpcHandler struct {
	client     *client.Client
	docService docService
}

func ExportRpcHandler(rpcServer *grpc.Server, client *client.Client) {
	docService := newDocService(client)

	rpcHandler := &RpcHandler{
		client:     client,
		docService: *docService,
	}

	vearchpb.RegisterRouterGRPCServiceServer(rpcServer, rpcHandler)
}

func (handler *RpcHandler) Space(ctx context.Context, req *vearchpb.RequestHead) (reply *vearchpb.Table, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = vearchpb.NewError(vearchpb.ErrorEnum_RECOVER, errors.New(cast.ToString(r)))
		}
	}()
	space, err := handler.client.Space(ctx, req.DbName, req.SpaceName)

	reply = &vearchpb.Table{}
	pkt := vearchpb.FieldType_value[strings.ToUpper(space.Engine.IdType)]
	tmi := &vearchpb.TableMetaInfo{PrimaryKeyType: vearchpb.FieldType(pkt),
		PartitionsNum: int32(space.PartitionNum),
		ReplicasNum:   int32(space.ReplicaNum),
	}
	tmi.FieldMetaInfo = make([]*vearchpb.FieldMetaInfo, 0)
	for name, field := range space.SpaceProperties {
		isIndex := false
		if field.Index != nil && *field.Index {
			isIndex = true
		}
		fmi := &vearchpb.FieldMetaInfo{Name: name,
			DataType: vearchpb.FieldType(field.FieldType),
			IsIndex:  isIndex,
		}
		if fmi.DataType == vearchpb.FieldType_VECTOR {
			storeType := ""
			if field.StoreType != nil {
				storeType = *field.StoreType
			}
			st := vearchpb.VectorMetaInfo_StoreType_value[strings.ToUpper(storeType)]
			sp, _ := field.StoreParam.MarshalJSON()
			fmi.VectorMetaInfo = &vearchpb.VectorMetaInfo{
				Dimension:  int32(field.Dimension),
				StoreType:  vearchpb.VectorMetaInfo_StoreType(st),
				StoreParam: string(sp),
			}
		}
		tmi.FieldMetaInfo = append(tmi.FieldMetaInfo, fmi)
	}
	reply.Name = space.Name
	reply.TableMetaInfo = tmi

	return reply, nil
}

func (handler *RpcHandler) Get(ctx context.Context, req *vearchpb.GetRequest) (reply *vearchpb.GetResponse, err error) {
	defer Cost("get", time.Now())
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.GetResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) Add(ctx context.Context, req *vearchpb.AddRequest) (reply *vearchpb.AddResponse, err error) {
	defer Cost("Add", time.Now())
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.AddResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) Update(ctx context.Context, req *vearchpb.UpdateRequest) (reply *vearchpb.UpdateResponse, err error) {
	defer Cost("Update", time.Now())
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.UpdateResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) Delete(ctx context.Context, req *vearchpb.DeleteRequest) (reply *vearchpb.DeleteResponse, err error) {
	defer Cost("Delete", time.Now())
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.DeleteResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) Search(ctx context.Context, req *vearchpb.SearchRequest) (reply *vearchpb.SearchResponse, err error) {
	defer Cost("Search", time.Now())
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.SearchResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) MSearch(ctx context.Context, req *vearchpb.MSearchRequest) (reply *vearchpb.SearchResponse, err error) {
	defer Cost("MSearch", time.Now())
	if req.SearchRequests == nil || len(req.SearchRequests) < 1 {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_RPC_PARAM_ERROR, fmt.Errorf("SearchRequests can not be nil or length smaller than 1"))
	}
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.SearchResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) Bulk(ctx context.Context, req *vearchpb.BulkRequest) (reply *vearchpb.BulkResponse, err error) {
	defer Cost("bulk", time.Now())
	res, err := handler.deal(ctx, req)
	if err != nil {
		return nil, err
	}
	reply, ok := res.(*vearchpb.BulkResponse)
	if !ok {
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_INTERNAL_ERROR, nil)
	}
	return reply, nil
}

func (handler *RpcHandler) deal(ctx context.Context, req Request) (reply interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = vearchpb.NewError(vearchpb.ErrorEnum_RECOVER, errors.New(cast.ToString(r)))
		}
	}()
	ctx, cancel := handler.setTimeout(ctx, req.GetHead())
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	switch v := req.(type) {
	case *vearchpb.GetRequest:
		reply = handler.docService.getDocs(ctx, v)
	case *vearchpb.AddRequest:
		reply = handler.docService.addDoc(ctx, v)
	case *vearchpb.UpdateRequest:
		reply = handler.docService.updateDoc(ctx, v)
	case *vearchpb.DeleteRequest:
		reply = handler.docService.deleteDocs(ctx, v)
	case *vearchpb.BulkRequest:
		reply = handler.docService.bulk(ctx, v)
	case *vearchpb.SearchRequest:
		reply = handler.docService.search(ctx, v)
	case *vearchpb.MSearchRequest:
		reply = handler.docService.bulkSearch(ctx, v.SearchRequests)
	default:
		return nil, vearchpb.NewError(vearchpb.ErrorEnum_METHOD_NOT_IMPLEMENT, nil)
	}
	return reply, nil
}

// Cost record how long the function use
func Cost(name string, t time.Time) {
	engTime := time.Now()
	log.Info("%s cost: [%v]", name, engTime.Sub(t))
}

func (handler *RpcHandler) setTimeout(ctx context.Context, head *vearchpb.RequestHead) (context.Context, context.CancelFunc) {
	if head.TimeOutMs < 1 || head.TimeOutMs > defaultTimeOutMs {
		head.TimeOutMs = defaultTimeOutMs
	}
	return context.WithTimeout(ctx, time.Duration(head.TimeOutMs)*time.Millisecond)
}

func (handler *RpcHandler) validateAuth(ctx context.Context, head *vearchpb.RequestHead) error {
	if config.Conf().Global.SkipAuth {
		return nil
	}
	return nil
}
