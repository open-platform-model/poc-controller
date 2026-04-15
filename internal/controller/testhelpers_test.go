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

package controller

import (
	"fmt"

	"cuelang.org/go/cue/cuecontext"

	"github.com/open-platform-model/poc-controller/pkg/provider"
)

// testProvider builds a minimal provider for controller tests.
// Produces a ConfigMap from each component's data.message field.
func testProvider() *provider.Provider {
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
		panic(fmt.Sprintf("compiling test provider: %v", data.Err()))
	}
	return &provider.Provider{
		Metadata: &provider.ProviderMetadata{
			Name:    "kubernetes",
			Version: "0.1.0",
		},
		Data: data,
	}
}
