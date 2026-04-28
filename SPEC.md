# SPEC.md — Anomaly Detector in 5G Core Network

## Overview

This project implements an **anomaly detector** embedded in the Service Communication Proxy (SCP) of a 5G core network. The detector monitors, validates, and recovers authentication messages flowing between Network Functions (NFs) during the 5G AKA authentication procedure.

The testbed is built on **free5GC** (open-source 5G core, Release 15) and **UERANSIM** (open-source 5G UE/RAN simulator). A modified version of free5GC is provided for the project.

---

## Learning Goals

- Understand the 5G AKA authentication procedure
- Understand 5G Service-Based Architecture (SBA) operation
- Work with free5GC and Go programming
- Read and interpret 3GPP specifications

---

## System Architecture

### 5G Testbed Components

| Component | Role |
|-----------|------|
| UE | User Equipment (simulated by UERANSIM) |
| (R)AN / gNodeB | Radio Access Network (simulated by UERANSIM) |
| AMF | Access and Mobility Management Function |
| AUSF | Authentication Server Function |
| UDM / ARPF / SIDF | Unified Data Management |
| UDR | Unified Data Repository |
| SMF | Session Management Function |
| UPF | User Plane Function |
| PCF | Policy Control Function |
| NRF | Network Repository Function |
| NSSF | Network Slice Selection Function |
| SCP | Service Communication Proxy *(project focus)* |

### SCP Internal Architecture

```
┌─────────────────────────┐
│           SCP           │
│  ┌────────┐ ┌────────┐  │
│  │ Proxy  │ │Detector│  │
│  └────────┘ └────────┘  │
└─────────────────────────┘
         ▲
        SBI
```

- **Proxy**: Forwards SBI messages to the target NF and to the Detector.
- **Detector**: Inspects messages for anomalies and recovers incorrect content before forwarding.

---

## NF IP Configuration

| NF | Address |
|----|---------|
| AMF | `127.0.0.18:8000` |
| AUSF | `127.0.0.9:8000` |
| UDM | `127.0.0.3:8000` |
| UDR | `127.0.0.4:8000` |
| SCP (Detector) | `127.0.0.113:8000` |

---

## 5G AKA Authentication Flow

### Initiation
1. UE sends **Registration Request** (SUCI or 5G-GUTI) to SEAF/AMF.
2. AMF sends `Nausf_UEAuthentication_Authenticate Request` (SUCI or SUPI, SN) to AUSF.
3. AUSF sends `Nudm_UEAuthentication_Get Request` (SUCI or SUPI, SN) to UDM.
4. UDM performs SUCI-to-SUPI de-concealment and selects the authentication method.
5. UDM fetches UE subscription data from UDR via `Nudr_DataRepository Request/Response`.
6. UDM returns a **5G HE AV** (and optionally SUPI, AKMA indication) to AUSF.

### Challenge & Verification
7. AUSF generates the AV, stores XRES*, computes HXRES*, and responds to AMF with `5G SE AV (RAND, AUTN, HXRES*)`.
8. AMF sends **NAS Auth-Req** (RAND, AUTN, ngKSI, ABBA) to UE.
9. UE calculates RES* and replies with **NAS Auth-Resp (RES*)**.
10. AMF sends `Nausf_UEAuthentication_Authenticate Request (RES*)` to AUSF.
11. AUSF computes HRES* and compares it to HXRES*; on success returns `Result, [SUPI], KSEAF`.

---

## SBI Endpoints Handled by the SCP Detector

| Endpoint | Direction |
|----------|-----------|
| `{apiRoot}/nausf-auth/v1/ue-authentications` | AMF → AUSF |
| `{apiRoot}/nudm-ueau/v1/{supiOrSuci}/securityinformation/generate-authdata` | AUSF → UDM |
| `{apiRoot}/nudr-dr/subscription-data/{ueid}/authenticationdata/authentication-subscription` | UDM → UDR |
| `{apiRoot}/nausf-auth/v1/ue-authentications/{authCtxId}/5g-aka-confirmation` | AMF → AUSF |

### Trusted Sources (guaranteed correct by spec)
- All messages **from AMF** and **from UDR**
- IE `rand` from UDM
- IEs `ausfInstanceId`, `authResult`, `_links` from AUSF

---

## Project Tasks

### Task I — Message Forwarding (50%)
Forward authentication messages to the correct target NFs via the SCP.

- Parse incoming SBI requests by URI path.
- Set `targetNfUri` to the appropriate NF address.
- Complete the `// TODO: Send request to target NF` sections in `handler.go`.

### Task II — Anomaly Detection (30%)
Verify the correctness of all Information Elements (IEs) in authentication messages. Three categories of anomalies must be detected:

| Anomaly Type | Error Message |
|---|---|
| Missing mandatory IE | `"mandatory type is absent"` |
| Condition not satisfied for a conditional IE | `"missing conditions"` |
| IE contains an incorrect value | `"unexpected value is received"` |

### Task III — Message Recovery (20%)
Use provided helper functions (in `derivation.go`) to recompute and restore correct IE values when an anomaly is detected.

---

## Codebase

### Files to Modify

| File | Purpose |
|------|---------|
| `~/project1/scp/detector/handler.go` | Main logic: forwarding, checking, recovering messages |
| `~/project1/scp/detector/derivation.go` | SUCI decryption and key derivation functions |
| `~/project1/scp/detector/context.go` | Global `AuthProcedureInfo` struct for shared state across handlers |

### Build Requirement
- Write a `Makefile` that compiles the SCP detector.
- Place the final binary at `~/project1/scp/bin/scp`.

---

## Output Format

All anomaly reports must use the **error-level logger**:

```go
logger.DetectorLog.Errorln()
logger.DetectorLog.Errorf()
```

### Format

```
<Fully-Qualified-Type-Name>: <Error message>
```

- `<Fully-Qualified-Type-Name>`: Dot-separated path from the top-level message IE type down to the specific member IE (case-insensitive).
- `<Error message>`: One of the three defined error strings.

### Sample Output

```
[ERRO][SCP][Detector] AuthenticationInfoRequest.ServingNetworkName: Mandatory type is absent
[ERRO][SCP][Detector] UeAuthenticationCtx.Av5gAka.HxresStar: Unexpected value is received
[ERRO][SCP][Detector] ConfirmationDataResponse.Kseaf: Miss condition
```

---

## Environment Setup

### VM Configuration
- Download the provided VM image.
- Create **two VMs**: one for UE/RAN, one for the 5G Core.
- Recommended hypervisor: VirtualBox
- Login: `mns2022` / `mns2022`

| VM | Network Interfaces |
|----|--------------------|
| 5GC VM | Interface1: NAT, Interface2: Host-Only |
| UE VM | Interface1: Host-Only |

### UERANSIM Config (`~/UERANSIM/config/free5gc-gnb.yaml`)

Set the following fields to match your local VM IPs:

```yaml
ngapIp: <UE VM Host-Only IP>   # e.g. 192.168.56.102
gtpIp:  <UE VM Host-Only IP>   # e.g. 192.168.56.102

amfConfigs:
  - address: <5GC VM Host-Only IP>  # e.g. 192.168.56.101
    port: 38412
```

---

## Running & Testing

### Run Modes

| Command | Description |
|---------|-------------|
| `~/project1/run.sh` | Normal 5GC (no SCP) |
| `~/project1/run.sh --buggy` | Buggy 5GC (no SCP) — auth fails |
| `~/project1/run.sh --with-scp` | Normal 5GC with SCP |
| `~/project1/run.sh --with-scp --buggy` | Buggy 5GC with SCP — use to test your detector |

### Start UE/RAN (from UE VM)

```bash
# Start gNodeB
~/UERANSIM/build/nr-gnb -c ~/UERANSIM/config/free5gc-gnb.yaml

# Start UE
~/UERANSIM/build/nr-ue -c ~/UERANSIM/config/free5gc-ue.yaml
```

### Verify Success

```bash
# UE should be able to reach the Internet
ping -I uesimtun0 8.8.8.8
```

### Observing Behavior
- Capture packets on the **loopback interface** with Wireshark to inspect SBI messages.
- SCP detector output and 5GC logs are printed to the terminal.

---

## Reference Specifications

| Spec | Relevant Sections | Topic |
|------|-------------------|-------|
| 3GPP TS 33.501 | §6.1, Annex A | Security architecture, UE auth flows, key derivation |
| 3GPP TS 29.503 | §5.4, §6.3 | UDM services, message structure |
| 3GPP TS 29.509 | §5.2, §6.1 | AUSF service message structure |
| 3GPP TS 29.505 | §5.2.2, §5.4 | UDR response structure for authentication |
| 3GPP TS 29.571 | — | Common SBI data types |
| 3GPP TS 29.501 | §5.2 | SBI API definition conventions |

---
