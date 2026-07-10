#!/usr/bin/env python3

"""Filter to fix the generated documentation for configuration schema."""

import fileinput


def align_markdown_table(md: str) -> str:
    """
    Align the columns of a Markdown table using consistent cell widths.
    
    Parameters:
        md (str): Markdown table content to format.
    
    Returns:
        str: The table with padded cells and uniformly formatted row separators.
    """
    lines = [line.strip() for line in md.strip().splitlines() if line.strip()]
    rows = [[cell.strip() for cell in line.strip("|").split("|")] for line in lines]

    widths = [max(len(row[i]) for row in rows) for i in range(len(rows[0]))]

    out = []
    for r, row in enumerate(rows):
        padded = [cell.ljust(widths[i]) for i, cell in enumerate(row)]
        out.append("| " + " | ".join(padded) + " |")
    return "\n".join(out)


def print_formatted_table(table_content: str) -> None:
    """
    Print an aligned Markdown table followed by a blank line.
    
    Parameters:
    	table_content (str): Markdown table content to align and print.
    """
    print(align_markdown_table(table_content))
    print()


def is_table_row(current_line: str) -> bool:
    """Determine whether a line represents a Markdown table row.
    
    Parameters:
        current_line (str): The line to inspect.
    
    Returns:
        bool: `true` if the line is non-empty and starts with `|`, `false` otherwise.
    """
    return len(current_line) > 0 and current_line[0] == "|"


def process_documentation(lines: list[str]) -> None:
    """
    Filter documentation lines and print the transformed output.
    
    Parameters:
        lines (list[str]): Documentation lines to process. Table blocks are aligned, API and base URL sections are removed, and the components heading is renamed.
    """
    in_table = False
    table_content = ""
    base_url_count = 0

    for current_line in lines:
        stripped = current_line.rstrip("\r\n")

        # skip lines below "Base URL" section
        if base_url_count > 0:
            base_url_count -= 1
            continue

        if stripped == "# 🛠️ APIs":
            continue

        if stripped == "## 🌍 Base URL":
            # skip next 5 lines
            base_url_count = 5
            continue

        if stripped == "# 📋 Components":
            print("# 📋 Configuration schema")
            continue

        # skip empty lines
        if len(current_line) == 0:
            print()
            continue

        if in_table:
            if is_table_row(current_line):
                table_content += current_line
            else:
                in_table = False
                print_formatted_table(table_content)
        else:
            if is_table_row(current_line):
                in_table = True
                table_content = current_line
            else:
                print(stripped)

    # make sure that table at the end of the document is rendered correctly
    if in_table:
        print_formatted_table(table_content)


lines = list(fileinput.input())
process_documentation(lines)
