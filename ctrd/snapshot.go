package ctrd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/snapshots"
	"github.com/opencontainers/image-spec/identity"
)

const defaultSnapshotterName = "overlayfs"

// CreateSnapshot creates a active snapshot with image's name and id.
func (c *Client) CreateSnapshot(ctx context.Context, id, ref string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}
	ctx = leases.WithLease(ctx, wrapperCli.lease.ID())

	image, err := wrapperCli.client.ImageService().Get(ctx, ref)
	if err != nil {
		return err
	}

	diffIDs, err := image.RootFS(ctx, wrapperCli.client.ContentStore(), platforms.Default())
	if err != nil {
		return err
	}

	parent := identity.ChainID(diffIDs).String()
	if _, err := wrapperCli.client.SnapshotService(defaultSnapshotterName).Prepare(ctx, id, parent); err != nil {
		return err
	}
	return nil
}

// GetSnapshot returns the snapshot's info by id.
func (c *Client) GetSnapshot(ctx context.Context, id string) (snapshots.Info, error) {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return snapshots.Info{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(defaultSnapshotterName)
	defer service.Close()

	return service.Stat(ctx, id)
}

// RemoveSnapshot removes the snapshot by id.
func (c *Client) RemoveSnapshot(ctx context.Context, id string) error {
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	service := wrapperCli.client.SnapshotService(defaultSnapshotterName)
	defer service.Close()

	return service.Remove(ctx, id)
}
