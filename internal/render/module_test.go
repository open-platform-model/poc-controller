package render

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"cuelang.org/go/cue/cuecontext"

	releasesv1alpha1 "github.com/open-platform-model/poc-controller/api/v1alpha1"
	"github.com/open-platform-model/poc-controller/pkg/core"
	"github.com/open-platform-model/poc-controller/pkg/provider"
)

func testdataDir(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

// testProvider builds a minimal provider with a single transformer that produces
// a ConfigMap from the component's data.message field.
func testProvider(t *testing.T) *provider.Provider {
	t.Helper()
	cueCtx := cuecontext.New()
	data := cueCtx.CompileString(`{
	metadata: {
		name:        "kubernetes"
		description: "Test provider"
		version:     "0.1.0"
	}
	#transformers: {
		"simple": {
			#transform: {
				#component: _
				#context: _
				output: {
					apiVersion: "v1"
					kind:       "ConfigMap"
					metadata: {
						name:      #context.#moduleReleaseMetadata.name
						namespace: #context.#moduleReleaseMetadata.namespace
						labels:    #context.#runtimeLabels
					}
					data: {
						message: #component.data.message
					}
				}
			}
		}
	}
}`)
	if data.Err() != nil {
		t.Fatalf("compiling test provider: %v", data.Err())
	}
	return &provider.Provider{
		Metadata: &provider.ProviderMetadata{
			Name:    "kubernetes",
			Version: "0.1.0",
		},
		Data: data,
	}
}

const testNamespace = "default"

func rawValues(json string) *releasesv1alpha1.RawValues {
	v := &releasesv1alpha1.RawValues{}
	v.Raw = []byte(json)
	return v
}

func TestRenderModule_Success(t *testing.T) {
	result, err := RenderModule(
		context.Background(),
		testdataDir("valid-module"),
		rawValues(`{"message": "hello"}`),
		testProvider(t),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	r := result.Resources[0]
	if r.Component != "web" {
		t.Errorf("expected component 'web', got %q", r.Component)
	}

	u, err := r.ToUnstructured()
	if err != nil {
		t.Fatalf("converting to unstructured: %v", err)
	}
	if u.GetKind() != "ConfigMap" {
		t.Errorf("expected kind ConfigMap, got %q", u.GetKind())
	}
	if u.GetName() != "test-module" {
		t.Errorf("expected name 'test-module', got %q", u.GetName())
	}
	if u.GetNamespace() != testNamespace {
		t.Errorf("expected namespace 'default', got %q", u.GetNamespace())
	}

	data, ok := u.Object["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data field in ConfigMap")
	}
	if data["message"] != "hello" {
		t.Errorf("expected data.message 'hello', got %v", data["message"])
	}
}

func TestRenderModule_NilValues(t *testing.T) {
	result, err := RenderModule(
		context.Background(),
		testdataDir("valid-module"),
		nil,
		testProvider(t),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(result.Resources))
	}

	u, err := result.Resources[0].ToUnstructured()
	if err != nil {
		t.Fatalf("converting to unstructured: %v", err)
	}
	data, ok := u.Object["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data field in ConfigMap")
	}
	if data["message"] != "default-hello" {
		t.Errorf("expected default value 'default-hello', got %v", data["message"])
	}
}

func TestRenderModule_InvalidValues(t *testing.T) {
	_, err := RenderModule(
		context.Background(),
		testdataDir("valid-module"),
		rawValues(`{"count": 5}`),
		testProvider(t),
	)
	if err == nil {
		t.Fatal("expected error for invalid values")
	}
	if !strings.Contains(err.Error(), "parsing module release") {
		t.Errorf("expected error about parsing module release, got: %v", err)
	}
}

func TestRenderModule_InvalidJSON(t *testing.T) {
	_, err := RenderModule(
		context.Background(),
		testdataDir("valid-module"),
		rawValues(`{bad json`),
		testProvider(t),
	)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
	if !strings.Contains(err.Error(), "compiling values") {
		t.Errorf("expected error about compiling values, got: %v", err)
	}
}

func TestRenderModule_NoComponents(t *testing.T) {
	_, err := RenderModule(
		context.Background(),
		testdataDir("no-components-module"),
		nil,
		testProvider(t),
	)
	if err == nil {
		t.Fatal("expected error for module with no components")
	}
	if !strings.Contains(err.Error(), "no resources rendered") {
		t.Errorf("expected error about no resources, got: %v", err)
	}
}

func TestRenderModule_RuntimeLabelsPresent(t *testing.T) {
	result, err := RenderModule(
		context.Background(),
		testdataDir("valid-module"),
		rawValues(`{"message": "hello"}`),
		testProvider(t),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, err := result.Resources[0].ToUnstructured()
	if err != nil {
		t.Fatalf("converting to unstructured: %v", err)
	}
	labels := u.GetLabels()

	if got := labels[core.LabelManagedBy]; got != core.LabelManagedByControllerValue {
		t.Errorf("expected managed-by %q, got %q", core.LabelManagedByControllerValue, got)
	}
	if got := labels[core.LabelModuleReleaseNamespace]; got != testNamespace {
		t.Errorf("expected module-release namespace %q, got %q", testNamespace, got)
	}
}

func TestRenderModule_InventoryEntriesMatchResources(t *testing.T) {
	result, err := RenderModule(
		context.Background(),
		testdataDir("valid-module"),
		rawValues(`{"message": "hello"}`),
		testProvider(t),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.InventoryEntries) != len(result.Resources) {
		t.Fatalf("expected %d inventory entries, got %d",
			len(result.Resources), len(result.InventoryEntries))
	}

	entry := result.InventoryEntries[0]
	if entry.Kind != "ConfigMap" {
		t.Errorf("expected inventory kind ConfigMap, got %q", entry.Kind)
	}
	if entry.Name != "test-module" {
		t.Errorf("expected inventory name 'test-module', got %q", entry.Name)
	}
	if entry.Namespace != testNamespace {
		t.Errorf("expected inventory namespace 'default', got %q", entry.Namespace)
	}
	if entry.Version != "v1" {
		t.Errorf("expected inventory version 'v1', got %q", entry.Version)
	}
}
