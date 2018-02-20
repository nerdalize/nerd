package v1

import (
	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dataset describes a nerd dataset.
type Dataset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DatasetSpec `json:"spec"`
}

// DatasetSpec is the spec for a Dataset resource
type DatasetSpec struct {
	StoreOptions    transferstore.StoreOptions
	ArchiverOptions transferarchiver.ArchiverOptions

	Options    map[string]string `json:"options"`
	Size       uint64            `json:"size"`
	InputFor   []string          `json:"input"`
	OutputFrom []string          `json:"output"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DatasetList is a list of Dataset resources
type DatasetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Dataset `json:"items"`
}
