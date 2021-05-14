What is this?
===

This project aims to be a tool for helping with DOCSIS proactive network maintenance (PNM).

See the [API specification](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/Sedl/docsis-pnm/master/apispec.yaml) to get an overview what is possible.

Features
---
* RESTful API
* PostgreSQL database for history data with automatic table partitioning
* Automatic deletion of data older than 14 days
* Collects downstream history directly from the modem via SNMP
* Collects upstream history directly from the CMTS
* Per modem traffic accounting
* Builtin TFTP server
* Get downstream OFDM MER from modem via TFTP
* Appropriate caching headers for caching history data older than one hour. See /nginx-cache and /docker-compose.yml for more details.
* Cloud native [Docker images](https://hub.docker.com/r/stephan256/docsis-pnm)

Requirements
---
* At least PostgreSQL 11 because we rely on some table partitioning features of
  Postgres 11

Planned features
---
* Upstream pre equalization analysis
* Grafana integration for up- and downstream monitoring
* DOCSIS 3.1 support

Tested with
---
* EuroDOCSIS
* Cisco cBR-8
* Harmonic CableOS
  
Tested modems
---
* AVM FRITZ!Box Cable
  * 6360
  * 6490
  * 6591
  * 6660
* Arris
  * TM822S
* Thomson THG571
* Technicolor TC4400
* Kathrein/BKTel
  * DCV8400
  * TDS 1030
  * TDS 10

Tested amplifiers
---
* Teleste
  * AC9000
  * AC9100
  * AC3200
* Kathrein TVM100
