What is this?
===

This project aims to be a tool for helping with DOCSIS proactive network maintenance (PNM).

Features
---
* PostgreSQL database for history data with automatic table partitioning
* Collects downstream and upstream history directly from the modem via SNMP
* Collects modem information from CMTS via SNMP
* RESTful API

Requirements
---
* At least PostgreSQL 11 because we rely on some table partitioning features of
  Postgres 11

Planned features
---
* Upstream pre equalization analysis
* Grafana integration for up- and downstream monitoring
* DOCSIS 3.1 support

Tested on
---
* EuroDOCSIS
* Cisco cBR-8
* A variety of cable modems, notably a variety of AVM Fritz!Box modems and
  Teleste RF amplifiers.
  
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
