package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	errorControl "go-blog/controllers/error"
	_ "go-blog/routers"
	db "go-blog/service/databsae"
	"go-blog/utils"
	"go-blog/utils/article"
)

func init() {
	conf, err := config.NewConfig("ini", "conf/app.conf")

	if err != nil {
		logrus.Fatalf(err.Error())
	}

	database, _ := db.NewDataBase(conf.String("db::dbType"))
	orm.RegisterDriver(database.GetDriverName(), database.GetDriver())
	orm.RegisterDataBase(database.GetAliasName(), database.GetDriverName(), database.GetStr())

	beego.AddFuncMap("IndexForOne", utils.IndexForOne)
	beego.AddFuncMap("IndexAddOne", utils.IndexAddOne)
	beego.AddFuncMap("IndexDecrOne", utils.IndexDecrOne)
	beego.AddFuncMap("StringReplace", utils.StringReplace)
	beego.AddFuncMap("StringSplit", utils.StringSplit)
	beego.AddFuncMap("TimeStampToTime", utils.TimeStampToTime)

	// 每天0点定时更新站点地图
	go func() {
		c := cron.New()
		//*/1 0 * * *
		// 0 0 * * *
		c.AddFunc("0 2 * * *", func() {
			//sitemap.Sitemap("./", conf.String("url"))
			getMidi, _ := conf.Bool("midi")
			if getMidi {
				article.GetMidi()
			}
		})
		c.Start()
	}()
}

func main() {
	beego.ErrorController(&errorControl.ErrorsController{})
	//bee generate appcode -tables="cron" -driver=mysql -conn="root:root@tcp(127.0.0.1:3306)/blog" -level=3
	beego.Run()
}
