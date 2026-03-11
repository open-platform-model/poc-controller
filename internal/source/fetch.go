package source

import "context"

// Fetcher fetches a resolved source artifact into a local directory.
type Fetcher interface {
	Fetch(ctx context.Context, artifactURL, artifactDigest, dir string) error
}
