package repository

import (
	"errors"
	"fmt"
	"github.com/Deansquirrel/goTextHandle/global"
	"github.com/Deansquirrel/goToolMSSql"
	"strings"
)

import log "github.com/Deansquirrel/goToolLog"

const (
	sqlGetRcodeFromHxDb = "" +
		"select rcode " +
		"from cfv4accwxmprel A " +
		"	inner join cfv4accstat B ON A.accid = B.accid " +
		"where B.acccellno = ?"
)

type HxDB struct {
}

func (hx *HxDB) GetRCode(phone string) ([]string, error) {
	repHxDbConnInfo := HxDBConnInfo{}
	connInfo, err := repHxDbConnInfo.GetConnInfo(global.SysConfig.UpdateInfo.UserKey)
	if err != nil {
		return nil, err
	}
	log.Debug(connInfo)
	connInfo = strings.Trim(connInfo, " ")
	if connInfo == "" {
		errMsg := "get hxDbConnStr error: empty"
		log.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	c := common{}
	connList, err := c.GetDBConfigFromStr(connInfo)
	if err != nil {
		return nil, err
	}
	if len(connList) < 1 {
		errMsg := fmt.Sprintf("tran hxDbConnStr to conn config list error: %s", "list is empty")
		log.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	rList := make([]string, 0)
	for _, dbConfig := range connList {
		list, err := hx.getRCodeByDBConfig(dbConfig, phone)
		if err != nil {
			return nil, err
		}
		for _, rCode := range list {
			rList = append(rList, rCode)
		}
	}
	return rList, nil
}

func (hx *HxDB) getRCodeByDBConfig(dbConfig *goToolMSSql.MSSqlConfig, phone string) ([]string, error) {
	c := common{}
	rows, err := c.GetRowsBySQL(dbConfig, sqlGetRcodeFromHxDb, phone)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var rList = make([]string, 0)
	var rCode string
	for rows.Next() {
		err := rows.Scan(&rCode)
		if err != nil {
			return nil, err
		}
		rList = append(rList, rCode)
	}
	if rows.Err() != nil {
		log.Error(err.Error())
		return nil, rows.Err()
	}
	return rList, nil
}
