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

### Server Setup

1. Create the playbook directory:
   ```bash
   sudo mkdir -p /etc/for/playbooks
   ```

2. Create a playbook file (e.g., `/etc/for/playbooks/example.yml`):
   ```yaml
   name: Example Playbook
   customer: mycustomer
   environment: production
   tasks:
     - name: Check disk space
       command: df -h
       when: hostname == "webserver1"
     
     - name: Check memory
       command: free -m
   ```

3. Start the server:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable for-server
   sudo systemctl start for-server
   ```

### Client Setup

1. Start the client with your customer name:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable for-client@mycustomer
   sudo systemctl start for-client@mycustomer
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
