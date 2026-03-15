#!/bin/bash
set -e

cat >> /var/lib/postgresql/data/postgresql.conf << 'EOF'

# Custom settings
listen_addresses = '*'
port = 5432

log_destination = 'stderr'
logging_collector = on
log_directory = '/var/log/postgresql'
log_filename = 'postgres.log'
log_statement = 'all'
log_min_duration_statement = 0
log_line_prefix = '%m [%p] %q%u@%d '
log_timezone = 'UTC'

timezone = 'UTC'

lc_messages = 'en_US.UTF-8'
lc_monetary = 'en_US.UTF-8'
lc_numeric = 'en_US.UTF-8'
lc_time = 'en_US.UTF-8'
EOF