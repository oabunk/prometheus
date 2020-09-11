// Copyright 2018 The Prometheus Authors
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

package tsdb

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"go.uber.org/atomic"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/util/testutil"
)

func BenchmarkHeadStripeSeriesCreate(b *testing.B) {
	chunkDir, err := ioutil.TempDir("", "chunk_dir")
	testutil.Ok(b, err)
	defer func() {
		testutil.Ok(b, os.RemoveAll(chunkDir))
	}()
	// Put a series, select it. GC it and then access it.
	h, err := NewHead(nil, nil, nil, 1000, chunkDir, nil, DefaultStripeSize, nil)
	testutil.Ok(b, err)
	defer h.Close()

	for i := 0; i < b.N; i++ {
		h.getOrCreate(uint64(i), labels.FromStrings("a", strconv.Itoa(i)))
	}
}

func BenchmarkHeadStripeSeriesCreateParallel(b *testing.B) {
	chunkDir, err := ioutil.TempDir("", "chunk_dir")
	testutil.Ok(b, err)
	defer func() {
		testutil.Ok(b, os.RemoveAll(chunkDir))
	}()
	// Put a series, select it. GC it and then access it.
	h, err := NewHead(nil, nil, nil, 1000, chunkDir, nil, DefaultStripeSize, nil)
	testutil.Ok(b, err)
	defer h.Close()

	var count atomic.Int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := count.Inc()
			h.getOrCreate(uint64(i), labels.FromStrings("a", strconv.Itoa(int(i))))
		}
	})
}

func BenchmarkHeadAppenderCommit(b *testing.B) {
	head, _ := newTestHead(b, 1000000, false)
	testutil.Ok(b, head.Init(0))

	for i := 0; i < b.N; i++ {
		app := head.Appender(context.Background())
		for j := 0; j < 10000; j++ {
			ls := labels.Labels{
				{
					Name:  "test",
					Value: "test",
				},
				{
					Name:  "num1",
					Value: strconv.Itoa(i),
				},
				{
					Name:  "num2",
					Value: strconv.Itoa(j),
				}}
			_, err := app.Add(ls, int64(j), 0)
			testutil.Ok(b, err)
		}
		testutil.Ok(b, app.Commit())
	}
	testutil.Ok(b, head.Close())
}
