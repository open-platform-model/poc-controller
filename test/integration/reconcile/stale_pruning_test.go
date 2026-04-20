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

var _ = Describe("Reconcile Stale Pruning", func() {
	// TODO (design 2.1): Validates Phase 4 stale-set computation + Phase 6 prune
	// during normal reconcile. Requires: first reconcile producing resources A
	// and B tracked in inventory → render result in second reconcile returns
	// only A → assert B moves to stale set, apply.Prune deletes B from the
	// cluster, inventory contains only A, outcome == AppliedAndPruned.
	It("should prune a resource removed from the render output", func() {
		Skip("TODO: implement stale prune path during normal reconcile")
	})

	// TODO (design 2.2): Validates Phase 6 prune gate when spec.Prune=false.
	// Requires: first reconcile with Prune=false producing A+B → second reconcile
	// renders only A → assert B is in stale set but remains in the cluster,
	// inventory retains B, outcome == Applied (not AppliedAndPruned).
	It("should retain stale resources when prune is disabled", func() {
		Skip("TODO: implement prune=false stale-skip path")
	})

	// TODO (design 2.3): Validates selective pruning via ComputeStaleSet identity
	// comparison across multiple resources. Requires: first reconcile producing
	// A+B+C → second reconcile renders A+C → assert only B is pruned, A and C
	// remain untouched, inventory reflects A+C.
	It("should prune only the removed resource when multiple exist", func() {
		Skip("TODO: implement multi-resource selective prune path")
	})
})
