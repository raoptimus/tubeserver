#!/usr/bin/env bash
sockstat -4 -p 80|grep server|awk '{ print $7 }'|sort|uniq -c|sort -n