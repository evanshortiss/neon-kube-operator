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
	BranchFrom            BranchFrom        `json:"from"`
	Type                  string            `json:"type"`
	RegionId              *string           `json:"regionId,omitempty"`
	IncludeCredentials    bool              `json:"includeCredentials,omitempty"`
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

type BranchFrom struct {
	BranchRef string `json:"branchRef,omitempty"`
	ProjectId string `json:"projectId,omitempty"`
	BranchId  string `json:"branchId,omitempty"`
}

type EndpointType string

const (
	EndpointTypeReadOnly  EndpointType = "read_only"
	EndpointTypeReadWrite EndpointType = "read_write"
)

// EndpointStatus defines the observed state of Endpoint
type EndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	State        EndpointState `json:"state"`
	Message      string        `json:"message,omitempty"`
	Id           string        `json:"id"`
	Host         string        `json:"host"`
	CurrentState string        `json:"currentState"`
	PendingState string        `json:"pendingState"`
	CreatedAt    string        `json:"createdAt"`
	UpdatedAt    string        `json:"updateAt"`
}

func (es *EndpointStatus) Reset() {
	es.Message = ""
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
