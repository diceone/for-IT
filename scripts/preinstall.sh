#!/bin/sh

# Create for group if it doesn't exist
if ! getent group for >/dev/null; then
    groupadd -r for
fi

# Create for user if it doesn't exist
if ! getent passwd for >/dev/null; then
    useradd -r -g for -s /sbin/nologin -d /etc/for for
fi

# Create necessary directories with proper permissions
mkdir -p /etc/for
mkdir -p /var/log/for
chown -R for:for /etc/for
chown -R for:for /var/log/for
chmod 755 /etc/for
chmod 755 /var/log/for
