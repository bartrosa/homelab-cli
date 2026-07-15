package components

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bartrosa/homelab-cli/internal/stack"
	"github.com/bartrosa/homelab-cli/internal/stack/gpu"
)

type dockerComponent struct{}

func (d *dockerComponent) ID() string { return "docker" }
func (d *dockerComponent) Description() string {
	return "Docker Engine (official docker-ce repo on Ubuntu)"
}
func (d *dockerComponent) DisplayName() string            { return "Docker" }
func (d *dockerComponent) Category() stack.Category       { return stack.CategoryContainer }
func (d *dockerComponent) DefaultVersion() string         { return "latest" }
func (d *dockerComponent) Requires() []string             { return nil }
func (d *dockerComponent) PathEntries() []stack.PathEntry { return nil }

func (d *dockerComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	if !cmdExists(ctx, env, "docker") {
		return false, "", nil
	}
	ver := versionOf(ctx, env, "docker", "compose", "version")
	if ver == "" {
		ver = versionOf(ctx, env, "docker", "--version")
	}
	return true, ver, nil
}

func (d *dockerComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	if env.Info.IsSilverblue && !opts.Force {
		return fmt.Errorf("silverblue: official recommendation is podman; use --force to install Docker anyway")
	}
	script := `set -e
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo tee /etc/apt/keyrings/docker.asc > /dev/null
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker "$USER"
`
	if opts.DryRun {
		return env.Runner.Run(ctx, "sh", "-c", script)
	}
	if err := env.Runner.Run(ctx, "sh", "-c", script); err != nil {
		return err
	}
	fmt.Fprintln(env.Stdout, "Log out and back in for docker group membership to take effect")
	return nil
}

type podmanComponent struct{}

func (p *podmanComponent) ID() string                     { return "podman" }
func (p *podmanComponent) DisplayName() string            { return "Podman" }
func (p *podmanComponent) Category() stack.Category       { return stack.CategoryContainer }
func (p *podmanComponent) Description() string            { return "Podman + podman-compose" }
func (p *podmanComponent) DefaultVersion() string         { return "system" }
func (p *podmanComponent) Requires() []string             { return nil }
func (p *podmanComponent) PathEntries() []stack.PathEntry { return nil }

func (p *podmanComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "podman")
	return ok, versionOf(ctx, env, "podman", "--version"), nil
}

func (p *podmanComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	if env.Info.IsSilverblue {
		fmt.Fprintln(env.Stdout, "podman preinstalled on Silverblue")
		return nil
	}
	return installPkg(ctx, env, "podman", "podman-compose", "podman-docker")
}

type duckdbComponent struct{}

func (d *duckdbComponent) ID() string               { return "duckdb" }
func (d *duckdbComponent) DisplayName() string      { return "DuckDB" }
func (d *duckdbComponent) Category() stack.Category { return stack.CategoryDatabaseEmbedded }
func (d *duckdbComponent) Description() string      { return "DuckDB CLI from GitHub releases" }
func (d *duckdbComponent) DefaultVersion() string   { return "latest" }
func (d *duckdbComponent) Requires() []string       { return nil }

func (d *duckdbComponent) PathEntries() []stack.PathEntry {
	return []stack.PathEntry{{Shell: "all", Marker: "user-local-bin", Content: `export PATH="$HOME/.local/bin:$PATH"`}}
}

func (d *duckdbComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "duckdb")
	return ok, versionOf(ctx, env, "duckdb", "--version"), nil
}

func (d *duckdbComponent) Install(ctx context.Context, env *stack.Env, _ stack.InstallOptions) error {
	home, _ := os.UserHomeDir()
	bin := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(bin, 0o750)
	// Simplified: download latest linux amd64 zip via gh api pattern
	script := fmt.Sprintf(`set -e
URL=$(curl -fsSL https://api.github.com/repos/duckdb/duckdb/releases/latest | grep -o 'https://[^"]*linux-amd64.zip' | head -1)
curl -fsSL "$URL" -o /tmp/duckdb.zip
unzip -o /tmp/duckdb.zip -d /tmp/duckdb-extract
install -m 0755 /tmp/duckdb-extract/duckdb %s/duckdb`, bin)
	return env.Runner.Run(ctx, "sh", "-c", script)
}

type cudaComponent struct{}

func (c *cudaComponent) ID() string               { return "cuda" }
func (c *cudaComponent) DisplayName() string      { return "CUDA Toolkit" }
func (c *cudaComponent) Category() stack.Category { return stack.CategoryGPU }
func (c *cudaComponent) Description() string      { return "NVIDIA CUDA toolkit (Ubuntu)" }
func (c *cudaComponent) DefaultVersion() string   { return "12-6" }
func (c *cudaComponent) Requires() []string       { return nil }

func (c *cudaComponent) PathEntries() []stack.PathEntry {
	return []stack.PathEntry{{Shell: "all", Marker: "cuda", Content: `if [ -d "/usr/local/cuda/bin" ]; then export PATH="/usr/local/cuda/bin:$PATH"; export LD_LIBRARY_PATH="/usr/local/cuda/lib64:${LD_LIBRARY_PATH:-}"; fi`}}
}

func (c *cudaComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	ok := cmdExists(ctx, env, "nvcc")
	return ok, versionOf(ctx, env, "nvcc", "--version"), nil
}

func (c *cudaComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	ok, err := gpu.DetectNvidia(ctx, env.Runner)
	if err != nil {
		return err
	}
	if !ok && !opts.Force {
		return fmt.Errorf("no NVIDIA GPU detected; use --force to override")
	}
	ver := opts.Version
	if ver == "" {
		ver = c.DefaultVersion()
	}
	if env.Info.IsSilverblue {
		fmt.Fprintln(env.Stderr, "! rpm-ostree CUDA install may require reboot")
		return env.Runner.Run(ctx, "sudo", "rpm-ostree", "install", "--idempotent", "akmod-nvidia", "xorg-x11-drv-nvidia-cuda")
	}
	script := fmt.Sprintf(`set -e
wget -q https://developer.download.nvidia.com/compute/cuda/repos/ubuntu2404/x86_64/cuda-keyring_1.1-1_all.deb
sudo dpkg -i cuda-keyring_1.1-1_all.deb
sudo apt-get update
sudo apt-get install -y cuda-toolkit-%s`, ver)
	if err := env.Runner.Run(ctx, "sh", "-c", script); err != nil {
		return err
	}
	fmt.Fprintln(env.Stdout, "CUDA installation may require a reboot. Verify with nvidia-smi after reboot.")
	return nil
}

type rocmComponent struct{}

func (r *rocmComponent) ID() string               { return "rocm" }
func (r *rocmComponent) DisplayName() string      { return "ROCm" }
func (r *rocmComponent) Category() stack.Category { return stack.CategoryGPU }
func (r *rocmComponent) Description() string      { return "AMD ROCm compute stack" }
func (r *rocmComponent) DefaultVersion() string   { return "latest" }
func (r *rocmComponent) Requires() []string       { return nil }

func (r *rocmComponent) PathEntries() []stack.PathEntry {
	return []stack.PathEntry{{Shell: "all", Marker: "rocm", Content: `if [ -d "/opt/rocm/bin" ]; then export PATH="/opt/rocm/bin:$PATH"; export LD_LIBRARY_PATH="/opt/rocm/lib:${LD_LIBRARY_PATH:-}"; fi`}}
}

func (r *rocmComponent) IsInstalled(ctx context.Context, env *stack.Env) (bool, string, error) {
	if cmdExists(ctx, env, "rocm-smi") {
		return true, versionOf(ctx, env, "rocm-smi", "--version"), nil
	}
	ok := cmdExists(ctx, env, "rocminfo")
	return ok, "", nil
}

func (r *rocmComponent) Install(ctx context.Context, env *stack.Env, opts stack.InstallOptions) error {
	ok, err := gpu.DetectAmd(ctx, env.Runner)
	if err != nil {
		return err
	}
	if !ok && !opts.Force {
		return fmt.Errorf("no AMD GPU detected; use --force to override")
	}
	if env.Info.IsSilverblue {
		return fmt.Errorf("ROCm on Silverblue is experimental; consider Fedora Workstation in a distrobox")
	}
	script := `set -e
wget -q https://repo.radeon.com/amdgpu-install/latest/ubuntu/noble/amdgpu-install_6.3.60300-1_all.deb
sudo apt install -y ./amdgpu-install_*.deb
sudo amdgpu-install --usecase=rocm --no-dkms -y
sudo usermod -a -G render,video "$USER"`
	return env.Runner.Run(ctx, "sh", "-c", script)
}

func init() {
	stack.Register(&dockerComponent{})
	stack.Register(&podmanComponent{})
	stack.Register(&duckdbComponent{})
	stack.Register(&cudaComponent{})
	stack.Register(&rocmComponent{})
}
