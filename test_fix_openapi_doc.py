"""Behavioral tests for the OpenAPI documentation command-line filter."""

from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("fix_openapi_doc.py")


class TestFixOpenapiDoc(unittest.TestCase):
    """Exercise the script in isolation because importing it consumes stdin."""

    def assert_filter(self, source, expected, *filenames):
        result = subprocess.run(
            [sys.executable, str(SCRIPT), *map(str, filenames)],
            input=source,
            text=True,
            capture_output=True,
            timeout=5,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stderr, "")
        self.assertEqual(result.stdout, expected)

    def test_plain_text_and_line_boundaries(self):
        cases = [
            ("", ""),
            ("\n", "\n"),
            ("last line", "last line\n"),
            ("first\n\nlast\n", "first\n\nlast\n"),
            ("first\r\nlast\r\n", "first\nlast\n"),
            ("  indented\t \nUnicode: λ 日本語\n", "  indented\t \nUnicode: λ 日本語\n"),
        ]
        for source, expected in cases:
            with self.subTest(source=source):
                self.assert_filter(source, expected)

    def test_only_exact_examples_heading_is_removed(self):
        self.assert_filter(
            "Examples\nExamples of use\n Examples\nExamples \nexamples\nExamples",
            "Examples of use\n Examples\nExamples \nexamples\n",
        )

    def test_json_blocks_are_removed(self):
        cases = [
            ("```json\n```\n", ""),
            ("before\n```json\n{\n  \"key\": 1\n}\n```\nafter\n", "before\nafter\n"),
            ("```json\n{}\n``` \t\nafter", "after\n"),
            ("```json\n{}\n```\nmiddle\n```json\n[]\n```", "middle\n"),
            ("before\n```json\nunterminated", "before\n"),
            ("```json", ""),
            # Language markers inside JSON must not enter protected-code mode.
            ("```json\n```python\nExamples\n |\n```\nafter\n", "after\n"),
        ]
        for source, expected in cases:
            with self.subTest(source=source):
                self.assert_filter(source, expected)

    def test_python_and_text_blocks_preserve_filter_markers(self):
        for language in ("python", "text"):
            for closing in ("```\n", "``` \t\n", ""):
                with self.subTest(language=language, closing=closing):
                    block = (
                        f"```{language}\nExamples\n```json\n"
                        "{\"example\": true}\n| row\n |\n" + closing
                    )
                    self.assert_filter(block, block)

    def test_filtering_resumes_after_protected_code(self):
        for language in ("python", "text"):
            with self.subTest(language=language):
                self.assert_filter(
                    f"```{language}\nExamples\n``` \t\n"
                    "Examples\n```json\n{}\n```\n| row\n |\nend\n",
                    f"```{language}\nExamples\n``` \t\n| row |\nend\n",
                )

    def test_unmatched_closing_fence_is_preserved(self):
        self.assert_filter("before\n``` \t\nafter\n", "before\n``` \t\nafter\n")

    def test_table_continuations(self):
        cases = [
            ("| row\n |\nafter\n", "| row |\nafter\n"),
            # A consumed continuation at EOF must not be printed twice.
            ("| row\n |", "| row |\n"),
            ("| row\r\n |\r\n", "| row |\n"),
            ("| first\n |\n| second\n |\n", "| first |\n| second |\n"),
            ("| row\n |\n |\nafter\n", "| row |\nafter\n"),
            ("| complete |\nlast", "| complete |\nlast\n"),
        ]
        for source, expected in cases:
            with self.subTest(source=source):
                self.assert_filter(source, expected)

    def test_near_match_continuations_are_preserved(self):
        for continuation in ("|", "  |", " | ", "\t|", " | extra"):
            with self.subTest(continuation=continuation):
                source = f"| row\n{continuation}\nend\n"
                self.assert_filter(source, source)

    def test_multiple_input_files_are_processed_as_one_stream(self):
        with tempfile.TemporaryDirectory() as directory:
            first = Path(directory) / "first.md"
            second = Path(directory) / "second.md"
            first.write_text("before\n```json\n{}\n", encoding="utf-8")
            second.write_text("```\n| row\n |\n", encoding="utf-8")
            self.assert_filter("", "before\n| row |\n", first, second)

    def test_explicit_stdin_filename(self):
        self.assert_filter("Examples\nkept\n", "kept\n", "-")


if __name__ == "__main__":
    unittest.main()
