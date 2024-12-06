#!/bin/sh

# Set proper permissions for binaries
chmod 755 /usr/local/bin/for-*

# Set proper permissions for config files
if [ -d /etc/for/environments ]; then
    chmod 755 /etc/for/environments
    chmod -R 644 /etc/for/environments/*
    chmod 755 $(find /etc/for/environments -type d)
    chown -R for:for /etc/for/environments
fi

# Reload systemd
systemctl daemon-reload

# Enable service (but don't start it)
if [ -f /etc/systemd/system/for-server.service ]; then
    systemctl enable for-server.service
fi
if [ -f /etc/systemd/system/for-client.service ]; then
    systemctl enable for-client.service
fi

echo "Installation complete! To start the service:"
echo "  systemctl start for-server  # For server"
echo "  systemctl start for-client  # For client"
