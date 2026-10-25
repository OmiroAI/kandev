# Offline Kodi QA

Build the preparation image with network access, then run its entry point with
networking disabled against the candidate source. No provider or publisher
credentials are needed.

```sh
docker build -f .kodi/Dockerfile -t kandev-qa:local .
docker run --rm --network none --read-only --cap-drop ALL \
  --security-opt no-new-privileges --cpus 2 --memory 8g --pids-limit 256 \
  --user "$(id -u):$(id -g)" --tmpfs /tmp:rw,nosuid,nodev,size=256m \
  --mount "type=bind,src=$PWD,dst=/workspace" \
  -e HOME=/tmp -e PYTHONDONTWRITEBYTECODE=1 \
  kandev-qa:local python3 .kodi/qa.py
```

Use a fresh disposable source export: the runner rejects pre-existing dependency
or build-cache directories. Dependency hashes must match the preparation image.
Changing locks requires rebuilding that image, never enabling network during QA.
The build uses a Dockerfile-specific ignore file and requires Docker BuildKit.
Legacy builders should use a committed-source context with dependencies and Git
metadata excluded; the repository's release-oriented root `.dockerignore` omits
tests/docs and is unsuitable for this QA build.

The scoped baseline covers Go accounting/Stats regressions across task/analytics packages, Stats unit tests, web
typecheck, locale validation, and specification checks. PostgreSQL parity and
Playwright browser flows require separately configured test fixtures. This
baseline does not claim deployment, provider-account or desktop acceptance.

Install a dedicated private independent-QA profile after review. Kodi's generic
Python QA image and 50 MiB/link-free source transport cannot run this repository.
The dedicated profile needs `archive_max_bytes: 536870912`,
`expanded_max_bytes: 536870912`, `materialize_internal_links: true`, and
`memory_mb: 8192`. Those values belong in trusted operator configuration, never
in a task payload. Point the coordinator's disabled repository entry at that
dedicated suite and use the matching bounded sender limit after helper review.
Keep the repository disabled until exact-commit independent acceptance,
different-model review, GitHub checks and human approval are recorded. Building
this image does not authorize publishing it or changing the running dashboard.
