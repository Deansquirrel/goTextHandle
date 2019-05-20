package repository

import (
	"errors"
	"fmt"
	"github.com/Deansquirrel/goToolMSSql"
)

import log "github.com/Deansquirrel/goToolLog"

const (
	sqlGetConnInfo = "" +
		"select [conn] " +
		"from ProxyConfigDB"
	sqlGetAppIdAndDomain = "" +
		"select [Authorization_appid],[Authorization_appid_domain] " +
		"from [Authorization_appid_Config] " +
		"where IsDelete = 0 " +
		"	and IsStop = 0 " +
		"	and [Authorization_appid] != 'tongyong'"
)

type ProxyConfigDB struct {
}

//func (p *ProxyConfigDB) Test() {
//	list, err := p.GetAppIdAndDomain()
//	if err != nil {
//		log.Error(err.Error())
//	} else {
//		for k, v := range list {
//			log.Debug(fmt.Sprintf("%s %s", k, v))
//		}
//	}
//}

//返回appid和domain的对应关系 map
func (p *ProxyConfigDB) GetAppIdAndDomain() (map[string]string, error) {
	config, err := p.getConnConfig()
	if err != nil {
		return nil, err
	}
	c := common{}
	rows, err := c.GetRowsBySQL(config, sqlGetAppIdAndDomain)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()
	var appid, domain string
	r := make(map[string]string)
	for rows.Next() {
		err = rows.Scan(&appid, &domain)
		if err != nil {
			log.Error(fmt.Sprintf("read rows error: %s", err.Error()))
			return nil, err
		}
		log.Debug(fmt.Sprintf("%s - %s", domain, appid))
		r[domain] = appid
	}
	if rows.Err() != nil {
		log.Error(fmt.Sprintf("read rows error: %s", rows.Err().Error()))
		return nil, err
	}
	log.Debug(fmt.Sprintf("map length %d", len(r)))
	return r, nil
}

func (p *ProxyConfigDB) getConnConfig() (dbConfig *goToolMSSql.MSSqlConfig, err error) {
	c := common{}
	config := c.GetConfigDBConfig()
	rows, err := c.GetRowsBySQL(config, sqlGetConnInfo)
	if err != nil {
		//log.Error(fmt.Sprintf("get conn info err: %s",err.Error()))
		return
	}
	defer func() {
		_ = rows.Close()
	}()
	var connStr string
	for rows.Next() {
		err = rows.Scan(&connStr)
		if err != nil {
			log.Error(fmt.Sprintf("read rows err: %s", err.Error()))
			return
		}
	}
	if rows.Err() != nil {
		err = rows.Err()
		log.Error(fmt.Sprintf("read rows err: %s", err.Error()))
		return
	}
	log.Debug(connStr)
	dbConfigList, err := c.GetDBConfigFromStr(connStr)
	if err != nil {
		return
	}
	if len(dbConfigList) > 0 {
		dbConfig = dbConfigList[0]
	} else {
		err = errors.New("config is null")
	}
	return
}
