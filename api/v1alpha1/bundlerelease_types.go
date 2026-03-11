/*
Copyright 2026.

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
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BundleReleaseSpec defines the desired state of BundleRelease
type BundleReleaseSpec struct {
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	SourceRef SourceReference `json:"sourceRef"`

	// +optional
	Values *RawValues `json:"values,omitempty"`

	// +optional
	Prune bool `json:"prune,omitempty"`

	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// +optional
	DependsOn []fluxmeta.NamespacedObjectReference `json:"dependsOn,omitempty"`
}

// BundleReleaseStatus defines the observed state of BundleRelease.
type BundleReleaseStatus struct {
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// conditions represent the current state of the BundleRelease resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// +optional
	Source *SourceStatus `json:"source,omitempty"`

	// +optional
	LastAttemptedAction string `json:"lastAttemptedAction,omitempty"`

	// +optional
	LastAttemptedAt *metav1.Time `json:"lastAttemptedAt,omitempty"`

	// +optional
	LastAttemptedSourceDigest string `json:"lastAttemptedSourceDigest,omitempty"`

	// +optional
	LastAttemptedConfigDigest string `json:"lastAttemptedConfigDigest,omitempty"`

	// +optional
	LastAttemptedRenderDigest string `json:"lastAttemptedRenderDigest,omitempty"`

	// +optional
	LastAppliedAt *metav1.Time `json:"lastAppliedAt,omitempty"`

	// +optional
	LastAppliedSourceDigest string `json:"lastAppliedSourceDigest,omitempty"`

	// +optional
	LastAppliedConfigDigest string `json:"lastAppliedConfigDigest,omitempty"`

	// +optional
	LastAppliedRenderDigest string `json:"lastAppliedRenderDigest,omitempty"`

	// +optional
	Inventory *Inventory `json:"inventory,omitempty"`

	// +optional
	Modules []ModuleStatusSummary `json:"modules,omitempty"`

	// +optional
	History []HistoryEntry `json:"history,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=br
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="Source",type=string,JSONPath=".status.source.artifactRevision"

// BundleRelease is the Schema for the bundlereleases API
type BundleRelease struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of BundleRelease
	// +required
	Spec BundleReleaseSpec `json:"spec"`

	// status defines the observed state of BundleRelease
	// +optional
	Status BundleReleaseStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// BundleReleaseList contains a list of BundleRelease
type BundleReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []BundleRelease `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BundleRelease{}, &BundleReleaseList{})
}
