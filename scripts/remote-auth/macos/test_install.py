"""Test installer control flow without changing macOS LaunchAgents."""

import os
from pathlib import Path
import subprocess
import tempfile
import unittest


# Sentinel that keeps the variable out of the installer environment.
UNSET = object()
# Sentinel that points the variable at the stub yc executable.
STUB = object()


class InstallTest(unittest.TestCase):
    def run_install(self, loaded=False, list_status=0, bootout_status=0, bootstrap_status=0,
                    yc_in_path=True, yc_path=STUB, oauth_dir=UNSET, bootstrap_failures=0):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            calls = root / "calls"
            plutil_calls = root / "plutil-calls"
            bootstrap_failures_file = root / "bootstrap-failures"
            calls.touch()
            plutil_calls.touch()
            commands = {
                "mkdir": "exit 0",
                "cp": "exit 0",
                "plutil": 'printf \'%s\\t\' "$@" >> "$TEST_PLUTIL_CALLS"; echo >> "$TEST_PLUTIL_CALLS"',
                "yandex-mcp": "exit 0",
                "ssh": "exit 0",
                "sleep": "exit 0",
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
    failures=$(cat "$TEST_BOOTSTRAP_FAILURES_FILE" 2>/dev/null || printf '0')
    if [ "$failures" -lt "$TEST_BOOTSTRAP_FAILURES" ]; then
      printf '%s\n' "$((failures + 1))" > "$TEST_BOOTSTRAP_FAILURES_FILE"
      echo 'bootstrap failed' >&2
      exit 5
    fi
    if [ "$TEST_BOOTSTRAP_STATUS" != 0 ]; then
      echo 'bootstrap failed' >&2
      exit "$TEST_BOOTSTRAP_STATUS"
    fi
    ;;
  *) echo 'unexpected launchctl command' >&2; exit 99 ;;
esac
''',
            }
            if yc_in_path:
                commands["yc"] = "exit 0"
            for name, body in commands.items():
                executable = root / name
                executable.write_text("#!/bin/sh\n" + body + "\n")
                executable.chmod(0o755)
            env = os.environ.copy()
            env.pop("YANDEX_MCP_AUTH_AGENT_YC_PATH", None)
            env.pop("YANDEX_MCP_AUTH_AGENT_OAUTH_DIR", None)
            if yc_path is STUB:
                env["YANDEX_MCP_AUTH_AGENT_YC_PATH"] = str(root / "yc")
            elif yc_path is not UNSET:
                env["YANDEX_MCP_AUTH_AGENT_YC_PATH"] = yc_path
            if oauth_dir is not UNSET:
                env["YANDEX_MCP_AUTH_AGENT_OAUTH_DIR"] = oauth_dir
            env.update({
                "HOME": str(root),
                "PATH": str(root) + ":/usr/bin:/bin",
                "YANDEX_MCP_BINARY": str(root / "yandex-mcp"),
                "YANDEX_MCP_AUTH_AGENT_PORT": "18765",
                "YANDEX_MCP_SSH_TARGETS": "first@example.invalid,second@example.invalid",
                "TEST_CALLS": str(calls),
                "TEST_PLUTIL_CALLS": str(plutil_calls),
                "TEST_LOADED": str(int(loaded)),
                "TEST_LIST_STATUS": str(list_status),
                "TEST_BOOTOUT_STATUS": str(bootout_status),
                "TEST_BOOTSTRAP_STATUS": str(bootstrap_status),
                "TEST_BOOTSTRAP_FAILURES": str(bootstrap_failures),
                "TEST_BOOTSTRAP_FAILURES_FILE": str(bootstrap_failures_file),
            })
            result = subprocess.run(
                ["/bin/sh", str(Path(__file__).with_name("install.sh"))],
                env=env, capture_output=True, text=True, timeout=10,
            )
            settings = {}
            for line in plutil_calls.read_text().splitlines():
                arguments = line.split("\t")
                if arguments[:1] == ["-replace"] and arguments[1].startswith("EnvironmentVariables."):
                    settings[arguments[1].removeprefix("EnvironmentVariables.")] = arguments[3]
            return result, calls.read_text().splitlines(), settings

    def test_first_install(self):
        result, calls, _ = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["list", "bootstrap"])
        plist = Path(__file__).with_name("auth-agent.plist").read_text()
        self.assertIn("YANDEX_MCP_SSH_TARGETS", plist)
        self.assertNotIn("YANDEX_MCP_SSH_TARGET</key>", plist)

    def test_reinstall(self):
        result, calls, _ = self.run_install(loaded=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["list", "bootout", "bootstrap"])

    def test_list_failure(self):
        result, calls, _ = self.run_install(list_status=5)
        self.assertEqual(result.returncode, 5)
        self.assertIn("list failed", result.stderr)
        self.assertEqual(calls, ["list"])

    def test_bootout_failure(self):
        result, calls, _ = self.run_install(loaded=True, bootout_status=5)
        self.assertEqual(result.returncode, 5)
        self.assertIn("bootout failed", result.stderr)
        self.assertEqual(calls, ["list", "bootout"])

    def test_transient_bootstrap_failure(self):
        result, calls, _ = self.run_install(bootstrap_failures=1)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["list", "bootstrap", "bootstrap"])

    def test_persistent_bootstrap_failure(self):
        result, calls, _ = self.run_install(bootstrap_status=5)
        self.assertEqual(result.returncode, 5)
        self.assertIn("bootstrap failed", result.stderr)
        self.assertEqual(calls, ["list", "bootstrap", "bootstrap", "bootstrap"])

    def test_install_without_yc_stores_command_name(self):
        result, calls, settings = self.run_install(yc_in_path=False, yc_path=UNSET)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(calls, ["list", "bootstrap"])
        self.assertEqual(settings["YANDEX_MCP_AUTH_AGENT_YC_PATH"], "yc")

    def test_install_with_yc_in_path_stores_absolute_path(self):
        result, _, settings = self.run_install(yc_path=UNSET)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(settings["YANDEX_MCP_AUTH_AGENT_YC_PATH"].endswith("/yc"))
        self.assertTrue(Path(settings["YANDEX_MCP_AUTH_AGENT_YC_PATH"]).is_absolute())

    def test_missing_explicit_yc_fails(self):
        result, calls, _ = self.run_install(yc_path="/nonexistent/yc")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("YANDEX_MCP_AUTH_AGENT_YC_PATH does not name an executable: /nonexistent/yc", result.stderr)
        self.assertEqual(calls, [])

    def test_absolute_oauth_dir_is_stored(self):
        result, _, settings = self.run_install(oauth_dir="/Users/test/custom oauth")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(settings["YANDEX_MCP_AUTH_AGENT_OAUTH_DIR"], "/Users/test/custom oauth")

    def test_relative_oauth_dir_fails(self):
        result, calls, settings = self.run_install(oauth_dir="custom-oauth")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("YANDEX_MCP_AUTH_AGENT_OAUTH_DIR", result.stderr)
        self.assertEqual(calls, [])
        self.assertEqual(settings, {})

    def test_unset_oauth_dir_stores_empty_value(self):
        result, _, settings = self.run_install()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(settings["YANDEX_MCP_AUTH_AGENT_OAUTH_DIR"], "")


if __name__ == "__main__":
    unittest.main()
