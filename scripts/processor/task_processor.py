import sys
from typing import Any, Dict, List, Optional

import requests

from ..updater import ContentUpdater


class TaskProcessor:
    def __init__(self, base_url: str, username: str):
        self.base_url: str = base_url
        self.username: str = username

    def process_tasks(self, tasks: List[Dict[str, Any]], updater: ContentUpdater) -> bool:
        any_updated: bool = False

        for task in tasks:
            anchor: Optional[str] = task.get("anchor")
            endpoint: Optional[str] = task.get("endpoint")
            parameters: Dict[str, Any] = task.get("params", {})

            if not anchor or not endpoint:
                print(f"FATAL: Invalid task configuration: {task}.", file=sys.stderr)
                sys.exit(1)

            if not endpoint.endswith('markdown'):
                print(f"FATAL: Endpoint must end with 'markdown': {endpoint}.", file=sys.stderr)
                sys.exit(1)

            if 'username' not in parameters:
                parameters['username'] = self.username

            request_url: str = f"{self.base_url}{endpoint}"
            print(f"Processing anchor '{anchor}' -> {request_url} with parameters {parameters}.")

            try:
                response: requests.Response = requests.get(request_url, params=parameters)
                response.raise_for_status()
                fetched_content: str = response.text

                updated, _ = updater.update_section(anchor, fetched_content)

                if updated:
                    print(f"Successfully updated section for '{anchor}'.")
                    any_updated = True
                else:
                    print(f"Warning: Anchor marker '{anchor}' (and its _END counterpart) not found in README.")

            except requests.RequestException as error:
                print(f"FATAL: Error fetching data from {request_url}: {error}.", file=sys.stderr)
                sys.exit(1)
        
        return any_updated
