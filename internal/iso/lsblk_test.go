package iso

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectUmountTargets(t *testing.T) {
	const raw = `{
  "blockdevices": [{
    "name": "sda",
    "mountpoint": null,
    "children": [{
      "name": "sda1",
      "mountpoint": "/media/brosa/UBUNTU 24_0"
    }]
  }]
}`
	var payload lsblkRoot
	require.NoError(t, json.Unmarshal([]byte(raw), &payload))
	var targets []string
	for _, d := range payload.BlockDevices {
		collectUmountTargets(d, &targets)
	}
	require.Equal(t, []string{"/media/brosa/UBUNTU 24_0", "/dev/sda1"}, targets)
}
