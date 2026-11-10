"""Regression tests for EC report input and optional export."""

import contextlib
import io
import json
import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import print_ec_failures


class ReportTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.previous_directory = Path.cwd()
        os.chdir(self.directory.name)
        self.addCleanup(os.chdir, self.previous_directory)
        self.report = {
            "components": [{"name": "café", "violations": [{"msg": "échec"}]}]
        }

    def run_report(self, report, export=False):
        Path("input.log").write_text(
            "ignored\n[report-json]" + json.dumps(report, ensure_ascii=False) + "\n",
            encoding="utf-8",
        )
        argv = ["print_ec_failures.py", "input.log"] + (["--json"] if export else [])
        stdout, stderr = io.StringIO(), io.StringIO()
        with patch("sys.argv", argv), contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
            result = print_ec_failures.main()
        return result, stdout.getvalue(), stderr.getvalue()

    def test_default_prints_utf8_without_export(self):
        result, output, error = self.run_report(self.report)
        self.assertEqual(result, 0)
        self.assertIn("Component: café", output)
        self.assertIn("#1: échec", output)
        self.assertIn("total checks: 1", output)
        self.assertEqual(error, "")
        self.assertFalse(Path("logs.json").exists())

    def test_opt_in_export(self):
        result, output, error = self.run_report(self.report, export=True)
        self.assertEqual(result, 0)
        self.assertIn("Component: café", output)
        self.assertEqual(json.loads(Path("logs.json").read_text(encoding="utf-8")), self.report)
        self.assertEqual(error, "")

    def test_existing_export_is_preserved_and_results_print(self):
        Path("logs.json").write_bytes(b"keep me")
        result, output, error = self.run_report(self.report, export=True)
        self.assertEqual(result, 1)
        self.assertIn("Component: café", output)
        self.assertIn("Unable to write JSON report:", error)
        self.assertEqual(Path("logs.json").read_bytes(), b"keep me")

    def test_invalid_shapes_never_create_or_modify_export(self):
        invalid = [None, [], "components", 42, {}, {"components": None},
                   {"components": {}}, {"components": "bad"}]
        for report in invalid:
            for existing in (False, True):
                with self.subTest(report=report, existing=existing):
                    destination = Path("logs.json")
                    if existing:
                        destination.write_bytes(b"keep me")
                    result, output, error = self.run_report(report, export=True)
                    self.assertEqual(result, 1)
                    self.assertEqual(output.strip(), "Unable to find components attribute in EC data")
                    self.assertEqual(error, "")
                    if existing:
                        self.assertEqual(destination.read_bytes(), b"keep me")
                        destination.unlink()
                    else:
                        self.assertFalse(destination.exists())

    def test_unwritable_export_does_not_prevent_terminal_report(self):
        original_open = Path.open

        def open_with_denied_export(path, *args, **kwargs):
            if path == Path("logs.json"):
                raise PermissionError("read-only working directory")
            return original_open(path, *args, **kwargs)

        with patch.object(Path, "open", open_with_denied_export):
            result, output, error = self.run_report(self.report, export=True)
        self.assertEqual(result, 1)
        self.assertIn("Component: café", output)
        self.assertIn("Unable to write JSON report:", error)

    def test_write_failure_does_not_prevent_terminal_report(self):
        with patch.object(print_ec_failures.json, "dump", side_effect=OSError("disk full")):
            result, output, error = self.run_report(self.report, export=True)
        self.assertEqual(result, 1)
        self.assertIn("Component: café", output)
        self.assertIn("Unable to write JSON report:", error)

    def test_empty_components_valid(self):
        result, output, error = self.run_report({"components": []})
        self.assertEqual((result, output, error), (0, "", ""))


if __name__ == "__main__":
    unittest.main()
