# For-IT Automation Framework

For-IT is a lightweight automation framework designed to execute tasks across multiple hosts. It follows a client-server architecture where the server distributes tasks to clients based on playbook configurations.

## Features

- Client-server architecture for distributed task execution
- YAML-based playbook configuration
- Real-time task execution and monitoring
- Support for conditional task execution
- Systemd integration for both server and client
- Cross-platform support (Linux and macOS)
- DEB and RPM package support

## Installation

### Using DEB/RPM Packages

1. Download the latest release from [GitHub Releases](https://github.com/diceone/for-IT/releases)

2. Install the server package:
   ```bash
   # For Debian/Ubuntu
   sudo dpkg -i for-server_*.deb
   
   # For RHEL/CentOS
   sudo rpm -i for-server_*.rpm
   ```

3. Install the client package:
   ```bash
   # For Debian/Ubuntu
   sudo dpkg -i for-client_*.deb
   
   # For RHEL/CentOS
   sudo rpm -i for-client_*.rpm
   ```

### Manual Installation

1. Download the latest release archive
2. Extract the archive:
   ```bash
   tar xzf for-IT_*.tar.gz
   ```
3. Copy binaries to /usr/local/bin:
   ```bash
   sudo cp for-server /usr/local/bin/
   sudo cp for-client /usr/local/bin/
   ```
4. Copy systemd service files:
   ```bash
   sudo cp systemd/for-server.service /etc/systemd/system/
   sudo cp systemd/for-client.service /etc/systemd/system/
   ```

## Configuration

### Directory Structure

The framework uses an Ansible-like directory structure:
```
/etc/for/
└── environments/
    ├── customer1/
    │   ├── dev.yml       # Development environment config
    │   ├── prod.yml      # Production environment config
    │   └── roles/        # Customer-specific roles
    │       └── mariadb/
    │           └── tasks.yml
    └── roles/            # Global roles
        └── common/       # Common tasks for all customers
            └── tasks.yml
```

### Environment Configuration

Each customer can have multiple environment configurations (e.g., `dev.yml`, `prod.yml`):
```yaml
name: customer1-production
description: Production environment for Customer1

variables:
  APP_ENV: production
  LOG_LEVEL: info
  CUSTOMER: customer1

# Role-specific configurations
mariadb:
  version: "10.11"
  port: 3306
  max_connections: 500

playbooks:
  basic_setup:
    name: Basic System Setup
    description: Install common tools and packages
    include: /etc/for/environments/roles/common/tasks.yml
```

### Roles

1. **Common Role** (`/etc/for/environments/roles/common/tasks.yml`):
   - Basic system configurations
   - Common package installation
   - System-wide settings

2. **Customer-Specific Roles** (`/etc/for/environments/customer1/roles/*/tasks.yml`):
   - Service-specific configurations
   - Customer-specific packages
   - Custom scripts and tools

### Server Setup

1. Create the environments directory:
   ```bash
   sudo mkdir -p /etc/for/environments
   ```

2. Copy your environment configurations and roles:
   ```bash
   sudo cp -r environments/* /etc/for/environments/
   ```

3. Start the server:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable for-server
   sudo systemctl start for-server
   ```

### Client Setup

1. Start the client with your customer name and environment:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable for-client@customer1
   sudo systemctl start for-client@customer1
   ```

2. Configure the client environment:
   ```bash
   # Edit the client service configuration
   sudo systemctl edit for-client@customer1
   ```
   Add:
   ```ini
   [Service]
   Environment=FOR_ENVIRONMENT=production  # or development
   ```

## Usage

### Server Command-Line Options

```bash
for-server [options]
  --addr string          Server address (default ":8080")
  --playbook-dir string  Directory containing playbook files (default "playbooks")
```

### Client Command-Line Options

```bash
for-client [options]
  --server string       Server address (default "localhost:8080")
  --interval duration   Check interval (default 5m)
  --customer string     Customer name (required)
  --environment string  Environment name (default "production")
  --dry-run            Dry run mode
```

## Playbook Format

```yaml
name: Playbook Name
customer: customer_name
environment: environment_name
tasks:
  - name: Task Name
    command: command_to_execute
    when: condition  # Optional condition
    variables:       # Optional environment variables
      KEY: value
```

## Development

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/diceone/for-IT.git
   cd for-IT
   ```

2. Build the binaries:
   ```bash
   go build ./cmd/server
   go build ./cmd/client
   ```

### Running Tests

```bash
go test ./...
```

## License

MIT License - see LICENSE file for details.
