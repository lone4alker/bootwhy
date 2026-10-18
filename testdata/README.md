# Fixtures

Run `scripts/capture.sh <name>` from each target Linux system to create a
fixture. Add an `expect.txt` beside every capture, for example:

```
shape=normal
```

Required clean captures are `ubuntu-24.04`, `fedora`, `arch`, `wsl2`,
`ubuntu-server`, and `in-progress`. Fault-injection captures belong below
`testdata/faults/`.

The local WSL capture could not be collected on 2026-10-18 because its service
failed to start with `Wsl/Service/E_UNEXPECTED`.
