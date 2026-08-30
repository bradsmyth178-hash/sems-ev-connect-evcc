# SEMS EV CONNECT Smart Charging

This repository contains the Smart Charging controller used by SEMS EV CONNECT.
It is based on EVCC `0.314.5` and carries SEMS EV CONNECT release `0.1.0`.

The custom integration path uses Home Assistant entities for both sides of the
energy flow:

- the charger adapter uses the SEMS EV CONNECT Home Assistant integration;
- the grid, solar and battery adapters use Home Assistant's GoodWe entities.

This build does not use EVCC's native GoodWe charger driver. It does not modify
or bypass EVCC sponsor-token checks.

The release image supports `linux/amd64` and `linux/arm64` and is published to
`ghcr.io/bradsmyth178-hash/sems-ev-connect-evcc`.

## Release verification

Run:

```sh
go test ./util/homeassistant ./util/templates
go test ./charger -run '^TestTemplates/sems-ev-connect$' -count=1
go test ./meter -run '^TestTemplates/sems-goodwe-home-energy' -count=1
```

The release workflow also checks that the maintained branch still descends
from the pinned EVCC `0.314.5` tag before an upstream update is accepted.
