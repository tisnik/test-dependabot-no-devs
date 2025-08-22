#!/usr/bin/env python3

import os

import ast
from pathlib import Path

directories = "src/auth/",


for path in Path("src").rglob("*"):
    if path.is_dir():
        directory = path
        cwd = os.getcwd()
        os.chdir(directory)

        try:
            with open("README.md", "w", encoding="utf-8", newline="\n") as indexfile:
                print(f"# List of source files stored in `{directory}` directory", file=indexfile)
                print("", file=indexfile)
                files = sorted(os.listdir())

                for file in files:
                    if file.endswith(".py"):
                        print(f"## [{file}]({file})", file=indexfile)
                        with open(file, "r", encoding="utf-8") as fin:
                            source = fin.read()
                        try:
                            mod = ast.parse(source)
                            doc = ast.get_docstring(mod)
                        except SyntaxError:
                            doc = None
                        if doc:
                            print(doc.splitlines()[0], file=indexfile)
                        print(file=indexfile)
        finally:
            os.chdir(cwd)
