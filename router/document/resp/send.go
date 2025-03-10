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

package resp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/vearch/vearch/util/netutil"
)

func SendError(ctx context.Context, w http.ResponseWriter, httpStatus int, errorMsg string) {
	netutil.NewResponse(w).SetHttpStatus(httpStatus).SendJson(NewBody(errorMsg, httpStatus))
}

func SendErrorMethodNotAllowed(ctx context.Context, w http.ResponseWriter, url string, method string, allowMethod string) {
	err := fmt.Errorf(ErrReasonIncorrectHttpMethod, url, method, allowMethod)
	netutil.NewResponse(w).SetHttpStatus(http.StatusMethodNotAllowed).SetAllowMethod(http.MethodPost).SendJson(NewBody(err.Error(), http.StatusMethodNotAllowed))
}

func SendErrorRootCause(ctx context.Context, w http.ResponseWriter, httpStatus int, errorType string, errorReason string) {
	netutil.NewResponse(w).SetHttpStatus(httpStatus).SendJson(NewBodyRootCause(errorType, errorReason, httpStatus))
}

func SendText(ctx context.Context, w http.ResponseWriter, resp string) {
	netutil.NewResponse(w).SetHttpStatus(http.StatusOK).SendText(resp)
}

func SendJsonBytes(ctx context.Context, w http.ResponseWriter, bytes []byte) {
	netutil.NewResponse(w).SetHttpStatus(http.StatusOK).SendJsonBytes(bytes)
}

func SendJson(ctx context.Context, w http.ResponseWriter, obj interface{}) {
	netutil.NewResponse(w).SetHttpStatus(http.StatusOK).SendJson(obj)
}

func SendJsonHttpReplySuccess(ctx context.Context, w http.ResponseWriter, obj interface{}) {
	netutil.NewResponse(w).SendJsonHttpReplySuccess(obj)
}
