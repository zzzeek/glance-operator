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

package v1beta1

import (
	"fmt"

	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DeploymentHash hash used to detect changes
	DeploymentHash = "deployment"
)

// GlanceAPISpec defines the desired state of GlanceAPI
type GlanceAPISpec struct {

	// Input parameter coming from glance template
	GlanceAPITemplate `json:",inline"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=internal;external;single
	// +kubebuilder:default=external
	APIType string `json:"apiType"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=glance
	// ServiceUser - optional username used for this service to register in glance
	ServiceUser string `json:"serviceUser"`

	// +kubebuilder:validation:Required
	// ServiceAccount - service account name used internally to provide GlanceAPI the default SA name
	ServiceAccount string `json:"serviceAccount"`

	// +kubebuilder:validation:Required
	// DatabaseHostname - Glance Database Hostname
	DatabaseHostname string `json:"databaseHostname"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=glance
	// DatabaseAccount - name of MariaDBAccount which will be used to connect.
	// As an interim solution, glance-controller will
	// generate the MariaDBAccount object of the given name if it's not
	// present in the cluster.  Later, openstack-operator will perform this
	// generation and also populate this field for the Glance objects it creates
	DatabaseAccount string `json:"databaseAccount"`

	// +kubebuilder:validation:Required
	// Secret containing OpenStack password information for glance AdminPassword
	Secret string `json:"secret"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default={database: GlanceDatabasePassword, service: GlancePassword}
	// PasswordSelectors - Selectors to identify the DB and ServiceUser password from the Secret
	PasswordSelectors PasswordSelector `json:"passwordSelectors"`

	// +kubebuilder:validation:Optional
	// ExtraMounts containing conf files and credentials
	ExtraMounts []GlanceExtraVolMounts `json:"extraMounts,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	// QuotaEnforce if true, per-tenant quotas are enforced according to the
	// registered keystone limits
	Quota bool `json:"quota"`

	// ImageCacheSize, provides the size of the cache that will be reflected in the image_cache_max_size parameter
	// +kubebuilder:default=""
	ImageCacheSize string `json:"imageCacheSize"`
}

// GlanceAPIStatus defines the observed state of GlanceAPI
type GlanceAPIStatus struct {
	// ReadyCount of glance API instances
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Map of hashes to track e.g. job status
	Hash map[string]string `json:"hash,omitempty"`

	// API endpoint
	APIEndpoints map[string]string `json:"apiEndpoint,omitempty"`

	// Conditions
	Conditions condition.Conditions `json:"conditions,omitempty" optional:"true"`

	// NetworkAttachments status of the deployment pods
	NetworkAttachments map[string][]string `json:"networkAttachments,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="NetworkAttachments",type="string",JSONPath=".status.networkAttachments",description="NetworkAttachments"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[0].status",description="Status"
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[0].message",description="Message"

// GlanceAPI is the Schema for the glanceapis API
type GlanceAPI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlanceAPISpec   `json:"spec,omitempty"`
	Status GlanceAPIStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlanceAPIList contains a list of GlanceAPI
type GlanceAPIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlanceAPI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GlanceAPI{}, &GlanceAPIList{})
}

// GetEndpoint - returns OpenStack endpoint url for type
func (instance GlanceAPI) GetEndpoint(endpointType endpoint.Endpoint) (string, error) {
	if url, found := instance.Status.APIEndpoints[string(endpointType)]; found {
		return url, nil
	}
	return "", fmt.Errorf("%s endpoint not found", string(endpointType))
}

// IsReady - returns true if GlanceAPI is reconciled successfully
func (instance GlanceAPI) IsReady() bool {
	return instance.Status.Conditions.IsTrue(condition.ReadyCondition)
}
