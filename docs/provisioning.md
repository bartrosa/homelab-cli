# Provisioning a new machine

End-to-end workflow shipped in **v0.2.0** (Provisioning Release). Requires a Linux host with `lab` installed for ISO download and USB burn.

## 1. Install `lab`

On your current workstation:

```bash
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash
# or pin the release:
curl -sSL https://raw.githubusercontent.com/bartrosa/homelab-cli/main/scripts/install.sh | bash -s -- --version v0.2.0

lab version
lab self-update --check
```

Release artifacts (`.tar.gz`, `.deb`, `.rpm`, `checksums.txt`) are on [GitHub Releases](https://github.com/bartrosa/homelab-cli/releases).

## 2. Create bootable USB (Linux)

```bash
lab iso list
lab iso download ubuntu-desktop
# or: lab iso download fedora-silverblue

lab iso disks
lab iso write                    # interactive: pick cached ISO + USB drive
# or:
lab iso write ubuntu-desktop --usb
```

Safety rules:

- **SYSTEM** disks are rejected unless `--force` (dangerous).
- Confirmation requires typing the full device path (e.g. `/dev/sdb`), not `yes`.
- `lab iso write` may invoke `sudo` for `dd` and `sync` when needed.

Cache directory: `~/.cache/homelab-cli/iso/` (override with `lab iso download --output`).

## 3. Install OS from USB

Boot the target machine from the USB drive and complete the Ubuntu Desktop or Fedora Silverblue installer.

## 4. Bootstrap essentials on the fresh OS

After first boot, install `lab` again on the new machine, then:

```bash
lab bootstrap essentials --dry-run          # preview (auto-detects Ubuntu vs Silverblue)
lab bootstrap essentials --yes              # run all sections

# selective:
lab bootstrap essentials --only cli-basics,build --yes
lab bootstrap essentials --skip docker --target ubuntu
lab bootstrap essentials --target silverblue --yes
```

Sections: `system-update`, `cli-basics`, `shell-tools`, `build`, `container-runtime`, `mise`, `distrobox` (Silverblue), `flatpak-flathub` (Silverblue).

On Fedora Silverblue, `rpm-ostree install` layers may require a **reboot** — `lab` prints a reminder and does not reboot automatically.

## 5. Next steps

```bash
lab toolchain install go rust python
lab pkg ensure ripgrep jq
# point homelab.root in ~/.config/homelab-cli/config.yaml, then:
lab services up ml-stack
lab bootstrap laptop --dry-run
```

See also: [`commands.md`](commands.md), [`configuration.md`](configuration.md), [`external-binaries.md`](external-binaries.md).
