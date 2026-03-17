# Cross-Chain Oracle: Heterogeneous Blockchain Pivot

## Project Overview
This project is a Go-based lightweight cross-chain oracle service.
It facilitates data interoperability between Hyperledger Fabric (Authority Chain) 
and FISCO BCOS (Business Chain) through state relay and verification.

## Core Architecture
- Authority Zone (Fabric): pkicert.go, spbft.go
- The Hub (Oracle): Go Oracle Service (oracle_main.go)
- Business Zone (FISCO BCOS): CertOracle.sol

## Technical Capabilities
- Authority Sourcing: Direct connection to Fabric Gateway v2.5 using x509 identity.
- Relay Mechanism: Linux system pipe for state injection into FISCO BCOS.
- Decision Logic: Integrated S-PBFT reputation assessment for data validation.
- Automation: Built-in cleanup scripts for Docker containers and zombie processes.

## Debugging Records
- Time Drift: Resolved x509 validation failures caused by snapshot restoration.
- Port Conflict: Fixed port 20200 occupation using physical cleanup scripts.
- SDK Migration: Implemented Fabric Gateway SDK to replace legacy versions.

## Roadmap
- Phase 1: MVP Prototype (API-triggered mode) - [Complete]
- Phase 2: Event Listener (Event-driven mode) - [Current]
- Phase 3: Native Go SDK Integration
- Phase 4: Visualization Dashboard

Built by Fan & AI Collaborator. 2026.
