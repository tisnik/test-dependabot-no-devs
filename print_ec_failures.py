#!/usr/bin/env python3

"""Display all Enterprise Contract failures found in EC job logs.

Usage:
print_ec_failures.py log_file_name.log
"""

import json
import sys
from pathlib import Path

PREFIX = "[report-json]"


def main() -> int:
    json_raw_data = ""

    if len(sys.argv) < 2:
        print("Usage: print_ec_failures.py log_file_name.log")
        return 1

    log_file_name = sys.argv[1]

    # read the log file that contains several sections
    # and where each line has some prefix
    try:
        with Path(log_file_name).open("r") as log_file:
            for line in log_file:
                # retrieve just those lines with JSON data
                # and append them
                if line.startswith(PREFIX):
                    stripped = line[len(PREFIX) :].strip()
                    json_raw_data += stripped
    except (FileNotFoundError, PermissionError) as e:
        print("Unable to open input file", e)
        return 1

    if not json_raw_data:
        print("Provided log file does not contain any reports")
        return 1

    json_data = None

    # deserialize JSON
    try:
        json_data = json.loads(json_raw_data)
    except json.JSONDecodeError as e:
        print("Unable to deserialize JSON data:", e)
        return 1

    # serialize into pretty-printed format
    with Path("logs.json").open("w") as json_file:
        json.dump(json_data, json_file, indent=4)

    if "components" not in json_data:
        print("Unable to find components attribute in EC data")
        return 1

    # print summary
    for component in json_data["components"]:
        print(f"Component: {component.get('name', 'unknown component!')}")
        violations = component.get("violations", [])
        warnings = component.get("warnings", [])
        successes = component.get("successes", [])
        total = (
            (num_successes := len(successes))
            + (num_warnings := len(warnings))
            + (num_violations := len(violations))
        )
        print(f"    total checks: {total}")
        print(f"    successes:    {num_successes}")
        print(f"    warnings:     {num_warnings}")
        print(f"    violations:   {num_violations}")
        for i, violation in enumerate(violations):
            print(f"        #{i+1}: {violation.get('msg', 'NO VIOLATION MSG!')}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
