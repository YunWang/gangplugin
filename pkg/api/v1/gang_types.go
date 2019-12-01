/*
Copyright 2019 wangyun.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

//All pods and jobs which are managed by gang must have the annotation whose key is GangKey and value is gang's name
const GangKey string = "batch.wangyun.com/gang"
//const GangKey string = "finalizer.batch.wangyun.com"
const FinalizerKey  string = "batch.wangyun.com/gang"

// GangSpec defines the desired state of Gang
type GangSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Gang. Edit Gang_types.go to remove/update
	MinGang int32 `json:"minGang,omitempty"`
}

type GangPhase string

const (
	GangRunningPhase   = "Running"
	GangPendingPhase   = "Pending"
	GangUnknownPhase   = "Unknown"
	GangCompletedPhase = "Completed"
)

// GangStatus defines the observed state of Gang
type GangStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Phase GangPhase `json:"phase,omitempty"`
	// PodPending means the pod has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	PodPending int32 `json:"pending,omitempty"`
	// PodRunning means the pod has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	PodRunning int32 `json:"running,omitempty"`
	// PodSucceeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	PodSucceeded int32 `json:"succeeded,omitempty"`
	// PodFailed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	PodFailed int32 `json:"failed,omitempty"`
	// PodUnknown means that for some reason the state of the pod could not be obtained, typically due
	// to an error in communicating with the host of the pod.
	PodUnknown int32 `json:"unknown,omitempty"`
}

// +kubebuilder:object:root=true

// Gang is the Schema for the gangs API
type Gang struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GangSpec   `json:"spec,omitempty"`
	Status GangStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GangList contains a list of Gang
type GangList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gang `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gang{}, &GangList{})
}
