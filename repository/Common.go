package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Deansquirrel/goTextHandle/global"
	"github.com/Deansquirrel/goToolCommon"
	"github.com/Deansquirrel/goToolMSSql"
	"strconv"
	"strings"
)

import log "github.com/Deansquirrel/goToolLog"

type common struct {
}

//获取配置库连接配置
func (c *common) GetConfigDBConfig() *goToolMSSql.MSSqlConfig {
	return &goToolMSSql.MSSqlConfig{
		Server: global.SysConfig.DB.Server,
		Port:   global.SysConfig.DB.Port,
		DbName: global.SysConfig.DB.DbName,
		User:   global.SysConfig.DB.User,
		Pwd:    global.SysConfig.DB.Pwd,
	}
}

func (c *common) GetDBConfigFromStr(connStr string) ([]*goToolMSSql.MSSqlConfig, error) {
	strList := strings.Split(connStr, "|")
	strList = goToolCommon.ClearBlock(strList)
	if len(strList)%5 != 0 {
		err := errors.New(fmt.Sprintf("config num error,exp 5,act %d", len(strList)))
		return nil, err
	}

	rList := make([]*goToolMSSql.MSSqlConfig, 0)
	configLength := int(len(strList) / 5)
	for i := 0; i < configLength; i++ {
		port, err := strconv.Atoi(strList[i*5+1])
		if err != nil {
			err = errors.New(fmt.Sprintf("convert port to int error [%s],port str: %s", err.Error(), strList[i*5+1]))
			return nil, err
		}
		dbConfig := &goToolMSSql.MSSqlConfig{
			Server: strList[i*5+0],
			Port:   port,
			User:   strList[i*5+2],
			Pwd:    strList[i*5+3],
			DbName: strList[i*5+4],
		}
		rList = append(rList, dbConfig)
	}

	for _, d := range rList {
		log.Debug(fmt.Sprintf("%s %d %s %s %s", d.Server, d.Port, d.User, d.Pwd, d.DbName))
	}

	return rList, nil
}

func (c *common) GetRowsBySQL(dbConfig *goToolMSSql.MSSqlConfig, sql string, args ...interface{}) (*sql.Rows, error) {
	conn, err := goToolMSSql.GetConn(dbConfig)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	if args == nil {
		rows, err := conn.Query(sql)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		return rows, nil
	} else {
		rows, err := conn.Query(sql, args...)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		return rows, nil
	}
}

func (c *common) SetRowsBySQL(dbConfig *goToolMSSql.MSSqlConfig, sql string, args ...interface{}) error {
	conn, err := goToolMSSql.GetConn(dbConfig)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if args == nil {
		_, err = conn.Exec(sql)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return nil
	} else {
		_, err := conn.Exec(sql, args...)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return nil
	}
}
