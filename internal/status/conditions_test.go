package status

import (
	"testing"

	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	releasesv1alpha1 "github.com/open-platform-model/poc-controller/api/v1alpha1"
)

// Compile-time interface compliance checks.
var _ conditions.Getter = (*releasesv1alpha1.ModuleRelease)(nil)
var _ conditions.Setter = (*releasesv1alpha1.ModuleRelease)(nil)

func newModuleRelease() *releasesv1alpha1.ModuleRelease {
	return &releasesv1alpha1.ModuleRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  "default",
			Generation: 1,
		},
	}
}

func TestMarkReconciling(t *testing.T) {
	obj := newModuleRelease()
	MarkReconciling(obj, SuspendedReason, "starting reconciliation")

	assert.True(t, conditions.IsTrue(obj, ReconcilingCondition))
	assert.True(t, conditions.IsUnknown(obj, ReadyCondition))
	assert.False(t, conditions.Has(obj, StalledCondition))
}

func TestMarkReconciling_RemovesStalled(t *testing.T) {
	obj := newModuleRelease()
	MarkStalled(obj, RenderFailedReason, "render error")
	assert.True(t, conditions.Has(obj, StalledCondition))

	MarkReconciling(obj, SuspendedReason, "retrying")
	assert.False(t, conditions.Has(obj, StalledCondition))
	assert.True(t, conditions.IsTrue(obj, ReconcilingCondition))
}

func TestMarkStalled(t *testing.T) {
	obj := newModuleRelease()
	MarkStalled(obj, RenderFailedReason, "render error")

	assert.True(t, conditions.IsTrue(obj, StalledCondition))
	assert.True(t, conditions.IsFalse(obj, ReadyCondition))
	assert.False(t, conditions.Has(obj, ReconcilingCondition))
}

func TestMarkStalled_RemovesReconciling(t *testing.T) {
	obj := newModuleRelease()
	MarkReconciling(obj, SuspendedReason, "working")
	assert.True(t, conditions.Has(obj, ReconcilingCondition))

	MarkStalled(obj, ApplyFailedReason, "apply error")
	assert.False(t, conditions.Has(obj, ReconcilingCondition))
	assert.True(t, conditions.IsTrue(obj, StalledCondition))
}

func TestMarkReady(t *testing.T) {
	obj := newModuleRelease()
	// Set up pre-existing conditions that should be cleared.
	MarkReconciling(obj, SuspendedReason, "working")

	MarkReady(obj, "all resources applied")

	assert.True(t, conditions.IsTrue(obj, ReadyCondition))
	assert.False(t, conditions.Has(obj, ReconcilingCondition))
	assert.False(t, conditions.Has(obj, StalledCondition))
	assert.Equal(t, ReconciliationSucceededReason, conditions.GetReason(obj, ReadyCondition))
}

func TestMarkNotReady(t *testing.T) {
	obj := newModuleRelease()
	MarkNotReady(obj, ArtifactFetchFailedReason, "fetch failed: timeout")

	assert.True(t, conditions.IsFalse(obj, ReadyCondition))
	assert.Equal(t, ArtifactFetchFailedReason, conditions.GetReason(obj, ReadyCondition))
	assert.Equal(t, "fetch failed: timeout", conditions.GetMessage(obj, ReadyCondition))
}

func TestMarkSourceReady(t *testing.T) {
	obj := newModuleRelease()
	MarkSourceReady(obj, "sha256:abc123")

	assert.True(t, conditions.IsTrue(obj, SourceReadyCondition))
	assert.Contains(t, conditions.GetMessage(obj, SourceReadyCondition), "sha256:abc123")
}

func TestMarkSourceNotReady(t *testing.T) {
	obj := newModuleRelease()
	MarkSourceNotReady(obj, SourceUnavailableReason, "source not found")

	assert.True(t, conditions.IsFalse(obj, SourceReadyCondition))
	assert.Equal(t, SourceUnavailableReason, conditions.GetReason(obj, SourceReadyCondition))
}

func TestMarkSourceReady_OverwritesNotReady(t *testing.T) {
	obj := newModuleRelease()
	MarkSourceNotReady(obj, SourceNotReadyReason, "pending")
	MarkSourceReady(obj, "sha256:def456")

	assert.True(t, conditions.IsTrue(obj, SourceReadyCondition))
}

func TestConditionConstants(t *testing.T) {
	assert.Equal(t, "Ready", ReadyCondition)
	assert.Equal(t, "Reconciling", ReconcilingCondition)
	assert.Equal(t, "Stalled", StalledCondition)
	assert.Equal(t, "SourceReady", SourceReadyCondition)
}

func TestReasonConstants(t *testing.T) {
	reasons := []string{
		SuspendedReason,
		SourceNotReadyReason,
		SourceUnavailableReason,
		ArtifactFetchFailedReason,
		ArtifactInvalidReason,
		RenderFailedReason,
		ApplyFailedReason,
		PruneFailedReason,
		ReconciliationSucceededReason,
	}
	for _, r := range reasons {
		assert.NotEmpty(t, r, "reason constant should not be empty")
	}
}
