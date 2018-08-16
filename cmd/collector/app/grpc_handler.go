// Copyright (c) 2018 The Jaeger Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/model"
	"github.com/jaegertracing/jaeger/proto-gen/api_v2"
)

// GRPCHandler implements gRPC CollectorService.
type GRPCHandler struct {
	logger        *zap.Logger
	spanProcessor SpanProcessor
}

// NewGRPCHandler registers routes for this handler on the given router
func NewGRPCHandler(logger *zap.Logger, spanProcessor SpanProcessor) *GRPCHandler {
	return &GRPCHandler{
		logger:        logger,
		spanProcessor: spanProcessor,
	}
}

// PostSpans implements gRPC CollectorService.
func (g *GRPCHandler) PostSpans(ctx context.Context, r *api_v2.PostSpansRequest) (*api_v2.PostSpansResponse, error) {
	// TODO
	fmt.Printf("PostSpans(%+v)\n", *r)
	for _, s := range r.Batch.Spans {
		println(s.OperationName)
	}
	return &api_v2.PostSpansResponse{Ok: true}, nil
}

// GetTrace gets trace
func (g *GRPCHandler) GetTrace(ctx context.Context, req *api_v2.GetTraceRequest) (*api_v2.GetTraceResponse, error) {
	return &api_v2.GetTraceResponse{
		Trace: &model.Trace{
			Spans: []*model.Span{
				{
					TraceID:       model.TraceID{Low: 123},
					SpanID:        model.NewSpanID(456),
					OperationName: "foo bar",
					StartTime:     time.Now(),
				},
			},
		},
	}, nil
}
