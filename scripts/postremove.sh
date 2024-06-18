#!/bin/sh

systemctl stop --now reddlinks.service
systemctl daemon-reload
userdel reddlinks
