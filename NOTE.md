# NEMS Trainning Project 4 Note

## Normal log of 5G AKA Authentication

```log
2026-04-28T16:25:18+08:00 [INFO][SCP][Detector] HandleUeAuthPostRequest
2026-04-28T16:25:18+08:00 [INFO][SCP][Detector] Handle GenerateAuthDataRequest
2026-04-28T16:25:18+08:00 [INFO][SCP][Detector] Handle QueryAuthSubsData
2026-04-28T16:25:18+08:00 [INFO][SCP][GIN] | 200 |       127.0.0.1 | GET     | /nudr-dr/v1/subscription-data/imsi-208930000000003/authentication-data/authentication-subscription |
2026-04-28T16:25:18+08:00 [INFO][SCP][GIN] | 200 |       127.0.0.1 | POST    | /nudm-ueau/v1/suci-0-208-93-0000-0-0-0000000003/security-information/generate-auth-data |
2026-04-28T16:25:18+08:00 [INFO][SCP][GIN] | 201 |       127.0.0.1 | POST    | /nausf-auth/v1/ue-authentications |
2026-04-28T16:25:18+08:00 [INFO][SCP][Detector] Auth5gAkaComfirmRequest
2026-04-28T16:25:18+08:00 [INFO][SCP][GIN] | 200 |       127.0.0.1 | PUT     | /nausf-auth/v1/ue-authentications/suci-0-208-93-0000-0-0-0000000003/5g-aka-confirmation |

```

