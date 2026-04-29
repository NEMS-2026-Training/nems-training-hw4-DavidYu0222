# NEMS Training Project 4

###### DavidYu0222

## SCP Workflow

![image](https://hackmd.io/_uploads/HJeUyEkAWx.png)

ref: https://www.techplayon.com/5g-authentication-and-key-management-aka-procedure/


## Bugs

### 1. Empty servingNetworkName

#### In `HandleGenerateAuthDataRequest`

Request OpenAPI: [AuthenticationInfoRequest](https://pkg.go.dev/github.com/free5gc/openapi/models#AuthenticationInfoRequest)

#### Description
From wireshark,
![螢幕擷取畫面 2026-04-28 171709](https://hackmd.io/_uploads/SkkMIGCTZx.png)

`ServingNetworkName` is empty.
This error is `Mandatory type is absent`.

In [3GPP TS 29.503 (UDM)](https://www.etsi.org/deliver/etsi_ts/129500_129599/129503/16.04.00_60/ts_129503v160400p.pdf) Page 212, the Fully-Qualified-Type-Name is `AuthenticationInfoRequest.ServingNetworkName`.

#### Solution
Fill the correct `ServingNetworkName`
(5G:mnc093.mcc208.3gppnetwork.org)

---

### 2. Incorrect HXRES*

#### In `HandleUeAuthPostRequest`

Response OpenAPI: [UeAuthenticationCtx](https://pkg.go.dev/github.com/free5gc/openapi/models#UeAuthenticationCtx), [Av5gAka](https://pkg.go.dev/github.com/free5gc/openapi/models#Av5gAka)

#### Description
The `HXRES*` from AUSF is not correct, which cause the AMF show the message `HRES* Validation Failure`.
This error is `Unexpected value is received`.
In [3GPP TS 29.509 (AUSF)](https://www.etsi.org/deliver/etsi_ts/129500_129599/129509/16.06.00_60/ts_129509v160600p.pdf) Page 31~32, the Fully-Qualified-Type-Name is `UEAuthenticationCtx.Av5gAka.HxresStar`

#### Solution
Derive the correct `HXRES*` from `Rand` and `XRES*`.
The `HXRES*` derivation function is in [3GPP TS 33.501 (Security)](https://www.etsi.org/deliver/etsi_ts/133500_133599/133501/16.03.00_60/ts_133501v160300p.pdf) Page 192.

---

### 3. Missing Kseaf 

#### In `HandleAuth5gAkaComfirmRequest`

Response OpenAPI: [ConfirmationDataResponse](https://pkg.go.dev/github.com/free5gc/openapi/models#ConfirmationDataResponse)

#### Description
The `Kseaf` from AUSF is empty, which cause the AMF show the message `Cause: Security mode rejected, upspecified`.
This error is `Miss condition`.
In [3GPP TS 29.509 (AUSF)](https://www.etsi.org/deliver/etsi_ts/129500_129599/129509/16.06.00_60/ts_129509v160600p.pdf) Page 32, the Fully-Qualified-Type-Name is `ConfirmationDataResponse.Kseaf`

#### Solution
Derive the correct `Kseaf` from `Kausf` and `ServingNetworkName`.
The Kseaf derivation function is in [3GPP TS 33.501 (Security)](https://www.etsi.org/deliver/etsi_ts/133500_133599/133501/16.03.00_60/ts_133501v160300p.pdf) Page 192.
