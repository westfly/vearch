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

package ginutil

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vearch/vearch/proto"
	"github.com/vearch/vearch/util/cbjson"
	"github.com/vearch/vearch/util/netutil"
	"net/http"
)

type Response struct {
	ginContext *gin.Context
	httpStatus int64
}

func New(ginContext *gin.Context) *Response {
	return &Response{
		ginContext: ginContext,
		httpStatus: http.StatusOK,
	}
}

func NewAutoMehtodName(ginContext *gin.Context) *Response {
	response := &Response{
		ginContext: ginContext,
		httpStatus: http.StatusOK,
	}
	return response
}

/*
 default status is 200
*/
func (this *Response) SetHttpStatus(httpStatus int64) *Response {
	this.httpStatus = httpStatus
	return this
}

func (this *Response) SendJson(data interface{}) {
	reply, err := cbjson.Marshal(data)
	if err != nil {
		this.SendJsonHttpReplyError(err)
		return
	}
	this.ginContext.Data(int(this.httpStatus), "application/json", reply)
}

func (this *Response) SendJsonHttpReplySuccess(data interface{}) {
	httpReply := &netutil.HttpReply{
		Code: pkg.ERRCODE_SUCCESS,
		Msg:  pkg.SUCCESS,
		Data: data,
	}
	this.SetHttpStatus(httpReply.Code)
	this.SendJson(httpReply)
}

func (this *Response) SendJsonHttpReplyError(err error) {
	if err == nil {
		err = fmt.Errorf("")
	}

	httpReply := &netutil.HttpReply{
		Code: pkg.ErrCode(err),
		Msg:  err.Error(),
	}
	this.SetHttpStatus(httpReply.Code)
	this.SendJson(httpReply)
}
