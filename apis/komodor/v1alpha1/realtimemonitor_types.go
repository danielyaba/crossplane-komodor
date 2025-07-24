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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// RealtimeMonitorParameters are the configurable fields of a RealtimeMonitor.
type RealtimeMonitorParameters struct {
	// Name of the monitor.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Sensors configuration.
	// +kubebuilder:validation:Required
	Sensors []apiextensionsv1.JSON `json:"sensors"`

	// Sinks configuration.
	// +kubebuilder:validation:Required
	Sinks apiextensionsv1.JSON `json:"sinks"`

	// Whether the monitor is active.
	// +kubebuilder:validation:Required
	// +kubebuilder:default=true
	Active bool `json:"active"`

	// Type of the monitor.
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// Optional variables.
	// +kubebuilder:validation:Optional
	Variables apiextensionsv1.JSON `json:"variables,omitempty"`

	// Optional sinks options.
	// +kubebuilder:validation:Optional
	SinksOptions map[string][]string `json:"sinksOptions,omitempty"`
}

// RealtimeMonitorObservation are the observable fields of a RealtimeMonitor.
type RealtimeMonitorObservation struct {
	ID           string                 `json:"id,omitempty"`
	CreatedAt    string                 `json:"createdAt,omitempty"`
	UpdatedAt    string                 `json:"updatedAt,omitempty"`
	IsDeleted    bool                   `json:"isDeleted,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Sensors      []apiextensionsv1.JSON `json:"sensors,omitempty"`
	Sinks        apiextensionsv1.JSON   `json:"sinks,omitempty"`
	Active       bool                   `json:"active,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Variables    apiextensionsv1.JSON   `json:"variables,omitempty"`
	SinksOptions map[string][]string    `json:"sinksOptions,omitempty"`
}

// A RealtimeMonitorSpec defines the desired state of a RealtimeMonitor.
type RealtimeMonitorSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RealtimeMonitorParameters `json:"forProvider"`
}

// A RealtimeMonitorStatus represents the observed state of a RealtimeMonitor.
type RealtimeMonitorStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RealtimeMonitorObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A RealtimeMonitor is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,komodor}
// +kubebuilder:rbac:groups=komodor.komodor.crossplane.io,resources=realtimemonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=komodor.komodor.crossplane.io,resources=realtimemonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=komodor.crossplane.io,resources=providerconfigs;providerconfigusages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events;secrets,verbs=get;list;watch;create;update;patch
type RealtimeMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RealtimeMonitorSpec   `json:"spec"`
	Status RealtimeMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RealtimeMonitorList contains a list of RealtimeMonitor
type RealtimeMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RealtimeMonitor `json:"items"`
}

// RealtimeMonitor type metadata.
var (
	RealtimeMonitorKind             = reflect.TypeOf(RealtimeMonitor{}).Name()
	RealtimeMonitorGroupKind        = schema.GroupKind{Group: Group, Kind: RealtimeMonitorKind}.String()
	RealtimeMonitorKindAPIVersion   = RealtimeMonitorKind + "." + SchemeGroupVersion.String()
	RealtimeMonitorGroupVersionKind = SchemeGroupVersion.WithKind(RealtimeMonitorKind)
)

func init() {
	SchemeBuilder.Register(&RealtimeMonitor{}, &RealtimeMonitorList{})
}
