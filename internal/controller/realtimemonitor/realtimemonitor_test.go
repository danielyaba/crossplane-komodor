/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package realtimemonitor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-komodor/apis/komodor/v1alpha1"
	komodorclient "github.com/crossplane/provider-komodor/internal/clients/komodor"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

// Mock Komodor client
type mockClient struct {
	getMonitorFn func(ctx context.Context, id string) (*komodorclient.Monitor, error)
}

func (m *mockClient) GetMonitor(ctx context.Context, id string) (*komodorclient.Monitor, error) {
	return m.getMonitorFn(ctx, id)
}

// Implement other methods as no-ops for interface compliance
func (m *mockClient) CreateMonitor(ctx context.Context, monitor *komodorclient.Monitor) (*komodorclient.Monitor, error) {
	return nil, nil
}
func (m *mockClient) UpdateMonitor(ctx context.Context, id string, monitor *komodorclient.Monitor) (*komodorclient.Monitor, error) {
	return nil, nil
}
func (m *mockClient) DeleteMonitor(ctx context.Context, id string) error {
	return nil
}

func TestObserve(t *testing.T) {
	type fields struct {
		client *mockClient
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	// Helper to marshal to apiextensionsv1.JSON
	marshalJSON := func(v interface{}) v1.JSON {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}
		return v1.JSON{Raw: b}
	}

	// In test cases, provide the expected monitor for each scenario
	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ResourceDoesNotExist": {
			reason: "If external name is empty, resource does not exist.",
			fields: fields{client: &mockClient{getMonitorFn: func(ctx context.Context, id string) (*komodorclient.Monitor, error) { return nil, nil }}},
			args: args{
				ctx: context.TODO(),
				mg:  &v1alpha1.RealtimeMonitor{},
			},
			want: want{
				o:   managed.ExternalObservation{ResourceExists: false},
				err: nil,
			},
		},
		"ResourceUpToDate": {
			reason: "If all fields match, resource is up to date.",
			fields: fields{client: &mockClient{getMonitorFn: func(ctx context.Context, id string) (*komodorclient.Monitor, error) {
				return &komodorclient.Monitor{
					ID:           "abc123",
					Name:         "foo",
					Sensors:      []map[string]interface{}{{"a": float64(1)}},
					Sinks:        map[string]interface{}{"b": float64(2)},
					Active:       true,
					Type:         "bar",
					Variables:    map[string]interface{}{"c": float64(3)},
					SinksOptions: map[string][]string{"notifyOn": {"x"}},
				}, nil
			}}},
			args: args{
				ctx: context.TODO(),
				mg: &v1alpha1.RealtimeMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{"crossplane.io/external-name": "abc123"},
					},
					Spec: v1alpha1.RealtimeMonitorSpec{
						ForProvider: v1alpha1.RealtimeMonitorParameters{
							Name:         "foo",
							Sensors:      []v1.JSON{marshalJSON(map[string]interface{}{"a": 1})},
							Sinks:        marshalJSON(map[string]interface{}{"b": 2}),
							Active:       true,
							Type:         "bar",
							Variables:    marshalJSON(map[string]interface{}{"c": 3}),
							SinksOptions: map[string][]string{"notifyOn": {"x"}},
						},
					},
					Status: v1alpha1.RealtimeMonitorStatus{},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				err: nil,
			},
		},
		"ResourceNotUpToDate": {
			reason: "If any field differs, resource is not up to date.",
			fields: fields{client: &mockClient{getMonitorFn: func(ctx context.Context, id string) (*komodorclient.Monitor, error) {
				return &komodorclient.Monitor{
					ID:           "abc123",
					Name:         "foo",
					Sensors:      []map[string]interface{}{{"a": float64(2)}}, // different value
					Sinks:        map[string]interface{}{"b": float64(2)},
					Active:       true,
					Type:         "bar",
					Variables:    map[string]interface{}{"c": float64(3)},
					SinksOptions: map[string][]string{"notifyOn": {"x"}},
				}, nil
			}}},
			args: args{
				ctx: context.TODO(),
				mg: &v1alpha1.RealtimeMonitor{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{"crossplane.io/external-name": "abc123"},
					},
					Spec: v1alpha1.RealtimeMonitorSpec{
						ForProvider: v1alpha1.RealtimeMonitorParameters{
							Name:         "foo",
							Sensors:      []v1.JSON{marshalJSON(map[string]interface{}{"a": 1})},
							Sinks:        marshalJSON(map[string]interface{}{"b": 2}),
							Active:       true,
							Type:         "bar",
							Variables:    marshalJSON(map[string]interface{}{"c": 3}),
							SinksOptions: map[string][]string{"notifyOn": {"x"}},
						},
					},
					Status: v1alpha1.RealtimeMonitorStatus{},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.fields.client}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}
