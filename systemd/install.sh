#!/bin/bash

# Print usage information
print_usage() {
    echo "Usage: $0 [-t type] [-s server_address]"
    echo "Options:"
    echo "  -t type           Installation type (server or client)"
    echo "  -s server_address Server address for client installation (e.g., localhost:8080)"
    echo "  -h               Show this help message"
    echo
    echo "Examples:"
    echo "  Install server: $0 -t server"
    echo "  Install client: $0 -t client -s localhost:8080"
}

# Default values
INSTALL_TYPE=""
SERVER_ADDRESS="localhost:8080"

# Parse command line arguments
while getopts "t:s:h" opt; do
    case $opt in
        t)
            INSTALL_TYPE=$OPTARG
            ;;
        s)
            SERVER_ADDRESS=$OPTARG
            ;;
        h)
            print_usage
            exit 0
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            print_usage
            exit 1
            ;;
    esac
done

# Validate installation type
if [ -z "$INSTALL_TYPE" ]; then
    echo "Error: Installation type (-t) is required"
    print_usage
    exit 1
fi

if [ "$INSTALL_TYPE" != "server" ] && [ "$INSTALL_TYPE" != "client" ]; then
    echo "Error: Invalid installation type. Must be 'server' or 'client'"
    print_usage
    exit 1
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Create for user and group
if ! getent group for >/dev/null; then
    groupadd -r for
fi

if ! getent passwd for >/dev/null; then
    useradd -r -g for -s /sbin/nologin -d /etc/for for
fi

# Install server
install_server() {
    echo "Installing For Automation Framework Server..."
    
    # Create necessary directories
    mkdir -p /etc/for/environments/roles/common
    mkdir -p /var/log/for

    # Create for user and group if they don't exist
    if ! getent group for >/dev/null; then
        groupadd -r for
    fi
    if ! getent passwd for >/dev/null; then
        useradd -r -g for -s /sbin/nologin -d /etc/for for
    fi

    # Copy server binary
    if [ ! -f ../for-server ]; then
        echo "Error: for-server binary not found"
        exit 1
    fi
    cp ../for-server /usr/local/bin/
    chmod 755 /usr/local/bin/for-server

    # Copy environment files
    if [ -d ../environments ]; then
        cp -r ../environments/* /etc/for/environments/
    else
        echo "Warning: environments directory not found"
        mkdir -p /etc/for/environments/roles/common
    fi

    # Copy systemd service
    cp for-server.service /etc/systemd/system/

    # Set permissions
    chown -R for:for /etc/for
    chown -R for:for /var/log/for
    chmod 755 /etc/for
    chmod 755 /etc/for/environments
    chmod -R 644 /etc/for/environments/*
    chmod 755 $(find /etc/for/environments -type d)

    # Reload systemd
    systemctl daemon-reload

    echo "Server installation complete!"
    echo "To start the server:"
    echo "systemctl enable --now for-server"
}

# Install client
install_client() {
    echo "Installing For Automation Framework Client..."
    
    # Create necessary directories
    mkdir -p /var/log/for

    # Create for user and group if they don't exist
    if ! getent group for >/dev/null; then
        groupadd -r for
    fi
    if ! getent passwd for >/dev/null; then
        useradd -r -g for -s /sbin/nologin -d /etc/for for
    fi

    # Copy client binary
    if [ ! -f ../for-client ]; then
        echo "Error: for-client binary not found"
        exit 1
    fi
    cp ../for-client /usr/local/bin/
    chmod 755 /usr/local/bin/for-client

    # Modify client service file with correct server address
    sed "s/localhost:8080/$SERVER_ADDRESS/g" for-client.service > /etc/systemd/system/for-client.service

    # Set permissions
    chown -R for:for /var/log/for
    chmod 755 /var/log/for

    # Reload systemd
    systemctl daemon-reload

    echo "Client installation complete!"
    echo "To start the client:"
    echo "systemctl enable --now for-client"
}

# Perform installation based on type
case "$INSTALL_TYPE" in
    server)
        install_server
        ;;
    client)
        if [ "$INSTALL_TYPE" = "client" ] && [ "$SERVER_ADDRESS" = "localhost:8080" ]; then
            echo "Warning: Using default server address (localhost:8080). Use -s to specify a different address."
        fi
        install_client
        ;;
esac

# Reload systemd
systemctl daemon-reload

echo
echo "Installation of $INSTALL_TYPE completed successfully!"
echo "For logs: tail -f /var/log/for/for-$INSTALL_TYPE.log"
