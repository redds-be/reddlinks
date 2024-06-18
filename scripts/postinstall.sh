#!/bin/sh

chmod 0644 /etc/systemd/system/reddlinks.service
systemctl daemon-reload
chmod 0660 /opt/reddlinks/.env
chown -R reddlinks:reddlinks /opt/reddlinks
