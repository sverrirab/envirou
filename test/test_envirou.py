from __future__ import print_function, unicode_literals
import os
import subprocess
import unittest

# Uses system python - replace this with e.g. 'python3', 'python27', etc
PYTHON = "python"


class ExecException(Exception):
    pass


def get_path(*segments):
    return os.path.normpath(os.path.join(os.path.dirname(__file__), "..", *segments))


def envirou(args=""):
    env = {
        "ENVIROU_HOME": get_path("test", "config_one"),
        "SHELL": "zsh",
        "VARIABLE_ONE": "first",
        "VARIABLE_TWO": "",
        "VARIABLE_THREE": "one two three",
    }
    cwd = get_path()
    cmd = " ".join(["python3", "envirou.py", args])
    p = subprocess.Popen(
        cmd,
        shell=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        bufsize=0,
        cwd=cwd,
        env=env,
    )
    stdout, stderr = p.communicate()
    if p.returncode != 0:
        raise ExecException(
            "Failed to execute {} [{}]: \nstdout: {}\nstderr: {}".format(
                cmd, p.returncode, stdout.decode(), stderr.decode()
            )
        )

    return stdout.decode(), stderr.decode()


class TestFailure(unittest.TestCase):
    def test_simple(self):
        stdout, stderr = envirou("--help")
        self.assertEqual(0, len(stdout))

    def test_variables(self):
        stdout, stderr = envirou()
        self.assertEqual(0, len(stdout))
        self.assertFind("VARIABLE_ONE", stderr)
        self.assertFind("VARIABLE_TWO", stderr)
        self.assertFind("VARIABLE_THREE", stderr)
        self.assertFind("one two three", stderr)

        self.assertNotFind("VARIABLE_FOUR", stderr)
        self.assertNotFind("four", stderr)

    def test_list(self):
        stdout, stderr = envirou("--list")
        self.assertEqual(0, len(stdout))
        lines = stderr.split("\n")
        self.assertEqual(4, len(lines))
        self.assertFind("# .ignore", lines[0])
        self.assertFind("# example", lines[1])
        self.assertFind("# test", lines[2])
        self.assertEqual("", lines[3])

    def test_profile(self):
        stdout, stderr = envirou("example")
        self.assertFind("profile", stderr)
        self.assertFind("example", stderr)
        self.assertFind("activated", stderr)
        shell_cmd = sorted(stdout.split("\n"))
        self.assertEqual(4, len(shell_cmd))
        self.assertEqual("", shell_cmd[0])
        self.assertEqual("export EXAMPLE_EMPTY_VARIABLE=;", shell_cmd[1])
        self.assertEqual('export EXAMPLE_OCCUPATION="elevator operator";', shell_cmd[2])
        self.assertEqual("unset EXAMPLE_UNSET_VARIABLE;", shell_cmd[3])

    def test_invalid_usage(self):
        with self.assertRaises(ExecException):
            envirou("--invalid")

    # Utility functions

    def assertFind(self, partial, full):
        self.assertNotEqual(
            -1, full.find(partial), "'{}' not found in '{}'".format(partial, full)
        )

    def assertNotFind(self, partial, full):
        self.assertEqual(
            -1, full.find(partial), "'{}' found in '{}'".format(partial, full)
        )
