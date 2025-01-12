// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opencensus // import "go.opentelemetry.io/otel/bridge/opencensus"

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ocmetricdata "go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricproducer"
	ocresource "go.opencensus.io/resource"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/metric/metricdata/metricdatatest"
)

func TestMetricProducer(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		desc      string
		input     []*ocmetricdata.Metric
		expected  []metricdata.ScopeMetrics
		expectErr bool
	}{
		{
			desc:     "empty",
			expected: nil,
		},
		{
			desc: "success",
			input: []*ocmetricdata.Metric{
				{
					Resource: &ocresource.Resource{
						Labels: map[string]string{
							"R1": "V1",
							"R2": "V2",
						},
					},
					TimeSeries: []*ocmetricdata.TimeSeries{
						{
							StartTime: now,
							Points: []ocmetricdata.Point{
								{Value: int64(123), Time: now},
							},
						},
					},
				},
			},
			expected: []metricdata.ScopeMetrics{{
				Scope: instrumentation.Scope{
					Name: scopeName,
				},
				Metrics: []metricdata.Metrics{
					{
						Data: metricdata.Gauge[int64]{
							DataPoints: []metricdata.DataPoint[int64]{
								{
									Attributes: attribute.NewSet(),
									StartTime:  now,
									Time:       now,
									Value:      123,
								},
							},
						},
					},
				},
			}},
		},
		{
			desc: "partial success",
			input: []*ocmetricdata.Metric{
				{
					Descriptor: ocmetricdata.Descriptor{
						Name:        "foo.com/bad-point",
						Description: "a bad type",
						Unit:        ocmetricdata.UnitDimensionless,
						Type:        ocmetricdata.TypeGaugeDistribution,
					},
				},
				{
					Resource: &ocresource.Resource{
						Labels: map[string]string{
							"R1": "V1",
							"R2": "V2",
						},
					},
					TimeSeries: []*ocmetricdata.TimeSeries{
						{
							StartTime: now,
							Points: []ocmetricdata.Point{
								{Value: int64(123), Time: now},
							},
						},
					},
				},
			},
			expected: []metricdata.ScopeMetrics{{
				Scope: instrumentation.Scope{
					Name: scopeName,
				},
				Metrics: []metricdata.Metrics{
					{
						Data: metricdata.Gauge[int64]{
							DataPoints: []metricdata.DataPoint[int64]{
								{
									Attributes: attribute.NewSet(),
									StartTime:  now,
									Time:       now,
									Value:      123,
								},
							},
						},
					},
				},
			}},
			expectErr: true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			fakeProducer := &fakeOCProducer{metrics: tc.input}
			metricproducer.GlobalManager().AddProducer(fakeProducer)
			defer metricproducer.GlobalManager().DeleteProducer(fakeProducer)
			output, err := NewMetricProducer().Produce(context.Background())
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, len(output), len(tc.expected))
			for i := range output {
				metricdatatest.AssertEqual(t, tc.expected[i], output[i])
			}
		})
	}
}

type fakeOCProducer struct {
	metrics []*ocmetricdata.Metric
}

func (f *fakeOCProducer) Read() []*ocmetricdata.Metric {
	return f.metrics
}
