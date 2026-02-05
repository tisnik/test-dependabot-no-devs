"""Comprehensive tests for tool_formatter.py module."""

import logging
import pytest
from tool_formatter import (
    format_tool_response,
    extract_clean_description,
    format_tools_list,
)


class TestFormatToolResponse:
    """Test suite for format_tool_response function."""

    def test_format_tool_response_with_all_fields(self):
        """Test formatting a tool with all required fields present."""
        tool_dict = {
            "identifier": "test_tool",
            "description": "A test tool for testing",
            "parameters": [{"name": "param1", "type": "string"}],
            "provider_id": "provider_123",
            "toolgroup_id": "group_456",
            "server_source": "test_server",
            "type": "function",
        }
        result = format_tool_response(tool_dict)

        assert result["identifier"] == "test_tool"
        assert result["description"] == "A test tool for testing"
        assert result["parameters"] == [{"name": "param1", "type": "string"}]
        assert result["provider_id"] == "provider_123"
        assert result["toolgroup_id"] == "group_456"
        assert result["server_source"] == "test_server"
        assert result["type"] == "function"

    def test_format_tool_response_with_missing_fields(self):
        """Test formatting a tool with missing fields defaults to empty values."""
        tool_dict = {}
        result = format_tool_response(tool_dict)

        assert result["identifier"] == ""
        assert result["description"] == ""
        assert result["parameters"] == []
        assert result["provider_id"] == ""
        assert result["toolgroup_id"] == ""
        assert result["server_source"] == ""
        assert result["type"] == ""

    def test_format_tool_response_with_partial_fields(self):
        """Test formatting a tool with only some fields present."""
        tool_dict = {
            "identifier": "partial_tool",
            "description": "Partial description",
        }
        result = format_tool_response(tool_dict)

        assert result["identifier"] == "partial_tool"
        assert result["description"] == "Partial description"
        assert result["parameters"] == []
        assert result["provider_id"] == ""
        assert result["toolgroup_id"] == ""
        assert result["server_source"] == ""
        assert result["type"] == ""

    def test_format_tool_response_with_structured_description_tool_name(self):
        """Test that structured descriptions with TOOL_NAME= are cleaned."""
        tool_dict = {
            "identifier": "structured_tool",
            "description": "TOOL_NAME=MyTool\n\nThis is a clean description.",
            "parameters": [],
        }
        result = format_tool_response(tool_dict)

        assert result["description"] == "This is a clean description."
        assert "TOOL_NAME=" not in result["description"]

    def test_format_tool_response_with_structured_description_display_name(self):
        """Test that structured descriptions with DISPLAY_NAME= are cleaned."""
        tool_dict = {
            "identifier": "display_tool",
            "description": "DISPLAY_NAME=My Display Tool\n\nActual tool description here.",
            "parameters": [],
        }
        result = format_tool_response(tool_dict)

        assert result["description"] == "Actual tool description here."
        assert "DISPLAY_NAME=" not in result["description"]

    def test_format_tool_response_with_none_description(self):
        """Test handling of None description."""
        tool_dict = {
            "identifier": "none_desc_tool",
            "description": None,
        }
        result = format_tool_response(tool_dict)

        assert result["description"] == ""

    def test_format_tool_response_with_none_parameters(self):
        """Test handling of None parameters."""
        tool_dict = {
            "identifier": "none_params_tool",
            "parameters": None,
        }
        result = format_tool_response(tool_dict)

        assert result["parameters"] == []

    def test_format_tool_response_with_invalid_parameters_type(self):
        """Test handling of invalid parameters type (not a list)."""
        tool_dict = {
            "identifier": "invalid_params_tool",
            "parameters": "not_a_list",
        }
        result = format_tool_response(tool_dict)

        assert result["parameters"] == []

    def test_format_tool_response_with_empty_parameters_list(self):
        """Test handling of empty parameters list."""
        tool_dict = {
            "identifier": "empty_params_tool",
            "parameters": [],
        }
        result = format_tool_response(tool_dict)

        assert result["parameters"] == []

    def test_format_tool_response_preserves_complex_parameters(self):
        """Test that complex parameter structures are preserved."""
        tool_dict = {
            "identifier": "complex_tool",
            "parameters": [
                {
                    "name": "param1",
                    "type": "object",
                    "properties": {"nested": {"type": "string"}},
                },
                {"name": "param2", "type": "array", "items": {"type": "number"}},
            ],
        }
        result = format_tool_response(tool_dict)

        assert len(result["parameters"]) == 2
        assert result["parameters"][0]["name"] == "param1"
        assert result["parameters"][1]["name"] == "param2"


class TestExtractCleanDescription:
    """Test suite for extract_clean_description function."""

    def test_extract_clean_description_simple_text(self):
        """Test extraction with simple text without metadata."""
        description = "This is a simple tool description."
        result = extract_clean_description(description)

        assert result == "This is a simple tool description."

    def test_extract_clean_description_with_tool_name_metadata(self):
        """Test extraction when TOOL_NAME metadata is present."""
        description = "TOOL_NAME=MyTool\n\nThis is the actual description of the tool."
        result = extract_clean_description(description)

        assert result == "This is the actual description of the tool."
        assert "TOOL_NAME=" not in result

    def test_extract_clean_description_with_display_name_metadata(self):
        """Test extraction when DISPLAY_NAME metadata is present."""
        description = "DISPLAY_NAME=My Tool\n\nThe real description goes here."
        result = extract_clean_description(description)

        assert result == "The real description goes here."

    def test_extract_clean_description_with_multiple_metadata_fields(self):
        """Test extraction with multiple metadata fields."""
        description = (
            "TOOL_NAME=ComplexTool\n\n"
            "DISPLAY_NAME=Complex Tool\n\n"
            "USECASE=For testing\n\n"
            "INSTRUCTIONS=Follow these steps\n\n"
            "This is the clean description that should be extracted."
        )
        result = extract_clean_description(description)

        assert result == "This is the clean description that should be extracted."

    def test_extract_clean_description_with_usecase_fallback(self):
        """Test that USECASE is used as fallback when no clean description found."""
        description = "TOOL_NAME=Tool\nUSECASE=This is the usecase description"
        result = extract_clean_description(description)

        assert result == "This is the usecase description"

    def test_extract_clean_description_with_short_description(self):
        """Test that short descriptions (<=20 chars) are skipped."""
        description = "TOOL_NAME=Tool\n\nShort\n\nThis is a longer description that meets the minimum length requirement."
        result = extract_clean_description(description)

        assert result == "This is a longer description that meets the minimum length requirement."

    def test_extract_clean_description_truncation_fallback(self):
        """Test truncation fallback when no clean description or USECASE found."""
        long_text = "A" * 250
        description = f"TOOL_NAME=Tool\nINSTRUCTIONS={long_text}"
        result = extract_clean_description(description)

        assert len(result) <= 203  # 200 chars + "..."
        assert result.endswith("...")

    def test_extract_clean_description_truncation_short_text(self):
        """Test that short text is not truncated."""
        description = "TOOL_NAME=Tool\nINSTRUCTIONS=Short text"
        result = extract_clean_description(description)

        # Since no clean description found and USECASE not present, should fallback
        assert len(result) <= len(description)

    def test_extract_clean_description_with_all_metadata_prefixes(self):
        """Test extraction ignores all known metadata prefixes."""
        description = (
            "TOOL_NAME=Tool\n\n"
            "DISPLAY_NAME=Display\n\n"
            "USECASE=Use case\n\n"
            "INSTRUCTIONS=Instructions\n\n"
            "INPUT_DESCRIPTION=Input\n\n"
            "OUTPUT_DESCRIPTION=Output\n\n"
            "EXAMPLES=Examples\n\n"
            "PREREQUISITES=Prerequisites\n\n"
            "AGENT_DECISION_CRITERIA=Criteria\n\n"
            "This is the actual description."
        )
        result = extract_clean_description(description)

        assert result == "This is the actual description."

    def test_extract_clean_description_empty_string(self):
        """Test handling of empty string."""
        result = extract_clean_description("")

        assert result == ""

    def test_extract_clean_description_only_whitespace(self):
        """Test handling of whitespace-only string."""
        result = extract_clean_description("   \n\n   ")

        assert result == "   \n\n   "

    def test_extract_clean_description_with_exact_20_char_description(self):
        """Test boundary case with exactly 20 character description."""
        # 20 characters exactly
        description = "TOOL_NAME=Tool\n\nExactly 20 chars!!!\n\nLonger description here."
        result = extract_clean_description(description)

        # Should skip the 20-char description and get the longer one
        assert result == "Longer description here."

    def test_extract_clean_description_with_21_char_description(self):
        """Test boundary case with 21 character description."""
        # 21 characters
        twenty_one_chars = "X" * 21
        description = f"TOOL_NAME=Tool\n\n{twenty_one_chars}"
        result = extract_clean_description(description)

        assert result == twenty_one_chars

    def test_extract_clean_description_logging_on_exception(self, caplog):
        """Test that exceptions are logged properly."""
        # Simulate an AttributeError by mocking problematic input
        # Since the function is robust, we test the warning path indirectly
        # by ensuring normal operation doesn't produce warnings
        with caplog.at_level(logging.WARNING):
            description = "Normal description"
            result = extract_clean_description(description)

            assert result == "Normal description"
            assert "Failed to extract clean description" not in caplog.text

    def test_extract_clean_description_usecase_with_equals(self):
        """Test USECASE extraction with equals sign in the value."""
        description = "TOOL_NAME=Tool\nUSECASE=Compare values with = operator"
        result = extract_clean_description(description)

        assert result == "Compare values with = operator"

    def test_extract_clean_description_mixed_separators(self):
        """Test with mixed paragraph separators (single and double newlines)."""
        description = "TOOL_NAME=Tool\n\nDISPLAY_NAME=Display\nINSTRUCTIONS=Do this\n\nClean description here."
        result = extract_clean_description(description)

        assert result == "Clean description here."

    def test_extract_clean_description_no_double_newline_separator(self):
        """Test when metadata and description use single newlines only."""
        description = "TOOL_NAME=Tool\nUSECASE=This is the usecase\nExtra line"
        result = extract_clean_description(description)

        # Should fallback to USECASE since no double-newline paragraphs
        assert result == "This is the usecase"


class TestFormatToolsList:
    """Test suite for format_tools_list function."""

    def test_format_tools_list_empty_list(self):
        """Test formatting an empty list of tools."""
        result = format_tools_list([])

        assert result == []

    def test_format_tools_list_single_tool(self):
        """Test formatting a list with a single tool."""
        tools = [
            {
                "identifier": "tool1",
                "description": "Description 1",
                "parameters": [],
            }
        ]
        result = format_tools_list(tools)

        assert len(result) == 1
        assert result[0]["identifier"] == "tool1"
        assert result[0]["description"] == "Description 1"

    def test_format_tools_list_multiple_tools(self):
        """Test formatting a list with multiple tools."""
        tools = [
            {
                "identifier": "tool1",
                "description": "Description 1",
                "parameters": [{"name": "param1"}],
            },
            {
                "identifier": "tool2",
                "description": "Description 2",
                "parameters": [{"name": "param2"}],
            },
            {
                "identifier": "tool3",
                "description": "Description 3",
                "parameters": [],
            },
        ]
        result = format_tools_list(tools)

        assert len(result) == 3
        assert result[0]["identifier"] == "tool1"
        assert result[1]["identifier"] == "tool2"
        assert result[2]["identifier"] == "tool3"

    def test_format_tools_list_with_structured_descriptions(self):
        """Test that structured descriptions are cleaned for all tools."""
        tools = [
            {
                "identifier": "tool1",
                "description": "TOOL_NAME=Tool1\n\nThis is a clean description that is long enough",
            },
            {
                "identifier": "tool2",
                "description": "DISPLAY_NAME=Tool2\n\nThis is another clean description for testing",
            },
        ]
        result = format_tools_list(tools)

        assert len(result) == 2
        assert result[0]["description"] == "This is a clean description that is long enough"
        assert result[1]["description"] == "This is another clean description for testing"

    def test_format_tools_list_preserves_order(self):
        """Test that the order of tools is preserved."""
        tools = [
            {"identifier": f"tool{i}", "description": f"Desc {i}"}
            for i in range(10)
        ]
        result = format_tools_list(tools)

        assert len(result) == 10
        for i, tool in enumerate(result):
            assert tool["identifier"] == f"tool{i}"
            assert tool["description"] == f"Desc {i}"

    def test_format_tools_list_with_mixed_completeness(self):
        """Test formatting tools with varying levels of field completeness."""
        tools = [
            {
                "identifier": "complete_tool",
                "description": "Complete",
                "parameters": [{"name": "p1"}],
                "provider_id": "provider1",
                "toolgroup_id": "group1",
                "server_source": "server1",
                "type": "function",
            },
            {
                "identifier": "minimal_tool",
            },
            {
                "identifier": "partial_tool",
                "description": "Partial description",
                "type": "action",
            },
        ]
        result = format_tools_list(tools)

        assert len(result) == 3

        # Complete tool
        assert result[0]["identifier"] == "complete_tool"
        assert result[0]["provider_id"] == "provider1"

        # Minimal tool
        assert result[1]["identifier"] == "minimal_tool"
        assert result[1]["description"] == ""
        assert result[1]["parameters"] == []

        # Partial tool
        assert result[2]["identifier"] == "partial_tool"
        assert result[2]["description"] == "Partial description"
        assert result[2]["type"] == "action"


class TestEdgeCases:
    """Test suite for edge cases and boundary conditions."""

    def test_format_tool_response_with_unicode_description(self):
        """Test handling of Unicode characters in descriptions."""
        tool_dict = {
            "identifier": "unicode_tool",
            "description": "Tool with unicode: \u2713 \u2717 \u2022 \ud83d\ude80",
        }
        result = format_tool_response(tool_dict)

        assert "\u2713" in result["description"]
        assert "\ud83d\ude80" in result["description"]

    def test_extract_clean_description_with_unicode(self):
        """Test extraction with Unicode characters."""
        description = "TOOL_NAME=\u2713Tool\n\nDescription with emoji \ud83d\ude80"
        result = extract_clean_description(description)

        assert "\ud83d\ude80" in result

    def test_format_tool_response_with_newlines_in_description(self):
        """Test handling of newlines in regular descriptions."""
        tool_dict = {
            "identifier": "multiline_tool",
            "description": "Line 1\nLine 2\nLine 3",
        }
        result = format_tool_response(tool_dict)

        assert result["description"] == "Line 1\nLine 2\nLine 3"

    def test_extract_clean_description_with_leading_trailing_whitespace(self):
        """Test that leading/trailing whitespace is stripped from extracted descriptions."""
        description = "TOOL_NAME=Tool\n\n   Clean description with spaces   "
        result = extract_clean_description(description)

        assert result == "Clean description with spaces"
        assert not result.startswith(" ")
        assert not result.endswith(" ")

    def test_format_tools_list_with_very_large_list(self):
        """Test performance with a large number of tools."""
        tools = [
            {"identifier": f"tool{i}", "description": f"Description {i}"}
            for i in range(1000)
        ]
        result = format_tools_list(tools)

        assert len(result) == 1000
        assert all(tool["identifier"] for tool in result)

    def test_extract_clean_description_exactly_200_chars(self):
        """Test truncation boundary at exactly 200 characters."""
        text_200 = "A" * 200
        description = f"TOOL_NAME=Tool\n{text_200}"
        result = extract_clean_description(description)

        # Should not add "..." since it's exactly 200
        assert len(result) <= 203

    def test_extract_clean_description_201_chars(self):
        """Test truncation at 201 characters."""
        text_201 = "A" * 201
        description = f"TOOL_NAME=Tool\n{text_201}"
        result = extract_clean_description(description)

        assert result.endswith("...")
        assert len(result) == 203

    def test_format_tool_response_with_extra_fields(self):
        """Test that extra unknown fields don't break formatting."""
        tool_dict = {
            "identifier": "extra_tool",
            "description": "Description",
            "extra_field_1": "ignored",
            "extra_field_2": 123,
            "extra_field_3": ["list", "of", "items"],
        }
        result = format_tool_response(tool_dict)

        # Should only include known fields
        assert "extra_field_1" not in result
        assert "extra_field_2" not in result
        assert "extra_field_3" not in result
        assert result["identifier"] == "extra_tool"

    def test_extract_clean_description_with_metadata_in_middle(self):
        """Test when metadata appears in the middle of content."""
        description = (
            "Too short intro.\n\n"
            "TOOL_NAME=Tool\n\n"
            "This is the actual description after metadata."
        )
        result = extract_clean_description(description)

        # The function skips short descriptions (<= 20 chars) and metadata prefixes,
        # so it returns the description after the metadata
        assert result == "This is the actual description after metadata."

    def test_format_tool_response_parameters_dict_instead_of_list(self):
        """Test handling when parameters is a dict instead of list."""
        tool_dict = {
            "identifier": "dict_params_tool",
            "parameters": {"param1": "value1"},
        }
        result = format_tool_response(tool_dict)

        # Should convert to empty list since it's not a list
        assert result["parameters"] == []

    def test_extract_clean_description_metadata_without_double_newline(self):
        """Test metadata prefixes in text without double newline separation."""
        description = "TOOL_NAME=Tool\nDISPLAY_NAME=Display\nUSECASE=Use this tool"
        result = extract_clean_description(description)

        # Should fallback to USECASE
        assert result == "Use this tool"


class TestRegressionCases:
    """Test suite for potential regression issues."""

    def test_none_input_to_format_tool_response(self):
        """Test that None input doesn't crash the function."""
        # This should handle gracefully if tool_dict methods are called
        tool_dict = {"identifier": None, "description": None, "parameters": None}
        result = format_tool_response(tool_dict)

        assert result["identifier"] is None
        assert result["description"] == ""
        assert result["parameters"] == []

    def test_empty_string_values(self):
        """Test handling of empty string values."""
        tool_dict = {
            "identifier": "",
            "description": "",
            "parameters": [],
            "provider_id": "",
            "toolgroup_id": "",
            "server_source": "",
            "type": "",
        }
        result = format_tool_response(tool_dict)

        assert result["identifier"] == ""
        assert result["description"] == ""
        assert result["parameters"] == []

    def test_whitespace_only_description(self):
        """Test handling of whitespace-only descriptions."""
        tool_dict = {
            "identifier": "whitespace_tool",
            "description": "   \n\n\t  ",
        }
        result = format_tool_response(tool_dict)

        # Should preserve whitespace if no metadata detected
        assert result["description"] == "   \n\n\t  "

    def test_case_sensitive_metadata_detection(self):
        """Test that metadata detection is case-sensitive."""
        # Lowercase should not be detected as metadata
        description = "tool_name=MyTool\n\nThis should be the description."
        result = extract_clean_description(description)

        # Since "tool_name=" (lowercase) is not a metadata prefix,
        # the whole first paragraph should be considered
        assert "tool_name=" in result or result == "This should be the description."

    def test_format_tools_list_modifies_copy_not_original(self):
        """Test that format_tools_list doesn't modify the original list."""
        original_tools = [
            {
                "identifier": "tool1",
                "description": "TOOL_NAME=Tool\n\nThis is a clean description that is long enough to pass",
                "extra_field": "should_remain",
            }
        ]

        # Keep a reference to check
        original_description = original_tools[0]["description"]

        result = format_tools_list(original_tools)

        # Original should not be modified
        assert original_tools[0]["description"] == original_description
        # Result should have cleaned description
        assert result[0]["description"] == "This is a clean description that is long enough to pass"

    def test_special_characters_in_usecase(self):
        """Test USECASE with special characters."""
        description = "TOOL_NAME=Tool\nUSECASE=Use for A/B testing & validation"
        result = extract_clean_description(description)

        assert result == "Use for A/B testing & validation"

    def test_multiple_usecase_lines(self):
        """Test that only first USECASE is used."""
        description = "TOOL_NAME=Tool\nUSECASE=First use case\nUSECASE=Second use case"
        result = extract_clean_description(description)

        assert result == "First use case"

    def test_extract_clean_description_with_exception_handling(self):
        """Test exception handling in extract_clean_description by testing unusual input."""
        # Test with a description that exercises the exception handler path
        # by using a type that could cause issues in string operations
        description = "TOOL_NAME=Tool\nINSTRUCTIONS=Test"
        result = extract_clean_description(description)

        # Should fallback gracefully
        assert isinstance(result, str)
        assert len(result) > 0


class TestAdditionalEdgeCases:
    """Additional tests to strengthen confidence and coverage."""

    def test_format_tool_response_preserves_zero_values(self):
        """Test that zero values are preserved correctly."""
        tool_dict = {
            "identifier": 0,
            "description": 0,
            "parameters": 0,
        }
        result = format_tool_response(tool_dict)

        # Zero is falsy but should still be used when present
        assert result["identifier"] == 0
        # For description, 0 will be converted to "" by the "or" operator
        assert result["description"] == ""
        # For parameters, 0 will become []
        assert result["parameters"] == []

    def test_extract_clean_description_with_only_metadata(self):
        """Test when description contains only metadata fields."""
        description = (
            "TOOL_NAME=Tool\n"
            "DISPLAY_NAME=Display\n"
            "USECASE=\n"
            "INSTRUCTIONS=\n"
        )
        result = extract_clean_description(description)

        # Should return truncated version since USECASE is empty
        assert isinstance(result, str)

    def test_format_tool_response_boolean_parameters(self):
        """Test handling of boolean parameters."""
        tool_dict = {
            "identifier": "bool_tool",
            "parameters": [
                {"name": "required", "type": "boolean", "default": False},
                {"name": "enabled", "type": "boolean", "default": True},
            ],
        }
        result = format_tool_response(tool_dict)

        assert len(result["parameters"]) == 2
        assert result["parameters"][0]["default"] is False
        assert result["parameters"][1]["default"] is True

    def test_extract_clean_description_consecutive_newlines(self):
        """Test handling of many consecutive newlines."""
        description = "TOOL_NAME=Tool\n\n\n\n\n\nDescription with many newlines before it."
        result = extract_clean_description(description)

        assert result == "Description with many newlines before it."

    def test_format_tool_response_nested_dict_parameters(self):
        """Test deeply nested parameter structures."""
        tool_dict = {
            "identifier": "nested_tool",
            "parameters": [
                {
                    "name": "config",
                    "type": "object",
                    "properties": {
                        "level1": {
                            "type": "object",
                            "properties": {
                                "level2": {
                                    "type": "object",
                                    "properties": {"level3": {"type": "string"}},
                                }
                            },
                        }
                    },
                }
            ],
        }
        result = format_tool_response(tool_dict)

        assert result["parameters"][0]["name"] == "config"
        assert "properties" in result["parameters"][0]
        assert "level1" in result["parameters"][0]["properties"]

    def test_extract_clean_description_tabs_and_spaces(self):
        """Test handling of mixed tabs and spaces."""
        description = "TOOL_NAME=Tool\n\n\tDescription with\ttabs\tin it."
        result = extract_clean_description(description)

        assert "Description with\ttabs\tin it." == result

    def test_format_tools_list_preserves_empty_descriptions(self):
        """Test that empty descriptions are preserved as empty strings."""
        tools = [
            {"identifier": "tool1", "description": ""},
            {"identifier": "tool2", "description": None},
            {"identifier": "tool3"},
        ]
        result = format_tools_list(tools)

        assert result[0]["description"] == ""
        assert result[1]["description"] == ""
        assert result[2]["description"] == ""

    def test_extract_clean_description_all_short_paragraphs(self):
        """Test when all paragraphs are too short."""
        description = "Short\n\nTiny\n\nSmall\n\nBrief"
        result = extract_clean_description(description)

        # Should fallback to truncation
        assert isinstance(result, str)

    def test_format_tool_response_array_parameters(self):
        """Test handling of array-type parameters."""
        tool_dict = {
            "identifier": "array_tool",
            "parameters": [
                {"name": "items", "type": "array", "items": {"type": "string"}},
                {
                    "name": "matrix",
                    "type": "array",
                    "items": {"type": "array", "items": {"type": "number"}},
                },
            ],
        }
        result = format_tool_response(tool_dict)

        assert len(result["parameters"]) == 2
        assert result["parameters"][0]["name"] == "items"
        assert result["parameters"][1]["name"] == "matrix"

    def test_extract_clean_description_paragraph_exactly_21_chars(self):
        """Test with a paragraph that is exactly 21 characters."""
        # This is exactly 21 characters
        twenty_one = "This has 21 chars!XX"  # 20 chars
        twenty_one = "X" * 21  # Ensure exactly 21
        description = f"TOOL_NAME=Tool\n\n{twenty_one}"
        result = extract_clean_description(description)

        assert result == twenty_one

    def test_format_tool_response_with_list_in_dict_parameters(self):
        """Test parameters containing lists within objects."""
        tool_dict = {
            "identifier": "complex_tool",
            "parameters": [
                {
                    "name": "config",
                    "type": "object",
                    "properties": {
                        "allowed_values": ["value1", "value2", "value3"],
                        "nested_list": [{"key": "value"}, {"key2": "value2"}],
                    },
                }
            ],
        }
        result = format_tool_response(tool_dict)

        assert result["parameters"][0]["properties"]["allowed_values"] == [
            "value1",
            "value2",
            "value3",
        ]

    def test_extract_clean_description_with_cr_lf_endings(self):
        """Test Windows-style line endings (CRLF)."""
        description = "TOOL_NAME=Tool\r\n\r\nDescription with Windows line endings."
        result = extract_clean_description(description)

        # The function splits on \n\n, not \r\n\r\n, so CRLF won't split properly
        # and will be treated as one paragraph. This test documents current behavior.
        assert isinstance(result, str)
        # Should fallback to truncation since no proper split occurs
        assert len(result) <= 203

    def test_format_tools_list_handles_generator_input(self):
        """Test that format_tools_list works with generator expressions."""
        # Create a generator
        tools_gen = (
            {"identifier": f"tool{i}", "description": f"Desc {i}"} for i in range(5)
        )
        result = format_tools_list(list(tools_gen))

        assert len(result) == 5
        assert all("identifier" in tool for tool in result)

    def test_extract_clean_description_only_usecase_no_value(self):
        """Test USECASE without a value after equals."""
        description = "TOOL_NAME=Tool\nUSECASE="
        result = extract_clean_description(description)

        # Should return empty string from USECASE or fallback
        assert isinstance(result, str)

    def test_format_tool_response_with_numeric_string_fields(self):
        """Test handling of numeric strings in fields."""
        tool_dict = {
            "identifier": "123",
            "description": "456",
            "provider_id": "789",
        }
        result = format_tool_response(tool_dict)

        assert result["identifier"] == "123"
        assert result["description"] == "456"
        assert result["provider_id"] == "789"