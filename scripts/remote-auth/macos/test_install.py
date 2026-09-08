"""Test installer control flow without changing macOS LaunchAgents."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class InstallTest(unittest.TestCase):
    def run_install(self, loaded=False, list_status=0, bootout_status=0, bootstrap_status=0):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            calls = root / "calls"
            commands = {
                "mkdir": "exit 0",
                "cp": "exit 0",
                "plutil": "exit 0",
                "yandex-mcp": "exit 0",
                "yc": "exit 0",
                "ssh": "exit 0",
                "launchctl": '''echo "$1" >> "$TEST_CALLS"
case "$1" in
  list)
    if [ "$TEST_LIST_STATUS" != 0 ]; then
      echo 'list failed' >&2
      exit "$TEST_LIST_STATUS"
    fi
    printf 'PID\\tStatus\\tLabel\\n'
    printf '%s\\n' '- 0 net.yandex-mcp.auth-agent.other'
    if [ "$TEST_LOADED" = 1 ]; then
      printf '%s\\n' '- 0 net.yandex-mcp.auth-agent'
    fi
    ;;
  bootout)
    if [ "$TEST_BOOTOUT_STATUS" != 0 ]; then
      echo 'bootout failed' >&2
      exit "$TEST_BOOTOUT_STATUS"
    fi
    ;;
  bootstrap)
    if [ "$TEST_BOOTSTRAP_STATUS" != 0 ]; then
      echo 'bootstrap failed' >&2
      exit "$TEST_BOOTSTRAP_STATUS"
    fi
    ;;
  *) echo 'unexpected launchctl command' >&2; exit 99 ;;
esac
''',
            }
            for name, body in commands.items():
                executable = root / name
                executable.write_text("#!/bin/sh\n" + body + "\n")
                executable.chmod(0o755)
            env = os.environ.copy()
            env.update({
                "PATH": str(root) + ":/usr/bin:/bin",
                "YANDEX_MCP_BINARY": str(root / "yandex-mcp"),
                "YANDEX_MCP_AUTH_AGENT_YC_PATH": str(root / "yc"),
                "YANDEX_MCP_AUTH_AGENT_PORT": "18765",
                "YANDEX_MCP_SSH_TARGET": "test@example.invalid",
                "TEST_CALLS": str(calls),
                "TEST_LOADED": str(int(loaded)),
                "TEST_LIST_STATUS": str(list_status),
                "TEST_BOOTOUT_STATUS": str(bootout_status),
                "TEST_BOOTSTRAP_STATUS": str(bootstrap_status),
            })
            result = subprocess.run(
                ["/bin/sh", str(Path(__file__).with_name("install.sh"))],
                env=env, capture_output=True, text=True, timeout=10,
            )
            return result, calls.read_text().splitlines()

    def test_first_install(self):
        result, calls = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["list", "bootstrap"])

    def test_reinstall(self):
        result, calls = self.run_install(loaded=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["list", "bootout", "bootstrap"])

    def test_list_failure(self):
        result, calls = self.run_install(list_status=5)
        self.assertEqual(result.returncode, 5)
        self.assertIn("list failed", result.stderr)
        self.assertEqual(calls, ["list"])

    def test_bootout_failure(self):
        result, calls = self.run_install(loaded=True, bootout_status=5)
        self.assertEqual(result.returncode, 5)
        self.assertIn("bootout failed", result.stderr)
        self.assertEqual(calls, ["list", "bootout"])

    def test_bootstrap_failure(self):
        result, calls = self.run_install(bootstrap_status=5)
        self.assertEqual(result.returncode, 5)
        self.assertIn("bootstrap failed", result.stderr)
        self.assertEqual(calls, ["list", "bootstrap"])


if __name__ == "__main__":
    unittest.main()
