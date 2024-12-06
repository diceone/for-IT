# Common Role

This role installs and configures basic tools and packages that are commonly needed across all systems.

## Included Tools and Packages

- vim: Text editor

## Usage

Include this role in your environment configuration:

```yaml
playbooks:
  basic_setup:
    name: Basic System Setup
    description: Install common tools and packages
    hosts: ["*"]  # Or specific host patterns
    include_roles:
      - common
```

## Package Manager Support

The role automatically detects and uses the appropriate package manager:
- apt-get (Debian/Ubuntu)
- dnf (Fedora/RHEL 8+)
- yum (CentOS/RHEL 7)
- pacman (Arch Linux)
