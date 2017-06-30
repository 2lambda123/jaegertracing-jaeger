// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package integration

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/uber/jaeger/model"
)

type TraceByTraceID []*model.Trace

func (s TraceByTraceID) Len() int      { return len(s) }
func (s TraceByTraceID) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TraceByTraceID) Less(i, j int) bool {
	if len(s[i].Spans) != len(s[j].Spans) {
		return len(s[i].Spans) < len(s[j].Spans)
	} else if len(s[i].Spans) == 0 {
		return true
	}
	return s[i].Spans[0].TraceID.Low < s[j].Spans[0].TraceID.Low
}

func CompareListOfTraces(t *testing.T, expected []*model.Trace, actual []*model.Trace) {
	sort.Sort(TraceByTraceID(expected))
	sort.Sort(TraceByTraceID(actual))
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		require.NoError(t, sortTraces(expected[i], actual[i]))
	}
	if !assert.EqualValues(t, expected, actual) {
		for _, err := range pretty.Diff(expected, actual) {
			t.Log(err)
		}
		out, err := json.Marshal(actual)
		assert.NoError(t, err)
		t.Logf("Actual traces: %s", string(out))
	}
}

type SpanBySpanID []*model.Span

func (s SpanBySpanID) Len() int           { return len(s) }
func (s SpanBySpanID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SpanBySpanID) Less(i, j int) bool { return s[i].SpanID < s[j].SpanID }

func CompareTraces(t *testing.T, expected *model.Trace, actual *model.Trace) {
	if expected.Spans == nil {
		require.Nil(t, actual.Spans)
		return
	}
	require.NoError(t, sortTraces(expected, actual))
	if !assert.EqualValues(t, expected, actual) {
		for _, err := range pretty.Diff(expected, actual) {
			t.Log(err)
		}
		out, err := json.Marshal(actual)
		assert.NoError(t, err)
		t.Logf("Actual trace: %s", string(out))
	}
}

func sortTraces(expected *model.Trace, actual *model.Trace) error {
	expectedSpans := expected.Spans
	actualSpans := actual.Spans
	if len(expectedSpans) != len(actualSpans) {
		return errors.New("traces have different number of spans")
	}
	sort.Sort(SpanBySpanID(expectedSpans))
	sort.Sort(SpanBySpanID(actualSpans))
	for i := range expectedSpans {
		if err := sortSpans(expectedSpans[i], actualSpans[i]); err != nil {
			return err
		}
	}
	return nil
}

func sortSpans(expected *model.Span, actual *model.Span) error {
	expected.NormalizeTimestamps()
	actual.NormalizeTimestamps()
	if err := sortTags(expected.Tags, actual.Tags); err != nil {
		return err
	}
	if err := sortLogs(expected.Logs, actual.Logs); err != nil {
		return err
	}
	if err := sortProcess(expected.Process, actual.Process); err != nil {
		return err
	}
	return nil
}

type TagByKey []model.KeyValue

func (t TagByKey) Len() int           { return len(t) }
func (t TagByKey) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TagByKey) Less(i, j int) bool { return t[i].Key < t[j].Key }

func sortTags(expected []model.KeyValue, actual []model.KeyValue) error {
	if len(expected) != len(actual) {
		return errors.New("tags have different length")
	}
	sort.Sort(TagByKey(expected))
	sort.Sort(TagByKey(actual))
	return nil
}

type LogByTimestamp []model.Log

func (t LogByTimestamp) Len() int           { return len(t) }
func (t LogByTimestamp) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t LogByTimestamp) Less(i, j int) bool { return t[i].Timestamp.Before(t[j].Timestamp) }

func sortLogs(expected []model.Log, actual []model.Log) error {
	if len(expected) != len(actual) {
		return errors.New("logs have different length")
	}
	sort.Sort(LogByTimestamp(expected))
	sort.Sort(LogByTimestamp(actual))
	for i := range expected {
		sortTags(expected[i].Fields, actual[i].Fields)
	}
	return nil
}

func sortProcess(expected *model.Process, actual *model.Process) error {
	return sortTags(expected.Tags, actual.Tags)
}
