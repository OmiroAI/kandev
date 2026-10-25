"""Capture prepared dependency identities during the isolated QA image build."""
import hashlib
import json
from pathlib import Path

root=Path('/opt/kandev/source')
paths=[root/'apps/backend/go.mod',root/'apps/backend/go.sum',
       root/'apps/pnpm-lock.yaml',root/'apps/pnpm-workspace.yaml']
paths.extend(p for p in (root/'apps').rglob('package.json')
             if not any(part in ('node_modules','.git') for part in p.relative_to(root).parts))
Path('/opt/kandev/locks.json').write_text(json.dumps({
    str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(paths)},indent=2)+'\n')
