"""Unit tests for utils/schema_dumper module."""

from typing import Any
import pytest

from utils.schema_dumper import recursive_update


def test_empty_input():
    original: dict[str, Any] = {}
    expected: dict[str, Any] = {}
    result = recursive_update(original)
    assert result == expected
    # ensure a new dict is returned, not the same object
    assert result is not original


def test_no_change_for_simple_schema():
    original: dict[str, Any] = {
        "type": "string",
        "maxLength": 10,
    }
    expected = original.copy()
    result = recursive_update(original)
    assert result == expected
    # ensure a new dict is returned, not the same object
    assert result is not original


def test_recursive_recurse_into_subdicts():
    original = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "exclusiveMinimum": 0},
        },
    }
    expected = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum": 0},
        },
    }
    result = recursive_update(original)
    assert result == expected


def test_exclusive_minimum_handling_positive_value():
    original = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "exclusiveMinimum": 100},
        },
    }
    expected = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum": 100},
        },
    }
    result = recursive_update(original)
    assert result == expected


def test_exclusive_minimum_handling_negative_value():
    original = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "exclusiveMinimum": -100},
        },
    }
    expected = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum": -100},
        },
    }
    result = recursive_update(original)
    assert result == expected


def test_anyof_with_null_transformed_to_nullable():
    original = {
        "anyOf": [
            {"type": "string"},
            {"type": "null"},
        ]
    }
    expected = {
        "type": "string",
        "nullable": True,
    }
    result = recursive_update(original)
    assert result == expected


def test_anyof_list_with_more_complex_first_entry():
    original = {
        "anyOf": [
            {"type": "array", "items": {"type": "integer"}},
            {"type": "null"},
        ]
    }
    expected = {
        "type": "array",
        "nullable": True,
    }
    result = recursive_update(original)
    assert result == expected


def test_anyof_not_transformed_when_conditions_not_met():
    # various conditions where anyOf should be left unchanged
    cases = [
        {"anyOf": "not-a-list"},
        {"anyOf": [{"type": "string"}]},  # length < 2
        {"anyOf": [{"notype": "x"}, {"type": "null"}]},  # first item missing type
        {"anyOf": [{"type": "string"}, {"type": "number"}]},  # second not null
    ]
    for original in cases:
        result = recursive_update(original)
        assert result == original


def test_mixed_keys_preserve_order_like_behavior():
    # verify that keys other than handled ones are preserved
    original = {
        "exclusiveMinimum": 5,
        "anyOf": [
            {"type": "integer"},
            {"type": "null"},
        ],
        "description": "example",
    }
    # exclusiveMinimum should become minimum; anyOf -> type+nullable and description preserved
    expected = {
        "minimum": 5,
        "type": "integer",
        "nullable": True,
        "description": "example",
    }
    result = recursive_update(original)
    assert result == expected


def test_deeply_nested_anyof_and_exclusiveMinimum():
    original = {
        "level1": {
            "level2": {
                "anyOf": [
                    {"type": "object", "properties": {"x": {"type": "string"}}},
                    {"type": "null"},
                ],
                "exclusiveMinimum": 1,
            }
        }
    }
    expected = {
        "level1": {
            "level2": {
                "type": "object",
                "nullable": True,
                "minimum": 1,
            }
        }
    }
    result = recursive_update(original)
    assert result == expected


def test_preserve_other_types_and_lists():
    original = {
        "type": "object",
        "required": ["a", "b"],
        "properties": {
            "a": {"type": "integer"},
            "b": {"anyOf": [{"type": "boolean"}, {"type": "null"}]},
        },
    }
    expected = {
        "type": "object",
        "required": ["a", "b"],
        "properties": {
            "a": {"type": "integer"},
            "b": {"type": "boolean", "nullable": True},
        },
    }
    assert recursive_update(original) == expected


def test_handles_none_values():
    original = {"key": None}
    # None values should be preserved
    assert recursive_update(original) == original


def test_handles_empty_values():
    original = {"key": []}
    # None values should be preserved
    assert recursive_update(original) == original


def test_anyof_with_additional_fields_on_first_item():
    original = {
        "anyOf": [
            {"type": "string", "format": "email", "maxLength": 50},
            {"type": "null"},
        ]
    }
    expected = {
        "type": "string",
        "nullable": True,
    }
    assert recursive_update(original) == expected
