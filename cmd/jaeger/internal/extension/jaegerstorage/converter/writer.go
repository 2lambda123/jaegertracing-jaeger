// Copyright (c) 2024 The Jaeger Authors.
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"context"
	"errors"
	"fmt"

	otlp2jaeger "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger"
	"go.opentelemetry.io/collector/pdata/ptrace"

	spanstore_v1 "github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/jaegertracing/jaeger/storage_v2/spanstore"
)

type TraceWriter struct {
	spanWriter spanstore_v1.Writer
}

func NewTraceWriter(spanWriter spanstore_v1.Writer) (spanstore.Writer, error) {
	return &TraceWriter{
		spanWriter: spanWriter,
	}, nil
}

// WriteTraces implements spanstore.Writer.
func (t *TraceWriter) WriteTraces(ctx context.Context, td ptrace.Traces) error {
	batches, err := otlp2jaeger.ProtoFromTraces(td)
	if err != nil {
		return fmt.Errorf("cannot transform OTLP traces to Jaeger format: %w", err)
	}
	var errs []error
	for _, batch := range batches {
		for _, span := range batch.Spans {
			if span.Process == nil {
				span.Process = batch.Process
			}
			errs = append(errs, t.spanWriter.WriteSpan(ctx, span))
		}
	}
	return errors.Join(errs...)
}