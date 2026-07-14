package infinityk8s

import "context"

// ProtectionSetSnapshots exports the private function for testing purposes.
func (a API) ProtectionSetSnapshots(ctx context.Context, fid string) (
	[]string,
	error,
) {
	return a.protectionSetSnapshots(ctx, fid)
}
