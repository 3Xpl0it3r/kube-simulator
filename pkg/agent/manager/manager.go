package manager

import (
	"context"

	coreapi "k8s.io/api/core/v1"
)

// Manager represent manager
type Manager interface {
	OnPodAdd(pod *coreapi.Pod) error
	OnPodUpdate(pod *coreapi.Pod) error
	OnPodDelete(pod *coreapi.Pod) error
	OnNodeAdd(node *coreapi.Node) error
	OnNodeUpdate(node *coreapi.Node) error
	OnNodeDelete(node *coreapi.Node) error
	Run(ctx context.Context)
}
