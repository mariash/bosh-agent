package handler

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

const (
	mbusHandlerLogTag       = "MBus Handler"
	responseMaxLengthErrMsg = "Response exceeded maximum allowed length"
	UnlimitedResponseLength = -1
)

func PerformHandler(request Request, handler Func, maxResponseLength int, logger boshlog.Logger) ([]byte, error) {
	response := handler(request)
	if response == nil {
		logger.Info(mbusHandlerLogTag, "Nil response returned from handler")
		return []byte{}, nil
	}

	respJSON, err := marshalResponse(response, maxResponseLength, logger)
	if err != nil {
		return respJSON, err
	}

	logger.Info(mbusHandlerLogTag, "Responding")
	logger.DebugWithDetails(mbusHandlerLogTag, "Payload", respJSON)

	return respJSON, nil
}

func buildErrorWithJSON(msg string, logger boshlog.Logger) ([]byte, error) {
	response := NewExceptionResponse(bosherr.Error(msg))

	respJSON, err := json.Marshal(response)
	if err != nil {
		return respJSON, bosherr.WrapError(err, "Marshalling JSON")
	}

	logger.Info(mbusHandlerLogTag, "Building error", msg)

	return respJSON, nil
}

func marshalResponse(response Response, maxResponseLength int, logger boshlog.Logger) ([]byte, error) {
	respJSON, err := json.Marshal(response)
	if err != nil {
		logger.Error(mbusHandlerLogTag, "Failed to marshal response: %s", err.Error())
		return respJSON, bosherr.WrapError(err, "Marshalling JSON response")
	}

	if maxResponseLength == UnlimitedResponseLength {
		return respJSON, nil
	}

	if len(respJSON) > maxResponseLength {
		respJSON, err = json.Marshal(response.Shorten())
		if err != nil {
			logger.Error(mbusHandlerLogTag, "Failed to marshal response: %s", err.Error())
			return respJSON, bosherr.WrapError(err, "Marshalling JSON response")
		}
	}

	if len(respJSON) > maxResponseLength {
		respJSON, err = buildErrorWithJSON(responseMaxLengthErrMsg, logger)
		if err != nil {
			logger.Error(mbusHandlerLogTag, "Failed to build 'max length exceeded' response: %s", err.Error())
			return respJSON, bosherr.WrapError(err, "Building error")
		}
	}

	return respJSON, nil
}
