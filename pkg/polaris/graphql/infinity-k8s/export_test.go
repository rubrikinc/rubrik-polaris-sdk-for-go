package infinityk8s

import "context"

// GetResourceSetSnapshots exports the private function for testing purposes.
func (a API) GetResourceSetSnapshots(ctx context.Context, fid string) (
	[]string,
	error,
) {
	return a.getResourceSetSnapshots(ctx, fid)
}
