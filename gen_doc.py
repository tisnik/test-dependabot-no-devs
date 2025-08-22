#!/usr/bin/env python3

import os

directories = "src/auth/",

for directory in directories:
    cwd = os.getcwd()
    os.chdir(directory)

    with open("README.md", "w") as indexfile:
        print(f"# List of source files stored in `{directory}` directory", file=indexfile)
        print("", file=indexfile)
        files = sorted(os.listdir())

        for file in files:
            if file.endswith(".py"):
                print(f"## [{file}]({file})", file=indexfile)
                with open(file, "r") as fin:
                    for line in fin:
                        line = line.strip()
                        if line.startswith('"""') and line.endswith('"""'):
                            line=line[3:][:-3]
                            print(line, file=indexfile)
                            break
                print("", file=indexfile)
    os.chdir(cwd)
