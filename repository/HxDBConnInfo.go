package repository

import (
	"fmt"
	log "github.com/Deansquirrel/goToolLog"
)

const (
	sqlGetHxDBConnInfo = "" +
		"SELECT FConnInfo " +
		"FROM HxDBConnInfo " +
		"WHERE FUserKey = ? " +
		"	And FIsStop = 0"
)

type HxDBConnInfo struct {
}

//获取核心库连接信息
func (p *HxDBConnInfo) GetConnInfo(key string) (string, error) {
	c := common{}
	config := c.GetConfigDBConfig()
	rows, err := c.GetRowsBySQL(config, sqlGetHxDBConnInfo, key)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = rows.Close()
	}()
	var connInfo string
	for rows.Next() {
		err := rows.Scan(&connInfo)
		if err != nil {
			log.Error(fmt.Sprintf("read data error: %s", rows.Err().Error()))
			return "", err
		}
	}
	if rows.Err() != nil {
		log.Error(fmt.Sprintf("read data error: %s", rows.Err().Error()))
		return "", rows.Err()
	}
	return connInfo, nil
}
