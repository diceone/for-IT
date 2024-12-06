# For Automation Framework

A distributed automation framework that allows running shell commands across multiple clients from a central server. The framework uses YAML-based configuration files similar to Ansible playbooks to define automation tasks.

## Features

- Central server for distributing automation tasks
- Client-based execution with hostname-based targeting
- Customer-specific environment configurations
- YAML-based configuration with environment support
- Automatic environment reloading when YAML files change
- Role-based task organization with customer-specific customization
- Automatic client inventory tracking with customer and environment information
- Task-level environment variables
- Conditional execution based on hostname patterns
- ETag-based change detection
- Systemd service support
- Secure execution with minimal privileges
- Ansible-style output formatting
- Dry run support

## Architecture

### Server Component
- Watches environment directory for YAML file changes
- Automatically reloads configurations when changes are detected
- Maintains client inventory
- Distributes tasks to matching clients
- Provides RESTful API for client communication
- Supports hostname-based targeting

### Client Component
- Periodically checks server for new tasks
- Executes tasks that match its hostname
- Reports execution results back to server
- Uses ETag-based caching to minimize network traffic
- Supports dry run mode
- Provides Ansible-style output formatting

## Installation

The framework provides multiple installation options for both server and client components:

### Pre-built Packages (Recommended)

#### Debian/Ubuntu:
```bash
# Install server
sudo dpkg -i for-server_<version>_linux_amd64.deb
sudo systemctl start for-server

# Install client
sudo dpkg -i for-client_<version>_linux_amd64.deb
sudo systemctl start for-client
```

#### RHEL/CentOS/Fedora:
```bash
# Install server
sudo rpm -i for-server_<version>_linux_amd64.rpm
sudo systemctl start for-server

# Install client
sudo rpm -i for-client_<version>_linux_amd64.rpm
sudo systemctl start for-client
```

#### macOS:
Download the appropriate archive for your architecture (amd64 or arm64):
```bash
tar xzf for_<version>_darwin_<arch>.tar.gz
sudo ./systemd/install.sh -t server  # For server
sudo ./systemd/install.sh -t client -s server.example.com:8080  # For client
```

### Building from Source

```bash
go build -o for-server ./cmd/server
go build -o for-client ./cmd/client
```

Then use the install script:
```bash
sudo ./systemd/install.sh -t server  # For server
sudo ./systemd/install.sh -t client -s server.example.com:8080  # For client
```

The installation process will:
- Create for user and group
- Set up necessary directories
- Install binaries and service files
- Configure proper permissions
- Enable systemd services (if applicable)

## Configuration

### Environment Structure

The environments directory supports both customer-specific and common roles:

```
environments/
├── roles/              # Common roles shared across customers
│   └── common/
│       ├── tasks.yml  # Basic tools and configurations
│       └── README.md
├── customer1/
│   ├── dev.yml
│   ├── prod.yml
│   └── roles/         # Customer-specific role customizations
│       └── mariadb/
│           ├── tasks.yml
│           └── README.md
└── customer2/
    ├── dev.yml
    └── roles/
        └── mariadb/
            ├── tasks.yml
            └── README.md
```

The structure includes:
- Common roles for shared functionality
- Customer-specific environment files
- Customer-specific role customizations
- Environment-specific configurations

### Common Roles

Common roles provide shared functionality across all customers:

```yaml
# environments/roles/common/tasks.yml
name: Basic Setup
description: Install and configure basic tools
tasks:
  - name: Install common packages
    command: |
      {% if .PackageManager == "apt" %}
      apt-get install -y vim
      {% else if .PackageManager == "dnf" %}
      dnf install -y vim
      {% else %}
      echo "Unsupported package manager"
      exit 1
      {% endif %}
```

To use a common role in an environment file:

```yaml
# customer1/dev.yml
playbooks:
  basic_setup:
    name: Basic System Setup
    description: Configure basic system tools
    hosts: ["*"]  # Apply to all hosts
    include_roles:
      - common
```

### Environment Files

Each customer can have multiple environment files (e.g., dev.yml, prod.yml) that define playbooks and variables specific to that environment.

Example environment file (`customer1/dev.yml`):
```yaml
name: customer1-development
description: Development environment for Customer1

variables:
  APP_ENV: development
  LOG_LEVEL: debug
  CUSTOMER: customer1

mariadb:
  version: "10.11"
  port: 3306
  bind_address: "0.0.0.0"
  max_connections: 151

playbooks:
  database_setup:
    name: Database Server Setup
    description: Install and configure MariaDB
    hosts: ["dev-db-*"]  # Will match dev-db-01.customer1.local, etc.
    include_roles:
      - mariadb
```

### Client Configuration

The client requires both customer and environment parameters:

```bash
for-client \
  --server localhost:8080 \
  --customer customer1 \
  --environment dev \
  --interval 30m
```

Available flags:
- `--server`: Server address (default: "localhost:8080")
- `--customer`: Customer name (required)
- `--environment`: Environment name (required)
- `--interval`: Check interval for continuous mode (default: 30m)
- `--run-once`: Execute once and exit
- `--dry-run`: Show what would be executed without making changes

Example with all options:
```bash
for-client \
  --server db.example.com:8080 \
  --customer customer1 \
  --environment prod \
  --interval 1h \
  --dry-run
```

### Roles

Roles are reusable collections of tasks. Each customer can have their own customized version of a role:

```
environments/
├── customer1/
│   ├── dev.yml
│   └── roles/
│       └── mariadb/
│           ├── tasks.yml
│           └── README.md
└── customer2/
    ├── dev.yml
    └── roles/
        └── mariadb/
            ├── tasks.yml
            └── README.md
```

Example role tasks file (`customer1/roles/mariadb/tasks.yml`):
```yaml
name: MariaDB Installation
description: Install and configure MariaDB server
tasks:
  - name: Install MariaDB packages
    command: dnf install -y mariadb-server
    when: "hostname matches 'db-*'"
    variables:
      CUSTOMER: "{{ .CUSTOMER }}"
      ENVIRONMENT: "{{ .ENVIRONMENT }}"
```

## Client Usage

The client supports several operation modes:

### Continuous Mode (Default)
```bash
for-client --server localhost:8080 --interval 30m
```

### Manual Trigger (Run Once)
```bash
for-client --server localhost:8080 --run-once
```

### Dry Run Mode
```bash
for-client --server localhost:8080 --dry-run
```

Available flags:
- `--server`: Server address (default: "localhost:8080")
- `--interval`: Check interval for continuous mode (default: 30m)
- `--run-once`: Execute once and exit
- `--dry-run`: Show what would be executed without making changes

Example output:
```
PLAY [dev-db-01] ******************************************************************

TASK [Install MariaDB packages] *************************************************
        changed: [localhost] => {"changed": true, "duration": 45.32s}
        Output:
          Package mariadb-server-10.11.5 installed

PLAY RECAP *********************************************************************
localhost                  : ok=3    changed=1    failed=0    skipped=0

Playbook finished in 45.45 seconds
```

## Server Features

### Client Inventory
The server automatically maintains an inventory of all connected clients in `/etc/for/inventory.json`. The inventory includes:
- Hostname
- IP address
- First seen timestamp
- Last seen timestamp
- Environment (if specified)
- Customer (if specified)

View the current inventory:
```bash
curl http://localhost:8080/inventory
```

Example inventory output:
```json
[
  {
    "hostname": "dev-db-01",
    "ip": "192.168.1.101",
    "customer": "customer1",
    "environment": "development",
    "last_seen": "2023-11-15T14:30:45Z",
    "first_seen": "2023-11-15T10:00:00Z"
  }
]
```

The inventory is automatically updated when:
- A client requests its playbooks
- A client submits task results
- A client performs a manual check

## Security

The framework implements several security measures:
- Runs with minimal privileges using dedicated system user
- Systemd service hardening (NoNewPrivileges, ProtectSystem, etc.)
- Environment-specific variables
- Hostname-based task targeting
- Conditional execution support

## Logging

- Server logs: `/var/log/for/for-server.log`
- Client logs: `/var/log/for/for-client.log`
- Systemd logs: `journalctl -u for-server` or `journalctl -u for-client`

## Development

1. Clone the repository
2. Install dependencies:
```bash
go mod download
```

3. Run tests:
```bash
go test ./...
```

## Directory Structure

```
for/
├── cmd/
│   ├── client/        # Client entry point
│   └── server/        # Server entry point
├── environments/      # Environment configurations
│   ├── customer1/
│   │   ├── dev.yml
│   │   └── roles/
│   │       └── mariadb/
│   │           ├── tasks.yml
│   │           └── README.md
│   └── customer2/
│       ├── dev.yml
│       └── roles/
│           └── mariadb/
│               ├── tasks.yml
│               └── README.md
├── internal/
│   ├── api/          # API implementations
│   ├── executor/     # Command execution
│   └── models/       # Data models
└── systemd/          # Systemd service files
```

## License

MIT License

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
