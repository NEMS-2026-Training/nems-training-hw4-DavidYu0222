package detector

import (
	"net/http"
	"strings"
	"encoding/hex"

	"github.com/free5gc/http_wrapper"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/scp/consumer"
	"github.com/free5gc/scp/logger"
)

const (
	ERR_MANDATORY_ABSENT = "Mandatory type is absent"
	ERR_MISS_CONDITION   = "Miss condition"
	ERR_VALUE_INCORRECT  = "Unexpected value is received"
)

// 4.AMF to AUSF
func HandleAuth5gAkaComfirmRequest(request *http_wrapper.Request) *http_wrapper.Response {
	logger.DetectorLog.Infof("Auth5gAkaComfirmRequest")
	updateConfirmationData := request.Body.(models.ConfirmationData)
	ConfirmationDataResponseID := request.Params["authCtxId"]

	// NOTE: The request from AMF is guaranteed to be correct

	// TODO: Send request to target NF by setting correct uri (forward)
	targetNfUri := "http://127.0.0.9:8000"	// AUSF's uri

	response, problemDetails, err := consumer.SendAuth5gAkaConfirmRequest(targetNfUri, ConfirmationDataResponseID, &updateConfirmationData)

	// TODO: Check IEs in response body is correct
	// Response OpenAPI: models.ConfirmationDataResponse
	logger.DetectorLog.Infof("CheckAuth5gAkaConfirmResponse")

	// Compute Kseaf for 5G AKA
	kseaf := retrieveKseaf(
		CurrentAuthProcedure.kausf,
		"6C",
		[]byte(CurrentAuthProcedure.servingNetworkName),
	)
	kseafHex := hex.EncodeToString(kseaf)

	// logger.DetectorLog.Infof("DEBUG Kseaf in response:    '%s'", response.Kseaf)
	// logger.DetectorLog.Infof("DEBUG kseafHex computed:    '%s'", kseafHex)

	// Check Kseaf
	if response.Kseaf == "" {
		logger.DetectorLog.Errorln("ConfirmationDataResponse.Kseaf: " + ERR_MISS_CONDITION)
		response.Kseaf = kseafHex
	} else if response.Kseaf != kseafHex {
		logger.DetectorLog.Errorln("ConfirmationDataResponse.Kseaf: " + ERR_VALUE_INCORRECT)
		response.Kseaf = kseafHex
	}

	if response != nil {
		return http_wrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
	logger.DetectorLog.Errorln(err)
	problemDetails = &models.ProblemDetails{
		Status: http.StatusForbidden,
		Cause:  "UNSPECIFIED",
	}
	return http_wrapper.NewResponse(http.StatusForbidden, nil, problemDetails)
}

// 1. AMF to AUSF
func HandleUeAuthPostRequest(request *http_wrapper.Request) *http_wrapper.Response {
	logger.DetectorLog.Infof("HandleUeAuthPostRequest")
	updateAuthenticationInfo := request.Body.(models.AuthenticationInfo)

	// NOTE: The request from AMF is guaranteed to be correct

	// TODO: Send request to target NF by setting correct uri
	targetNfUri := "http://127.0.0.9:8000"	// AUSF's uri

	response, respHeader, problemDetails, err := consumer.SendUeAuthPostRequest(targetNfUri, &updateAuthenticationInfo)

	// TODO: Check IEs in response body is correct
	// Response OpenAPI: models.UeAuthenticationCtx
	logger.DetectorLog.Infof("CheckUeAuthResonse")

	// Retrieve HXRES*
	hxresStar := CurrentAuthProcedure.hxresStar
    hxresStarHex := hex.EncodeToString(hxresStar)

	// Check Hxres*
    if authDataMap, ok := response.Var5gAuthData.(map[string]interface{}); ok {
        // Safely extract HxresStar from the map
        respHxresStar, hasHxresStar := authDataMap["hxresStar"].(string)

		// logger.DetectorLog.Infof("DEBUG HxresStar in response: '%s'", hxres)
		// logger.DetectorLog.Infof("DEBUG hxresStarHex computed: '%s'", hxresStarHex)

        if !hasHxresStar || respHxresStar == "" {
            logger.DetectorLog.Errorln("UeAuthenticationCtx.Av5gAka.HxresStar: " + ERR_MANDATORY_ABSENT)
            authDataMap["hxresStar"] = hxresStarHex
        } else if respHxresStar != hxresStarHex {
            logger.DetectorLog.Errorln("UeAuthenticationCtx.Av5gAka.HxresStar: " + ERR_VALUE_INCORRECT)
            authDataMap["hxresStar"] = hxresStarHex
        }

        // Re-assign back if necessary (maps are passed by reference, but just to be safe)
        response.Var5gAuthData = authDataMap
    } else {
        logger.DetectorLog.Errorf("Var5gAuthData is not map[string]interface{}")
    }

	if response != nil {
		return http_wrapper.NewResponse(http.StatusCreated, respHeader, response)
	} else if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
	logger.DetectorLog.Errorln(err)
	problemDetails = &models.ProblemDetails{
		Status: http.StatusForbidden,
		Cause:  "UNSPECIFIED",
	}
	return http_wrapper.NewResponse(http.StatusForbidden, nil, problemDetails)
}

// 2. AUSF to UDM
func HandleGenerateAuthDataRequest(request *http_wrapper.Request) *http_wrapper.Response {
	logger.DetectorLog.Infoln("Handle GenerateAuthDataRequest")

	authInfoRequest := request.Body.(models.AuthenticationInfoRequest)
	supiOrSuci := request.Params["supiOrSuci"]

	// TODO: Check IEs in request body is correct

	// Derive serving network name from SUCI
	parts := strings.Split(supiOrSuci, "-")	 // Format: "suci-0-<MCC>-<MNC>-..."  e.g. "suci-0-208-93-0000-0-0-0000000003"
	mcc := parts[2]                          // "208"
	mnc := parts[3]                          // "93"
	if len(mnc) == 2 {
		mnc = "0" + mnc                      // pad to 3 digits → "093"
	}
	servingNetworkName := "5G:mnc" + mnc + ".mcc" + mcc + ".3gppnetwork.org" // Format is "5G:mnc<MNC>.mcc<MCC>.3gppnetwork.org"
	CurrentAuthProcedure.servingNetworkName = servingNetworkName

	// Check ServingNetworkName
	if authInfoRequest.ServingNetworkName == "" {
		logger.DetectorLog.Errorln("AuthenticationInfoRequest.ServingNetworkName: " + ERR_MANDATORY_ABSENT)
		authInfoRequest.ServingNetworkName = servingNetworkName
	}

	// TODO: Send request to target NF by setting correct uri
	targetNfUri := "http://127.0.0.3:8000"	// UDM's uri

	response, problemDetails, err := consumer.SendGenerateAuthDataRequest(targetNfUri, supiOrSuci, &authInfoRequest)

	// TODO: Check IEs in response body is correct
	// Response OpenAPI: models.AuthenticationInfoResult
	logger.DetectorLog.Infoln("CheckGenerateAuthDataResponse")

	// Retrieve necessary parameters for derivation
	xres, sqnXorAk, ck, ik, autn := retrieveBasicDeriveFactor(&CurrentAuthProcedure.AuthSubsData, response.AuthenticationVector.Rand)
	//_, _, _, _, _ = xres, sqnXorAk, ck, ik, autn

	autnHex := hex.EncodeToString(autn)

	// Compute XRES*
	randBytes, _ := hex.DecodeString(response.AuthenticationVector.Rand)
	xresStar := retrieveXresStar(
		append(ck, ik...),                              // Key = CK || IK
		"6B",                                           // FC = 0x6B
		[]byte(authInfoRequest.ServingNetworkName),     // P0 = SN name
		randBytes,                                      // P1 = RAND
		xres,                                           // P2 = XRES
	)
	xresStarHex := hex.EncodeToString(xresStar)

	// Compute HXRES*
	hxresStar := retrieveHxresStar(append(randBytes, xresStar...))	// P0 || P1 for HXRES* is RAND || XRES*
	CurrentAuthProcedure.hxresStar = hxresStar

	// Compute Kausf for 5G AKA
	kausf := retrieve5GAkaKausf(
		append(ck, ik...),                              // Key = CK || IK
		"6A",                                           // FC = 0x6A
		[]byte(authInfoRequest.ServingNetworkName),     // P0 = SN name
		sqnXorAk,                                      // P1 = SQN ⊕ AK
	)
	CurrentAuthProcedure.kausf = kausf
	kausfHex := hex.EncodeToString(kausf)

	// logger.DetectorLog.Infof("DEBUG Autn in response:     '%s'", response.AuthenticationVector.Autn)
	// logger.DetectorLog.Infof("DEBUG autnHex computed:     '%s'", autnHex)
	// logger.DetectorLog.Infof("DEBUG XresStar in response: '%s'", response.AuthenticationVector.XresStar)
	// logger.DetectorLog.Infof("DEBUG xresStarHex computed: '%s'", xresStarHex)
	// logger.DetectorLog.Infof("DEBUG Kausf in response:    '%s'", response.AuthenticationVector.Kausf)
	// logger.DetectorLog.Infof("DEBUG kausfHex computed:    '%s'", kausfHex)

	// Check AUTN
	if response.AuthenticationVector.Autn == "" {
		logger.DetectorLog.Errorln("AuthenticationInfoResult.AuthenticationVector.Autn: " + ERR_MANDATORY_ABSENT)
		response.AuthenticationVector.Autn = autnHex
	} else if response.AuthenticationVector.Autn != autnHex {
		logger.DetectorLog.Errorln("AuthenticationInfoResult.AuthenticationVector.Autn: " + ERR_VALUE_INCORRECT)
		response.AuthenticationVector.Autn = autnHex
	}

	// Check XRES*
	if response.AuthenticationVector.XresStar == "" {
		logger.DetectorLog.Errorln("AuthenticationInfoResult.AuthenticationVector.XresStar: " + ERR_MANDATORY_ABSENT)
		response.AuthenticationVector.XresStar = xresStarHex
	} else if response.AuthenticationVector.XresStar != xresStarHex {
		logger.DetectorLog.Errorln("AuthenticationInfoResult.AuthenticationVector.XresStar: " + ERR_VALUE_INCORRECT)
		response.AuthenticationVector.XresStar = xresStarHex
	}

	// Check Kausf
	if response.AuthenticationVector.Kausf == "" {
		logger.DetectorLog.Errorln("AuthenticationInfoResult.AuthenticationVector.Kausf: " + ERR_MANDATORY_ABSENT)
		response.AuthenticationVector.Kausf = kausfHex
	} else if response.AuthenticationVector.Kausf != kausfHex {
		logger.DetectorLog.Errorln("AuthenticationInfoResult.AuthenticationVector.Kausf: " + ERR_VALUE_INCORRECT)
		response.AuthenticationVector.Kausf = kausfHex
	}

	if response != nil {
		return http_wrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
	logger.DetectorLog.Errorln(err)
	problemDetails = &models.ProblemDetails{
		Status: http.StatusForbidden,
		Cause:  "UNSPECIFIED",
	}
	return http_wrapper.NewResponse(http.StatusForbidden, nil, problemDetails)
}

// 3. UDM to UDR
func HandleQueryAuthSubsData(request *http_wrapper.Request) *http_wrapper.Response {
	logger.DetectorLog.Infof("Handle QueryAuthSubsData")

	ueId := request.Params["ueId"]

	// TODO: Send request to correct NF by setting correct uri
	targetNfUri := "http://127.0.0.4:8000" // UDR's uri

	response, problemDetails, err := consumer.SendAuthSubsDataGet(targetNfUri, ueId)

	// NOTE: The response from UDR is guaranteed to be correct
	CurrentAuthProcedure.AuthSubsData = *response

	if response != nil {
		return http_wrapper.NewResponse(http.StatusOK, nil, response)
	} else if problemDetails != nil {
		return http_wrapper.NewResponse(int(problemDetails.Status), nil, problemDetails)
	}
	logger.DetectorLog.Errorln(err)
	problemDetails = &models.ProblemDetails{
		Status: http.StatusForbidden,
		Cause:  "UNSPECIFIED",
	}
	return http_wrapper.NewResponse(http.StatusForbidden, nil, problemDetails)
}
