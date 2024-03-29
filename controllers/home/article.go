package home

import (
    "encoding/json"
    "fmt"
    "go-blog/models/admin"
    "go-blog/utils"
    "html/template"
    "strconv"
    "strings"
    "time"

    "github.com/astaxie/beego"
    "github.com/astaxie/beego/orm"
    "github.com/astaxie/beego/validation"
)

type ArticleController struct {
    BaseController
}

// 类表
func (c *ArticleController) List() {

    o := orm.NewOrm()
    var setting admin.Setting
    o.QueryTable(new(admin.Setting)).Filter("name", "limit").One(&setting)
    l := setting.Value
    limit, err := strconv.ParseInt(l, 10, 64)
    if err != nil || limit == 0 {
        li, _ := beego.AppConfig.Int64("limit")
        limit = li
    }

    //limit, _ := beego.AppConfig.Int64("limit") // 一页的数量
    page, _ := c.GetInt64("page", 1) // 页数
    offset := (page - 1) * limit     // 偏移量
    categoryId, _ := c.GetInt("c", 0)
    if categoryId == 0 {
        categoryId = 2
        //c.Abort("404")
    }
    //o := orm.NewOrm()
    article := new(admin.Article)

    var articles []*admin.Article
    qs := o.QueryTable(article)
    qs = qs.Filter("status", 1)
    qs = qs.Filter("User__Name__isnull", false)
    qs = qs.Filter("Category__Name__isnull", false)

    search := c.GetString("s")
    if search != "" {
        qs = qs.Filter("title__icontains", search)
        c.Data["s"] = search
    }
    author := c.GetString("a")
    if author != "" {
        qs = qs.Filter("singer__icontains", author)
        c.Data["a"] = search
    }

    if categoryId != 0 {

        category := new(admin.Category)
        var categorys []*admin.Category
        cqs := o.QueryTable(category)
        cqs = cqs.Filter("status", 1)
        cqs.OrderBy("-sort").All(&categorys)

        ids := utils.CategoryTreeR(categorys, categoryId, 0)

        var cids []int
        cids = append(cids, categoryId)
        for _, v := range ids {
            cids = append(cids, v.Id)
        }

        /*c.Data["json"] = cids
          c.ServeJSON()
          c.StopRun()*/

        qs = qs.Filter("Category__ID__in", cids)

    }
    c.Data["CategoryID"] = &categoryId
    // 查出当前分类下的所有子分类id

    date := c.GetString("date")
    if date != "" {
        if len(date) == 7 {
            start := ""
            end := ""
            dateNumStr := beego.Substr(date, len("2018-"), 2)
            yearNumStr := beego.Substr(date, len("20"), 2)
            dateNum, _ := strconv.Atoi(dateNumStr)
            yearNum, _ := strconv.Atoi(yearNumStr)

            start = utils.SubString(date, len("2018-01")) + "-01 00:00:00"
            if dateNum >= 12 {
                endYearStr := strconv.Itoa(yearNum + 1)
                end = utils.SubString(date, len("20")) + endYearStr + "-01-01 00:00:00"
            }

            if dateNum < 9 {
                endStr := strconv.Itoa(dateNum + 1)
                end = utils.SubString(date, len("2018-0")) + endStr + "-01 00:00:00"
            }
            if dateNum >= 9 && dateNum < 12 {
                endStr := strconv.Itoa(dateNum + 1)
                end = utils.SubString(date, len("2018-")) + endStr + "-01 00:00:00"
            }

            /*c.Data["json"] = []string{start,end}
              c.ServeJSON()
              c.StopRun()*/

            qs = qs.Filter("created__gte", start)
            qs = qs.Filter("created__lte", end)
            c.Data["Date"] = utils.SubString(start, len("2018-01"))

        } else {
            date = utils.SubString(date, len("2018-01-01"))
            tm, _ := time.Parse("2006-01-02", date)
            unix := tm.Unix() //1566432000

            startFormat := time.Unix(unix, 0).Format("2006-01-02 15:04:05")
            moreUnix, _ := utils.ToInt64(int64(60 * 60 * 24))
            endFormat := time.Unix(unix+moreUnix, 0).Format("2006-01-02 15:04:05")
            start := utils.SubString(startFormat, len("2018-01-01")) + " 00:00:00"
            end := utils.SubString(endFormat, len("2018-01-01")) + " 00:00:00"

            // 刷选
            qs = qs.Filter("created__gte", start)
            qs = qs.Filter("created__lte", end)
            c.Data["Date"] = utils.SubString(start, len("2018-01-01"))
        }
    }

    tag := c.GetString("tag")
    if tag != "" {
        qs = qs.Filter("tag__icontains", tag)
        c.Data["keywordsTag"] = tag
    }

    // 统计
    count, err := qs.Count()
    if err != nil {
        c.Abort("404")
    }

    // 获取数据
    _, err = qs.OrderBy("-recommend", "-id", "-pv").RelatedSel().Limit(limit).Offset(offset).All(&articles)
    if err != nil {
        panic(err)
    }
    c.Data["Data"] = &articles
    c.Data["Paginator"] = utils.GenPaginator(page, limit, count)

    // Menu
    c.Log("article")

    if categoryId == 0 {
        c.Data["index"] = "博客列表"
    } else {
        categoryKey := admin.Category{Id: categoryId}
        err = o.Read(&categoryKey)
        if err == nil {
            c.Data["index"] = categoryKey.Name
            if author != "" {
                c.Data["keyword"] = fmt.Sprintf("%s钢琴谱", author)
                c.Data["index"] = fmt.Sprintf("%s 免费钢琴乐谱合集", author)
                c.Data["description"] = fmt.Sprintf("一起学%s钢琴谱，%s热门钢琴谱。从流行歌曲到最受欢迎的古典音乐，让你学会用钢琴弹奏最美妙的歌曲。", author, author)
            }
        } else {
            c.Data["index"] = "博客列表"
        }
    }

    //fmt.Println(&articles)
    switch c.Template {
    case "nihongdengxia":
        c.TplName = "home/" + c.Template + "/read.html"
    default:
        c.TplName = "home/" + c.Template + "/list.html"
    }
}

// 详情
func (c *ArticleController) Detail() {

    id := c.Ctx.Input.Param(":id")
    viewType := c.GetString("type")
    // 基础数据
    o := orm.NewOrm()
    article := new(admin.Article)
    var articles []*admin.Article
    qs := o.QueryTable(article)
    err := qs.Filter("id", id).RelatedSel().One(&articles)
    if err != nil {
        c.Abort("404")
    }

    /*c.Data["json"]= &articles
      c.ServeJSON()
      c.StopRun()*/
    keywordsArr := strings.Split(articles[0].Keywords, ",")
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, "歌词") {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"歌词")
    }
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, "五线谱") {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"五线谱")
    }
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, "钢琴谱") {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"钢琴谱")
    }
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, articles[0].Singer) {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"-"+articles[0].Singer)
    }
    articles[0].Keywords = strings.Join(keywordsArr, ",")

    c.Data["Data"] = &articles[0]

    if beego.AppConfig.String("view") == "default" {
        var listData = make(map[string][]*admin.Article)
        var list []*admin.Article
        _, err = o.QueryTable(article).Filter("status", 1).Filter("User__Name__isnull", false).Filter("Category__Name__isnull", false).OrderBy("id").RelatedSel().All(&list, "id", "title")

        for _, v := range list {
            listData[v.Category.Name] = append(listData[v.Category.Name], v)
        }
        c.Data["List"] = &listData
        articleId, _ := strconv.Atoi(id)
        c.Data["ArticleId"] = articleId
        /*c.Data["json"]= &listData
          c.ServeJSON()
          c.StopRun()*/

    }

    var other admin.Other

    if &articles[0].Other != nil {
        json.Unmarshal([]byte(articles[0].Other), &other)
    }

    other.SubjectInfo = strings.Replace(other.SubjectInfo, "\n", "<br>", -1)
    c.Log("detail")
    titleIndex := strings.ReplaceAll(articles[0].Title, "（五线谱、双手简谱、数字谱、Midi、PDF）免费下载", "")
    c.Data["index"] = &titleIndex
    c.Data["Other"] = other

    if viewType == "single" {
        c.TplName = "home/" + c.Template + "/doc.html"
    } else if viewType == "ms" || articles[0].Category.Id == 2 {
        c.TplName = "home/" + c.Template + "/ms.html"
    } else {
        c.TplName = "home/" + c.Template + "/detail.html"
    }
    //c.TplName = "home/nihongdengxia/review.html"
}

// 播放
func (c *ArticleController) Playback() {
    id := c.Ctx.Input.Param(":id")
    // 基础数据
    o := orm.NewOrm()
    article := new(admin.Article)
    var articles []*admin.Article
    qs := o.QueryTable(article)
    err := qs.Filter("id", id).RelatedSel().One(&articles)
    if err != nil {
        c.Abort("404")
    }

    articles[0].Uuid = fmt.Sprintf("/static/pianomusic/%s/%s.mxl", c.getPath(articles[0].Id), articles[0].Uuid)

    keywordsArr := strings.Split(articles[0].Keywords, ",")
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, "歌词") {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"歌词下载")
    }
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, "五线谱") {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"五线谱下载")
    }
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, "钢琴谱") {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"钢琴谱下载")
    }
    if articles[0].Category.Id == 2 && !strings.Contains(articles[0].Keywords, articles[0].Singer) {
        keywordsArr = append(keywordsArr, keywordsArr[0]+"-"+articles[0].Singer)
    }
    keywordsArr = append(keywordsArr, keywordsArr[0]+"-在线练琴")
    articles[0].Keywords = strings.Join(keywordsArr, ",")

    articles[0].Remark = "在线练习并下载" + keywordsArr[0] + "钢琴谱。" + articles[0].Remark
    c.Data["Data"] = &articles[0]

    c.Log("playback")
    titleIndex := articles[0].Title + "在线练琴"
    c.Data["index"] = &titleIndex

    c.TplName = "home/" + c.Template + "/playback.html"
}

func (c *ArticleController) getPath(id int) string {
    return fmt.Sprintf("%03d/%07d", (id/10000)+1, id)
}

// 统计访问量
func (c *ArticleController) Pv() {

    ids := c.Ctx.Input.Param(":id")
    id, _ := strconv.Atoi(ids)
    /*c.Data["json"] = c.Input()
      c.ServeJSON()
      c.StopRun()*/

    response := make(map[string]interface{})

    o := orm.NewOrm()

    article := admin.Article{Id: id}
    if o.Read(&article) == nil {
        article.Pv = article.Pv + 1

        valid := validation.Validation{}
        valid.Required(article.Id, "Id")
        if valid.HasErrors() {
            // 如果有错误信息，证明验证没通过
            // 打印错误信息
            for _, err := range valid.Errors {
                //log.Println(err.Key, err.Message)
                response["msg"] = "Error."
                response["code"] = 500
                response["err"] = err.Key + " " + err.Message
                c.Data["json"] = response
                c.ServeJSON()
                c.StopRun()
            }
        }

        if _, err := o.Update(&article); err == nil {
            response["msg"] = "Success."
            response["code"] = 200
            response["id"] = id
        } else {
            response["msg"] = "Error."
            response["code"] = 500
            response["err"] = err.Error()
        }
    } else {
        response["msg"] = "Error."
        response["code"] = 500
        response["err"] = "ID 不能为空！"
    }

    c.Data["json"] = response
    c.ServeJSON()
    c.StopRun()
}

// 评论
func (c *ArticleController) Review() {

    aid, _ := c.GetInt("aid")
    name := c.GetString("name")
    review := c.GetString("review")
    site := c.GetString("site", "")

    o := orm.NewOrm()
    reviewsMd := admin.Review{
        Name:      template.HTMLEscapeString(name),
        Review:    template.HTMLEscapeString(review),
        Site:      template.HTMLEscapeString(site),
        ArticleId: aid,
        Status:    1,
        Customer:  &admin.Customer{Id: 1},
    }

    response := make(map[string]interface{})

    valid := validation.Validation{}
    valid.Required(reviewsMd.Name, "Name")
    valid.Required(reviewsMd.Review, "Review")
    valid.Required(reviewsMd.ArticleId, "ArticleId")

    if valid.HasErrors() {
        // 如果有错误信息，证明验证没通过
        // 打印错误信息
        for _, err := range valid.Errors {
            //log.Println(err.Key, err.Message)
            response["msg"] = "新增失败！"
            response["code"] = 500
            response["err"] = err.Key + " " + err.Message
            c.Data["json"] = response
            c.ServeJSON()
            c.StopRun()
        }
    }

    // 更新评论数量
    article := admin.Article{Id: aid}
    o.Read(&article)
    article.Review = article.Review + 1
    o.Update(&article)

    if id, err := o.Insert(&reviewsMd); err == nil {
        response["msg"] = "新增成功！"
        response["code"] = 200
        response["id"] = id
    } else {
        response["msg"] = "新增失败！"
        response["code"] = 500
        response["err"] = err.Error()
    }

    c.Data["json"] = response
    c.ServeJSON()
    c.StopRun()
}

func (c *ArticleController) ReviewList() {

    id := c.Ctx.Input.Param(":id")
    //limit, _ := beego.AppConfig.Int64("limit") // 一页的数量
    limit := int64(20)
    page, _ := c.GetInt64("page", 1) // 页数
    offset := (page - 1) * limit     // 偏移量
    response := make(map[string]interface{})

    o := orm.NewOrm()
    review := new(admin.Review)

    var reviews []*admin.Review
    qs := o.QueryTable(review)
    qs = qs.Filter("status", 1)
    qs = qs.Filter("article_id", id)

    // 获取数据
    _, err := qs.OrderBy("-id").RelatedSel().Limit(limit).Offset(offset).All(&reviews)

    if err != nil {
        response["msg"] = "Error."
        response["code"] = 500
        c.Data["json"] = response
        c.ServeJSON()
        c.StopRun()
    }

    // 统计
    count, err := qs.Count()
    if err != nil {
        response["msg"] = "Error."
        response["code"] = 500
        c.Data["json"] = response
        c.ServeJSON()
        c.StopRun()
    }

    response["Data"] = &reviews
    response["Paginator"] = utils.GenPaginator(page, limit, count)

    response["msg"] = "Success."
    response["code"] = 200
    c.Data["json"] = response
    c.ServeJSON()
    c.StopRun()

}

func (c *ArticleController) Like() {

    response := make(map[string]interface{})
    ip := c.Ctx.Input.IP()
    id, _ := c.GetInt("id")

    o := orm.NewOrm()
    qs := o.QueryTable(new(admin.Log))

    qs = qs.Filter("ip", ip)
    qs = qs.Filter("create__gte", beego.Date(time.Now(), "Y-m-d 00:00:00"))
    qs = qs.Filter("create__lte", beego.Date(time.Now(), "Y-m-d H:i:s"))
    qs = qs.Filter("page", "like"+strconv.Itoa(id))

    count, e := qs.Count()
    if e != nil {
        response["msg"] = "Error."
        response["code"] = 500
        response["err"] = e.Error()
        c.Data["json"] = response
        c.ServeJSON()
        c.StopRun()
    }
    if count >= 1 {
        response["msg"] = "Error."
        response["code"] = 500
        response["err"] = "亲，点赞过了，明天再来哦！"
        c.Data["json"] = response
        c.ServeJSON()
        c.StopRun()
    }

    c.Log("like" + strconv.Itoa(id))

    article := admin.Article{Id: id}
    if o.Read(&article) == nil {
        article.Like = article.Like + 1

        valid := validation.Validation{}
        valid.Required(article.Id, "Id")
        if valid.HasErrors() {
            // 如果有错误信息，证明验证没通过
            // 打印错误信息
            for _, err := range valid.Errors {
                //log.Println(err.Key, err.Message)
                response["msg"] = "Error."
                response["code"] = 500
                response["err"] = err.Key + " " + err.Message
                c.Data["json"] = response
                c.ServeJSON()
                c.StopRun()
            }
        }

        if _, err := o.Update(&article); err == nil {
            response["msg"] = "Success."
            response["code"] = 200
            response["like"] = article.Like
        } else {
            response["msg"] = "Error."
            response["code"] = 500
            response["err"] = err.Error()
        }
    } else {
        response["msg"] = "Error."
        response["code"] = 500
        response["err"] = "ID 不能为空！"
    }

    c.Data["json"] = response
    c.ServeJSON()
    c.StopRun()
}
