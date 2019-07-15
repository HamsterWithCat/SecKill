package main

import (
	_ "SecProxy/router"

	"github.com/astaxie/beego"
)

func main() {
	//加载配置项
	err := initConfig()
	if err != nil {
		panic(err)
		return
	}
	//初始化redis etcd组件
	err = initSec()
	if err != nil {
		panic(err)
		return
	}
	beego.Run()
}
