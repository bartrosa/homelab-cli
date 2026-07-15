package bootstrap

import (
	"context"
	"fmt"
	"io"
	"os/user"

	"github.com/bartrosa/homelab-cli/internal/exec"
)

// InstallDocker sets up Docker CE on Ubuntu via official apt repository.
func InstallDocker(ctx context.Context, runner exec.Runner, stdout, _ io.Writer) error {
	if _, err := runner.RunWithOutput(ctx, "docker", "--version"); err == nil {
		fmt.Fprintln(stdout, "docker already installed")
		return addUserToDockerGroup(ctx, runner, stdout)
	}

	script := `set -e
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo ${VERSION_CODENAME}) stable" > /etc/apt/sources.list.d/docker.list
apt-get update
apt-get install -y --no-install-recommends docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
`
	if err := runner.Run(ctx, "sudo", "bash", "-c", script); err != nil {
		return fmt.Errorf("docker install: %w", err)
	}
	return addUserToDockerGroup(ctx, runner, stdout)
}

func addUserToDockerGroup(ctx context.Context, runner exec.Runner, stdout io.Writer) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	if err := runner.Run(ctx, "sudo", "usermod", "-aG", "docker", u.Username); err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Added user to docker group — re-login or run: newgrp docker")
	return nil
}
