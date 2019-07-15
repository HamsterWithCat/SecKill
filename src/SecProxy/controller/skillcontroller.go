package controller

import (
	"github.com/astaxie/beego"
)

type SkillController struct {
	beego.Controller
}

func (sc *SkillController) SecKill() {
	sc.Data["json"] = "sec kill"
	sc.ServeJSON()
}
func (sc *SkillController) SecInfo() {
	sc.Data["json"] = "sec info"
	sc.ServeJSON()
}
