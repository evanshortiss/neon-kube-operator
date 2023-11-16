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

// BranchSpec defines the desired state of Branch
type BranchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ProjectId        string  `json:"projectId"`
	ParentId         *string `json:"parentId,omitempty"`
	ParentStartPoint *Parent `json:"parentStartPoint,omitempty"`
}

// +kubebuilder:validation:MaxProperties=1
type Parent struct {
	Lsn       *string `json:"lsn,omitempty"`
	Timestamp *string `json:"timestamp,omitempty"`
}

// BranchStatus defines the observed state of Branch
type BranchStatus struct {
	State     BranchState `json:"state"`
	Message   string      `json:"message"`
	Id        string      `json:"id"`
	Name      string      `json:"name"`
	ProjectId string      `json:"projectId"`
	ParentId  string      `json:"parentId"`
	ParentLsn string      `json:"parentLsn"`
	Primary   bool        `json:"primary"`
	CreatedAt string      `json:"createdAt"`
	UpdatedAt string      `json:"updateAt"`
}

func (bs *BranchStatus) Reset() {
	bs.Message = ""
}

type BranchState string

const (
	BranchStateCreating BranchState = "creating"
	BranchStateCreated  BranchState = "created"
	BranchStateDeleting BranchState = "deleting"
)

func (b BranchState) Exists() bool {
	return b == BranchStateCreated || b == BranchStateDeleting
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Branch is the Schema for the branches API
type Branch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BranchSpec   `json:"spec,omitempty"`
	Status BranchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BranchList contains a list of Branch
type BranchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Branch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Branch{}, &BranchList{})
}
