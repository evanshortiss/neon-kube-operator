/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EndpointSpec defines the desired state of Endpoint
type EndpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProjectId             string            `json:"projectId"`
	BranchId              string            `json:"branchId"`
	RegionId              *string           `json:"regionId,omitempty"`
	Type                  string            `json:"type"`
	Settings              map[string]string `json:"settings,omitempty"`
	AutoscalingLimitMinCu *int              `json:"autoscalingLimitMinCu,omitempty"`
	AutoscalingLimitMaxCu *int              `json:"autoscalingLimitMaxCu,omitempty"`
	Provisioner           *string           `json:"provisioner,omitempty"`
	PoolerEnabled         *bool             `json:"poolerEnabled,omitempty"`
	PoolerMode            *string           `json:"poolerMode,omitempty"`
	Disabled              *bool             `json:"disabled,omitempty"`
	PasswordlessAccess    *bool             `json:"passwordless_access,omitempty"`
	SuspendTimeoutSeconds *int64            `json:"suspendTimeoutSeconds,omitempty"`
}

// EndpointStatus defines the observed state of Endpoint
type EndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State EndpointState `json:"state"`
}

type EndpointState string

const (
	EndpointStateCreating EndpointState = "creating"
	EndpointStateCreated  EndpointState = "created"
	EndpointStateDeleting EndpointState = "deleting"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Endpoint is the Schema for the endpoints API
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointSpec   `json:"spec,omitempty"`
	Status EndpointStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Endpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
