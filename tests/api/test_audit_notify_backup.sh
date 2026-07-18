#!/usr/bin/env bash
set -euo pipefail
bash API_tests/test_admin_ops.sh
bash API_tests/test_notification_read.sh
