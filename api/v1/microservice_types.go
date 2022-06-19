/*
Copyright 2022.

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
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MicroServiceSpec defines the desired state of MicroService
type MicroServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MicroService. Edit microservice_types.go to remove/update
	Image      string      `json:"image"`
	Config     []v1.EnvVar `json:"config,omitempty"`
	Port       int32       `json:"port,omitempty"`
	StartPoint bool        `json:"startPoint,omitempty"`
}

// MicroServiceStatus defines the observed state of MicroService
type MicroServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PodName string `json:"podName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MicroService is the Schema for the microservices API
type MicroService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MicroServiceSpec   `json:"spec,omitempty"`
	Status MicroServiceStatus `json:"status,omitempty"`
}

func (ms *MicroService) String() string {
	return fmt.Sprintf("Image [%s], Port [%d], Configs [%s], StartPoint [%v], PodName [%s]",
		ms.Spec.Image,
		ms.Spec.Port,
		ms.Spec.Config,
		ms.Spec.StartPoint,
		ms.Status.PodName)
}

//+kubebuilder:object:root=true

// MicroServiceList contains a list of MicroService
type MicroServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MicroService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MicroService{}, &MicroServiceList{})
}
