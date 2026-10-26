"""Offline acceptance entry point; dependency preparation belongs in the QA image."""
import hashlib
import json
import os
import shutil
import subprocess
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
PREPARED=Path('/opt/kandev')


def verify_locks(root, manifest):
    expected_locks=json.loads(manifest.read_text())
    actual_manifests={str(p.relative_to(root)) for p in (root/'apps').rglob('package.json')
                      if 'node_modules' not in p.relative_to(root).parts}
    prepared_manifests={name for name in expected_locks if name.endswith('/package.json')}
    if actual_manifests != prepared_manifests:
        raise RuntimeError('Dependency identity mismatch: workspace manifests')
    for name, expected in expected_locks.items():
        path=root/name
        if not path.is_file() or hashlib.sha256(path.read_bytes()).hexdigest()!=expected:
            raise RuntimeError('Dependency identity mismatch: '+name)


def run(*args, cwd=ROOT):
    subprocess.run(args,cwd=cwd,check=True)


def main():
    verify_locks(ROOT,PREPARED/'locks.json')
    temporary=ROOT/'.kodi-tmp'
    temporary.mkdir()
    os.environ.update({'TMPDIR':str(temporary),'GOPROXY':'off','GOTOOLCHAIN':'local','GOSUMDB':'off',
        'GOMODCACHE':'/root/go/pkg/mod','GOCACHE':str(ROOT/'.kodi-go-cache'),
        'GOFLAGS':'-trimpath -tags=fts5 -p=1','GOMAXPROCS':'2',
        'GOGC':'50','GOMEMLIMIT':'1536MiB',
        'NODE_OPTIONS':'--max-old-space-size=6144','CI':'true'})
    # Dependencies are copied only after the complete prepared identity check.
    prepared_source=PREPARED/'source'
    for modules in prepared_source.glob('apps/**/node_modules'):
        if 'node_modules' in modules.relative_to(prepared_source).parts[:-1]: continue
        target=ROOT/modules.relative_to(prepared_source)
        if target.exists() or target.is_symlink():
            raise RuntimeError('QA workspace must not contain node_modules')
        shutil.copytree(modules,target,symlinks=True)
    shutil.copytree(PREPARED/'go-build',ROOT/'.kodi-go-cache')
    run('python3','-m','unittest','discover','-s','.kodi','-p','test_qa.py','-v')
    run('python3','scripts/list-docs.py','validate')
    run('python3','scripts/lint-spec-files.py','--all')
    run('go','test','-count=1','-run','Usage|Stats|Activity','./internal/task/models','./internal/task/dto',
        './internal/task/repository/sqlite','./internal/task/service',
        './internal/task/handlers','./internal/task/usage','./internal/analytics/...',
        './internal/persistence/requiredstores',
        cwd=ROOT/'apps/backend')
    run('pnpm','test','app/stats','lib/api/domains/stats-api.test.ts',cwd=ROOT/'apps/web')
    run('pnpm','run','typecheck',cwd=ROOT/'apps/web')
    run('pnpm','run','i18n:check',cwd=ROOT/'apps/web')


if __name__=='__main__': main()
