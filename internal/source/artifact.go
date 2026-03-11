package source

import fluxmeta "github.com/fluxcd/pkg/apis/meta"

// ArtifactRef is a small internal wrapper around Flux artifact metadata.
type ArtifactRef struct {
	Artifact *fluxmeta.Artifact
}
