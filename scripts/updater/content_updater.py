import re
from typing import List, Tuple, Dict, Any, Match, Pattern
from datetime import datetime


class ContentUpdater:
    def __init__(self, content: str):
        self.content: str = content

    def update_section(self, anchor: str, new_section_content: str) -> Tuple[bool, str]:
        if anchor in self.content:
            self.content = self.content.replace(anchor, new_section_content)

            return True, self.content

        return False, self.content

    def sort_latest_activity_blocks(self) -> None:
        pattern: Pattern[str] = re.compile(r'(<!-- LATEST_ACTIVITY: (.*?) -->.*?<!-- LATEST_ACTIVITY_END -->)', re.DOTALL)
        
        matches: List[Match[str]] = list(pattern.finditer(self.content))
        
        if not matches:
            return

        # Extract blocks and their timestamps.
        blocks: List[Dict[str, Any]] = []
        match: Match[str]

        for match in matches:
            full_block: str = match.group(0)
            timestamp_str: str = match.group(2).strip()
            timestamp: datetime

            try:
                # Parse RFC3339 timestamp (e.g., 2024-03-13T14:30:00Z).
                # Using fromisoformat for basic ISO8601 support. 
                # Note: Python 3.11+ supports 'Z', for older versions replace 'Z' with '+00:00'.
                timestamp = datetime.fromisoformat(timestamp_str.replace('Z', '+00:00'))
            except ValueError:
                # If timestamp parsing fails, use min date to push to bottom.
                timestamp = datetime.min
            
            blocks.append({
                'timestamp': timestamp,
                'content': full_block
            })

        # Sort blocks by timestamp descending.
        blocks.sort(key=lambda x: x['timestamp'], reverse=True)
        
        new_content_parts: List[str] = []
        last_index: int = 0
        
        i: int
        match: Match[str]

        for i, match in enumerate(matches):
            # Append content before the match.
            new_content_parts.append(self.content[last_index:match.start()])
            
            # Append the sorted block corresponding to this position.
            # i.e., the first match slot gets the 1st sorted block (newest).
            new_content_parts.append(blocks[i]['content'])
            
            last_index = match.end()
        
        # Append remaining content.
        new_content_parts.append(self.content[last_index:])
        
        self.content = "".join(new_content_parts)
