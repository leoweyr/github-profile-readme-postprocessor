import argparse
import json
import os
import sys
from typing import Any, Dict, List

current_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, current_dir)

from updater import ContentUpdater
from processor import TaskProcessor


def main() -> None:
    parser: argparse.ArgumentParser = argparse.ArgumentParser()
    parser.add_argument("--username", required=True)
    parser.add_argument("--readme_template_path", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--tasks", required=True)

    arguments: argparse.Namespace = parser.parse_args()

    try:
        tasks: List[Dict[str, Any]] = json.loads(arguments.tasks)

        if not isinstance(tasks, list):
            raise ValueError("Tasks must be a JSON list.")
    except (json.JSONDecodeError, ValueError) as error:
        print(f"FATAL: Error parsing JSON inputs: {error}.", file=sys.stderr)
        sys.exit(1)

    readme_template_path: str = arguments.readme_template_path

    if not os.path.exists(readme_template_path):
        print(f"FATAL: Template file not found: {readme_template_path}.", file=sys.stderr)
        sys.exit(1)

    with open(readme_template_path, 'r', encoding='utf-8') as file_stream:
        content: str = file_stream.read()

    updater: ContentUpdater = ContentUpdater(content)
    processor: TaskProcessor = TaskProcessor(base_url="http://localhost:8080", username=arguments.username)
    
    any_updated: bool = processor.process_tasks(tasks, updater)

    if any_updated:
        output_path: str = arguments.output

        with open(output_path, 'w', encoding='utf-8') as file_stream:
            file_stream.write(updater.content)

        print(f"README update complete. Output written to {output_path}.")
    else:
        output_path: str = arguments.output

        with open(output_path, 'w', encoding='utf-8') as file_stream:
            file_stream.write(updater.content)

        print(f"Process complete. Output written to {output_path}.")

if __name__ == "__main__":
    main()
