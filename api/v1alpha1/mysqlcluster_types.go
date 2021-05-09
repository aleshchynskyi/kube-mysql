/*
Copyright 2021.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MysqlClusterSpecConfig struct {
	Name string           `json:"name,omitempty"`
	Spec *MysqlConfigSpec `json:"spec,omitempty"`
}

type MysqlClusterStorage struct {
	VolumeSource *v1.VolumeSource         `json:"volumeSource,omitempty"`
	Resources    *v1.ResourceRequirements `json:"resources,omitempty"`
}

// MysqlClusterSpec defines the desired state of MysqlCluster
type MysqlClusterSpec struct {
	Config   MysqlClusterSpecConfig `json:"config,omitempty"`
	Replicas *int32                 `json:"replicas,omitempty"`
	Storage  MysqlClusterStorage    `json:"storage,omitempty"`
}

// MysqlClusterStatus defines the observed state of MysqlCluster
type MysqlClusterStatus struct {
	ConfigSpec          MysqlConfigSpec `json:"configSpec,omitempty"`
	StatefulSet         string          `json:"statefulSet,omitempty"`
	Service             string          `json:"service,omitempty"`
	StorageVolumeSource v1.VolumeSource `json:"StorageVolumeSource,omitempty"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MysqlCluster is the Schema for the mysqlclusters API
type MysqlCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlClusterSpec   `json:"spec,omitempty"`
	Status MysqlClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlClusterList contains a list of MysqlCluster
type MysqlClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlCluster{}, &MysqlClusterList{})
}
