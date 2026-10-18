# Remaining machine and release work

## Real captures and fault injection

- Repair or start local WSL Ubuntu, then run `sh scripts/capture.sh wsl2`.
- Capture clean `ubuntu-24.04`, `fedora`, `arch`, and `ubuntu-server` boots.
- Capture `in-progress` immediately after login to verify `systemd-analyze` behaviour.
- Add an `expect.txt` to every clean capture with `shape=normal`.
- On fresh VM snapshots, create every fault listed in `docs/build_plan.md` section 7.1 and capture it under `testdata/faults/`.
- Verify exact journal message IDs/text, default timeout values, rotational-disk detection, and JSON support on each systemd version.

## Validation on machines

- Run `go test ./...`, `go test -race ./...`, parser fuzz targets, and `go vet ./...` on Windows and Linux.
- Run `bootwhy --from testdata/<fixture>`, `--json`, and `explain 1` for every fixture; inspect exit codes and golden output.
- Compare every displayed time with the matching `systemd-analyze` command on each VM.
- Run clean and fault boots through the accuracy checks. No medium/high-confidence result may contradict the injected cause.
- Test ordinary and root journal access, reverse dependency collection, a physical SSD machine, and a physical HDD machine.

## Deferred release work

Phase 7 packaging, licence, README, man page, release hosting, tags, AUR publishing, and fresh-VM package installation are intentionally out of scope for this implementation pass.
