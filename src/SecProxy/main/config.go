package main

import (
	"fmt"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//全局变量保存配置
var (
	secKillConf = &SecSkillConfig{
		secProductInfo: make(map[int]*SecProductInfoConf, 1024),
	}
)

// 秒杀系统配置
type SecSkillConfig struct {
	redisConf        RedisConf
	etcdConf         EtcdConf
	secProductInfo   map[int]*SecProductInfoConf
	logPath          string
	logLevel         string
	rwSecProductLock sync.RWMutex
}
type RedisConf struct {
	redisAddr      string
	redisMaxIdle   int
	redisMaxActive int
	redisIdleTime  int
}
type EtcdConf struct {
	etcdAddr          string
	etcdTimeout       int
	etcdSecKeyPrefix  string
	etcdSecProductKey string
}
type SecProductInfoConf struct {
	ProductId int
	StartTime int
	EndTime   int
	Status    int
	Total     int
	Leave     int
}

//初始haul配置项
func initConfig() (err error) {
	redisAddr := beego.AppConfig.String(("redis_addr"))
	etcdAddr := beego.AppConfig.String("etcd_addr")

	logs.Debug("read config succ, redis addr:%v", redisAddr)
	logs.Debug("read config succ, etcd addr:%v", etcdAddr)

	secKillConf.redisConf.redisAddr = redisAddr
	secKillConf.etcdConf.etcdAddr = etcdAddr

	if len(redisAddr) == 0 || len(etcdAddr) == 0 {
		err = fmt.Errorf("init config failed,redis[%s] or etcd[%s] config is null", redisAddr, etcdAddr)
		return
	}

	redisMaxIdle, err := beego.AppConfig.Int("redis_max_idle")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_max_idle[%s] error:%v ", redisMaxIdle, err)
		return
	}
	redisMaxActive, err := beego.AppConfig.Int("redis_max_active")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_max_active[%s] error:%v ", redisMaxActive, err)
		return
	}
	redisIdleTimeout, err := beego.AppConfig.Int("redis_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed,redis_idle_timeout[%s] error:%v ", redisIdleTimeout, err)
		return
	}
	secKillConf.redisConf.redisMaxIdle = redisMaxIdle
	secKillConf.redisConf.redisMaxActive = redisMaxActive
	secKillConf.redisConf.redisIdleTime = redisIdleTimeout
	etcdTimeout, err := beego.AppConfig.Int("etcd_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed,etcd_timeout[%s] error", etcdTimeout)
		return
	}
	etcdSecKeyPrefix := beego.AppConfig.String("etcd_sec_key_prefix")
	if len(etcdSecKeyPrefix) == 0 {
		err = fmt.Errorf("init config failed,etcd_sec_key_prefix[%s] error", etcdSecKeyPrefix)
		return
	}
	secKillConf.etcdConf.etcdSecKeyPrefix = etcdSecKeyPrefix

	etcdSecProductKey := beego.AppConfig.String("etcd_product_key")
	if len(etcdSecProductKey) == 0 {
		err = fmt.Errorf("init config failed,etcd_Product_key[%s] error", etcdSecProductKey)
		return
	}
	secKillConf.etcdConf.etcdSecProductKey = fmt.Sprintf("%s/%s", secKillConf.etcdConf.etcdSecKeyPrefix, etcdSecProductKey)

	secKillConf.logPath = beego.AppConfig.String(("log_path"))
	secKillConf.logLevel = beego.AppConfig.String("log_level")
	return
}
