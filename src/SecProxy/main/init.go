package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.etcd.io/etcd/mvcc/mvccpb"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	etcd_client "go.etcd.io/etcd/clientv3"
)

var (
	redisPool  *redis.Pool
	etcdClient *etcd_client.Client
)

func convertLogLevel(level string) int {
	switch level {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

//用于初始化工作
func initLogs() (err error) {
	config := make(map[string]interface{})
	config["filename"] = secKillConf.logPath
	config["level"] = convertLogLevel(secKillConf.logLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marshal failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}
func initRedis() (err error) {
	redisPool = &redis.Pool{
		MaxIdle:     secKillConf.redisConf.redisMaxIdle,
		MaxActive:   secKillConf.redisConf.redisMaxActive,
		IdleTimeout: time.Duration(secKillConf.redisConf.redisIdleTime) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.redisConf.redisAddr)
		},
	}
	conn := redisPool.Get()
	defer conn.Close()
	_, err = conn.Do("ping")
	if err != nil {
		err = fmt.Errorf("connecte to redis failed err:%v", err)
		return
	}
	return
}
func initEtcd(etcdAddr string) (err error) {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: time.Duration(secKillConf.etcdConf.etcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed,err:", err)
		return
	}
	etcdClient = cli
	return
}
func loadSecConf() (err error) {
	key := secKillConf.etcdConf.etcdSecProductKey
	resp, err := etcdClient.Get(context.Background(), key)
	if err != nil {
		logs.Error("get [%s] from etcd failed,err:%v", key, err)
		return
	}
	var secProductInfo []SecProductInfoConf
	for k, v := range resp.Kvs {
		logs.Debug("key[%s] values[%v]", k, v)
		err = json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			logs.Error("Unmarshal sec product info failed,err :%v", err)
			return
		}
		logs.Debug("sec info conf is[%v]", secProductInfo)
	}
	secKillConf.rwSecProductLock.Lock()
	for _, v := range secProductInfo {
		secKillConf.secProductInfo[v.ProductId] = &v
		return
	}
	secKillConf.rwSecProductLock.Unlock()
	return
}
func initSec() (err error) {
	err = initLogs()
	if err != nil {
		logs.Error("init logs failes,err:", err)
	}
	err = initRedis()
	if err != nil {
		logs.Error("init redis filed,err:", err)
		return
	}
	err = initEtcd(secKillConf.etcdConf.etcdAddr)
	if err != nil {
		logs.Error("init redis filed,err:", err)
		return
	}
	err = loadSecConf()
	if err != nil {
		logs.Error("init sec conf failed,err", err)
		return
	}
	// 启动watch进程 用来观察配置是否发生变化
	initSecProductWatcher()

	logs.Info("init sec success")
	return
}
func initSecProductWatcher() {
	go watchSecProductKey(secKillConf.etcdConf.etcdSecProductKey)
}
func watchSecProductKey(key string) {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: time.Duration(secKillConf.etcdConf.etcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed,err", err)
		return
	}
	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo []SecProductInfoConf
		var getConfSucc = true
		for wresp := range rch {
			for _, ev := range wresp.Events {
				//key 删除事件
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s]'s config deleted", key)
					continue
				}
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key[%s],Unmarshal[%s],err:%v", err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd succ,%v", secProductInfo)
			}
			if getConfSucc {
				logs.Debug("get config from etcd succ,%v", secProductInfo)
				updateSecProductInfo(secProductInfo)
			}
		}
	}
}
func updateSecProductInfo(secProductInfo []SecProductInfoConf) {
	// secKillConf.rwSecProductLock.Lock()
	// for _, v := range secProductInfo {
	// 	secKillConf.secProductInfo[v.ProductId] = &v
	// 	return
	// }
	// secKillConf.rwSecProductLock.UnLock()
	//写时复制优化 copyonwrite
	var tmp map[int]*SecProductInfoConf = make(map[int]*SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		tmp[v.ProductId] = &v
	}
	secKillConf.rwSecProductLock.Lock()
	secKillConf.secProductInfo = tmp
	secKillConf.rwSecProductLock.Unlock()
}
