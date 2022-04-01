package errorControl

import "github.com/astaxie/beego"
/**
  该控制器处理页面错误请求
*/
type ErrorsController struct {
    beego.Controller
}
func (c *ErrorsController) Error401() {
    c.Data["content"] = "未经授权，请求要求验证身份"
    c.TplName="error/error.html"
}
func (c *ErrorsController) Error403() {
    c.Data["content"] = "服务器拒绝请求"
    c.TplName="error/error.html"
}
func (c *ErrorsController) Error404() {
    c.Data["content"] = "很抱歉您访问的地址或者方法不存在"
    c.TplName="error/error.html"
}

func (c *ErrorsController) Error500() {
    c.Data["content"] = "server error"
    c.TplName = "error/error.html"
}
func (c *ErrorsController) Error503() {
    c.Data["content"] = "服务器目前无法使用（由于超载或停机维护）"
    c.TplName = "error/error.html"
}
