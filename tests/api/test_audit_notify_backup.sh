#!/usr/bin/env bash
set -euo pipefail
bash tests/api/test_admin_ops.sh
bash tests/api/test_notification_read.sh
