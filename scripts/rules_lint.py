### rules_lint.py
"""Basic linter for Windsurf workspace rule files.
Fails if any file exceeds 6 000 characters or misses an activation tag.
"""
from __future__ import annotations
import pathlib, re, sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
RULES_DIR = ROOT / '.windsurf' / 'rules'
MAX_LEN = 6000
ACTIVATION_RE = re.compile(r'<glob|Always On|Manual|Model Decision', re.I)


def lint() -> int:
    rc = 0
    for path in RULES_DIR.rglob('*.md'):
        txt = path.read_text('utf8')
        if len(txt) > MAX_LEN:
            print(f'{path} exceeds {MAX_LEN} chars', file=sys.stderr)
            rc = 1
        if not ACTIVATION_RE.search(txt):
            print(f'{path} missing activation marker', file=sys.stderr)
            rc = 1
    return rc

if __name__ == '__main__':
    sys.exit(lint())