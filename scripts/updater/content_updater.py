from typing import Tuple


class ContentUpdater:
    def __init__(self, content: str):
        self.content: str = content

    def update_section(self, anchor: str, new_section_content: str) -> Tuple[bool, str]:
        if anchor in self.content:
            self.content = self.content.replace(anchor, new_section_content)

            return True, self.content

        return False, self.content
