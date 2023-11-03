package infinityk8s

import "context"

// GetProtectionSetSnapshots exports the private function for testing purposes.
func (a API) GetProtectionSetSnapshots(ctx context.Context, fid string) (
	[]string,
	error,
) {
	return a.getProtectionSetSnapshots(ctx, fid)
}
