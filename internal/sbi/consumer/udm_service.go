package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ShouheiNishi/openapi5g/models"
	udm_ueau "github.com/ShouheiNishi/openapi5g/udm/ueau"
	utils_error "github.com/ShouheiNishi/openapi5g/utils/error"
	ausf_context "github.com/free5gc/ausf/internal/context"
	oldModels "github.com/free5gc/openapi/models"
	"github.com/free5gc/util/httpclient"
	"github.com/samber/lo"
)

type nudmService struct {
	consumer *Consumer

	ueauMu sync.RWMutex

	ueauClients map[string]*udm_ueau.ClientWithResponses
}

func (s *nudmService) getUdmUeauClient(uri string) (*udm_ueau.ClientWithResponses, error) {
	if uri == "" {
		return nil, fmt.Errorf("empty URI")
	}
	s.ueauMu.RLock()
	client, ok := s.ueauClients[uri]
	if ok {
		s.ueauMu.RUnlock()
		return client, nil
	}

	editor, err := ausf_context.GetSelf().GetTokenRequestEditor(context.TODO(), models.ServiceNameNudmUeau, models.NFTypeUDM)
	if err != nil {
		s.ueauMu.RUnlock()
		return nil, err
	}

	uriFull := uri + "/nudm-ueau/v1"
	client, err = udm_ueau.NewClientWithResponses(uriFull, func(c *udm_ueau.Client) error {
		c.Client = httpclient.GetHttpClient(uriFull)
		return nil
	}, udm_ueau.WithRequestEditorFn(editor))
	if err != nil {
		s.ueauMu.RUnlock()
		return nil, err
	}

	s.ueauMu.RUnlock()
	s.ueauMu.Lock()
	defer s.ueauMu.Unlock()
	s.ueauClients[uri] = client
	return client, nil
}

func (s *nudmService) SendAuthResultToUDM(
	id string,
	authType models.AuthType,
	success bool,
	servingNetworkName, udmUrl string,
) error {
	timeNow := time.Now()

	self := s.consumer.Context()

	authEvent := models.AuthEvent{
		TimeStamp:          timeNow,
		AuthType:           authType,
		Success:            success,
		ServingNetworkName: servingNetworkName,
		NfInstanceId:       self.GetSelfID(),
	}

	client, err := s.getUdmUeauClient(udmUrl)
	if err != nil {
		return err
	}

	rsp, err := client.ConfirmAuthWithResponse(context.Background(), id, authEvent)
	if err != nil || rsp.StatusCode() != http.StatusCreated {
		return utils_error.ExtractAndWrapOpenAPIError("udm_ueau.ConfirmAuthWithResponse", rsp, err)
	}
	return nil
}

func (s *nudmService) GenerateAuthDataApi(
	udmUrl string,
	supiOrSuci string,
	oldAuthInfoReq oldModels.AuthenticationInfoRequest,
) (*oldModels.AuthenticationInfoResult, error, *models.ProblemDetails) {
	var newAuthInfoReq models.AuthenticationInfoRequest
	if buf, err := json.Marshal(oldAuthInfoReq); err != nil {
		return nil, err, nil
	} else if err := json.Unmarshal(buf, &newAuthInfoReq); err != nil {
		return nil, err, nil
	} else {
		newResult, errOrig, pd := s.realGenerateAuthDataApi(udmUrl, models.SupiOrSuci(supiOrSuci), newAuthInfoReq)
		if newResult == nil {
			return nil, errOrig, pd
		}
		if buf, err := json.Marshal(newResult); err != nil {
			return nil, err, nil
		} else {
			var oldResult oldModels.AuthenticationInfoResult
			if err := json.Unmarshal(buf, &oldResult); err != nil {
				return nil, err, nil
			} else {
				return &oldResult, errOrig, nil
			}
		}
	}
}

func (s *nudmService) realGenerateAuthDataApi(
	udmUrl string,
	supiOrSuci models.SupiOrSuci,
	authInfoReq models.AuthenticationInfoRequest,
) (*models.AuthenticationInfoResult, error, *models.ProblemDetails) {
	client, err := s.getUdmUeauClient(udmUrl)
	if err != nil {
		return nil, err, nil
	}

	rsp, err := client.GenerateAuthDataWithResponse(context.TODO(), supiOrSuci, authInfoReq)
	if err != nil || rsp.JSON200 == nil {
		err = utils_error.ExtractAndWrapOpenAPIError("udm_ueau.GenerateAuthDataWithResponse", rsp, err)
		return nil, err, lo.ToPtr(utils_error.ErrorToProblemDetails(err))
	}
	return rsp.JSON200, nil, nil
}