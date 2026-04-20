/*
Copyright 2026.

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

package reconcile_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Reconcile State Recovery", func() {
	// TODO (design 3.1): Validates Stalled → Ready recovery when the missing
	// source finally appears. Requires: first reconcile with missing
	// OCIRepository → assert Stalled=True → create the OCIRepository → second
	// reconcile → assert Ready=True, Stalled condition cleared, resources
	// applied, history reflects the transition.
	It("should recover from Stalled when the source becomes available", func() {
		Skip("TODO: implement Stalled→Ready recovery")
	})

	// TODO (design 3.2): Validates SoftBlocked → Ready recovery when a not-ready
	// source transitions to ready. Requires: first reconcile with OCIRepository
	// present but not ready → assert SourceReady=False, 30s requeue → patch
	// OCIRepository status to ready with an artifact → second reconcile →
	// assert Ready=True, SourceReady=True, resources applied.
	It("should recover from SoftBlocked when the source becomes ready", func() {
		Skip("TODO: implement SoftBlocked→Ready recovery")
	})

	// TODO (design 3.3): Validates suspend → unsuspend resumes full reconcile.
	// Requires: first reconcile with spec.suspend=true → assert early return,
	// no managed resources, Reconciling condition set → patch spec.suspend=false
	// → second reconcile → assert Ready=True, resources applied, stale suspend
	// conditions cleared.
	It("should resume reconciliation when suspend is cleared", func() {
		Skip("TODO: implement suspend→unsuspend recovery")
	})
})
