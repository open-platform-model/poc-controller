package render

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	releasesv1alpha1 "github.com/open-platform-model/poc-controller/api/v1alpha1"
	"github.com/open-platform-model/poc-controller/internal/inventory"
	"github.com/open-platform-model/poc-controller/pkg/core"
	"github.com/open-platform-model/poc-controller/pkg/loader"
	"github.com/open-platform-model/poc-controller/pkg/module"
	"github.com/open-platform-model/poc-controller/pkg/provider"
	pkgrender "github.com/open-platform-model/poc-controller/pkg/render"
)

// RenderResult holds the output of a successful RenderModule call.
// Contains both the rendered resources and their inventory entries, giving the
// caller everything needed for apply + inventory in one call.
type RenderResult struct {
	// Resources is the ordered list of rendered Kubernetes resources.
	Resources []*core.Resource

	// InventoryEntries are the CRD-typed inventory entries built from Resources.
	InventoryEntries []releasesv1alpha1.InventoryEntry

	// Warnings are non-fatal render warnings (e.g., unhandled traits).
	Warnings []string
}

// RenderModule is the single entry point for CUE module rendering in the controller.
// It encapsulates the full CLI pipeline:
//
//  1. Load the CUE module from the extracted artifact directory.
//  2. Extract module metadata and #config schema.
//  3. Convert RawValues (JSON) to a cue.Value.
//  4. Inject #runtimeLabels with controller identity metadata.
//  5. Build a release via module.ParseModuleRelease.
//  6. Render via render.ProcessModuleRelease with the caller-supplied provider.
//  7. Convert rendered resources to inventory entries.
//
// The provider is loaded from the controller-owned catalog at startup and passed
// through the reconciler. RenderModule does not load or discover providers.
// The values parameter may be nil if the module has no required config or has defaults.
func RenderModule(
	ctx context.Context,
	moduleDir string,
	values *releasesv1alpha1.RawValues,
	prov *provider.Provider,
) (*RenderResult, error) {
	cueCtx := cuecontext.New()

	// Load the CUE module package.
	raw, err := loader.LoadModulePackage(cueCtx, moduleDir)
	if err != nil {
		return nil, fmt.Errorf("loading module package: %w", err)
	}

	// Extract module metadata and #config schema.
	mod, err := extractModule(raw, moduleDir)
	if err != nil {
		return nil, fmt.Errorf("extracting module: %w", err)
	}

	// Convert CRD values to cue.Value slice.
	var cueValues []cue.Value
	if values != nil && values.Raw != nil {
		compiled := cueCtx.CompileBytes(values.Raw, cue.Filename("values"))
		if compiled.Err() != nil {
			return nil, fmt.Errorf("compiling values: %w", compiled.Err())
		}
		cueValues = append(cueValues, compiled)
	}

	// Inject #runtimeLabels before building the release. The definition is
	// available to the CUE module's templates during evaluation.
	runtimeLabels := cueCtx.CompileString(fmt.Sprintf(`{
	%q: %q
}`, core.LabelManagedBy, core.LabelManagedByControllerValue), cue.Filename("runtimeLabels"))
	if runtimeLabels.Err() != nil {
		return nil, fmt.Errorf("compiling runtime labels: %w", runtimeLabels.Err())
	}
	raw = raw.FillPath(cue.ParsePath("#runtimeLabels"), runtimeLabels)

	// Build the release: validate values against #config, fill, ensure concrete.
	rel, err := module.ParseModuleRelease(ctx, raw, mod, cueValues)
	if err != nil {
		return nil, fmt.Errorf("parsing module release: %w", err)
	}

	// Render with the caller-supplied provider, overriding runtime labels
	// so rendered resources carry controller identity instead of CLI identity.
	controllerLabels := map[string]string{
		core.LabelManagedBy:              core.LabelManagedByControllerValue,
		core.LabelModuleReleaseNamespace: rel.Metadata.Namespace,
	}
	result, err := pkgrender.ProcessModuleRelease(ctx, rel, prov, controllerLabels)
	if err != nil {
		return nil, fmt.Errorf("processing module release: %w", err)
	}

	if len(result.Resources) == 0 {
		return nil, fmt.Errorf("module %q: no resources rendered", mod.Metadata.Name)
	}

	// Convert resources to inventory entries.
	entries, err := buildInventoryEntries(result.Resources)
	if err != nil {
		return nil, fmt.Errorf("building inventory entries: %w", err)
	}

	return &RenderResult{
		Resources:        result.Resources,
		InventoryEntries: entries,
		Warnings:         result.Warnings,
	}, nil
}

// extractModule extracts the Module struct from a raw CUE value loaded from disk.
func extractModule(raw cue.Value, moduleDir string) (module.Module, error) {
	metaVal := raw.LookupPath(cue.ParsePath("metadata"))
	if !metaVal.Exists() {
		return module.Module{}, fmt.Errorf("module missing metadata field")
	}

	meta := &module.ModuleMetadata{}
	if err := metaVal.Decode(meta); err != nil {
		return module.Module{}, fmt.Errorf("decoding module metadata: %w", err)
	}

	config := raw.LookupPath(cue.ParsePath("#config"))

	return module.Module{
		Metadata:   meta,
		Config:     config,
		Raw:        raw,
		ModulePath: moduleDir,
	}, nil
}

// buildInventoryEntries converts rendered resources to inventory entries.
func buildInventoryEntries(resources []*core.Resource) ([]releasesv1alpha1.InventoryEntry, error) {
	entries := make([]releasesv1alpha1.InventoryEntry, 0, len(resources))
	for _, r := range resources {
		u, err := r.ToUnstructured()
		if err != nil {
			return nil, fmt.Errorf("converting resource %s to unstructured: %w", r, err)
		}
		entries = append(entries, inventory.NewEntryFromResource(u))
	}
	return entries, nil
}
