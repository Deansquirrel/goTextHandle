package service

import (
	"encoding/json"
	"fmt"
	log "github.com/Deansquirrel/goToolLog"
	"github.com/kataras/iris/core/errors"
	"io/ioutil"
	"net/http"
)

type crmDz struct {
	serverAddress string
}

func NewCrmDz(serverAddress string) *crmDz {
	return &crmDz{
		serverAddress: serverAddress,
	}
}

type UpdateWxMembershipNoResponse struct {
	DbCommitted bool   `json:"dbCommitted"`
	WxCardNo    string `json:"wxcardno"`
	Description string `json:"description"`
}

const apiRouterUpdateWxMembershipNo = "/ApiV1/CRMYwDzService/UpdateWeixinMembershipno"

func (s *crmDz) UpdateWxMembershipNo(rCode string) error {
	log.Debug(rCode)
	log.Debug(apiRouterUpdateWxMembershipNo)
	api := s.serverAddress +
		apiRouterUpdateWxMembershipNo +
		fmt.Sprintf("?wxcardcode=%s", rCode)
	log.Debug(api)

	resp, _ := http.Get(api)
	defer func() {
		_ = resp.Body.Close()
	}()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug(string(body))
	var rep UpdateWxMembershipNoResponse
	err := json.Unmarshal(body, &rep)
	if err != nil {
		return err
	}
	if !rep.DbCommitted {
		errMsg := fmt.Sprintf("update error: %s", rep.Description)
		log.Error(errMsg)
		return errors.New(errMsg)
	}
	log.Warn(fmt.Sprintf("%s is updated %s", rCode, rep.WxCardNo))
	return nil
}
