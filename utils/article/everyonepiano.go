package article

import (
    "context"
    "fmt"
    "github.com/PuerkitoBio/goquery"
    "github.com/astaxie/beego/orm"
    "github.com/beevik/etree"
    "github.com/gocolly/colly"
    "github.com/gocolly/colly/extensions"
    "github.com/google/uuid"
    "go-blog/models/admin"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"
)

func GetMidi() {
    baseUrl := "https://www.everyonepiano.cn"
    url := "https://www.everyonepiano.cn/Music.html"
    c := colly.NewCollector()
    //c.WithTransport(&http.Transport{
    //    Proxy: http.ProxyFromEnvironment,
    //    DialContext: (&net.Dialer{
    //        Timeout:   30 * time.Second,
    //        KeepAlive: 30 * time.Second,
    //    }).DialContext,
    //    MaxIdleConns:          100,
    //    IdleConnTimeout:       90 * time.Second,
    //    TLSHandshakeTimeout:   10 * time.Second,
    //    ExpectContinueTimeout: 1 * time.Second,
    //})
    c.Limit(&colly.LimitRule{
        DomainGlob:  "*everyonepiano.*",
        Parallelism: 1,
        //Delay: 1 * time.Second,
        //RandomDelay: 5 * time.Second,
    })
    extensions.RandomUserAgent(c)
    extensions.Referer(c)
    err := c.Post("https://www.everyonepiano.cn/Login/index",
        map[string]string{
            "username": "gydcd",
            "password": "starx.19800724",
            "submit":   "1",
            "care_url": "https://www.everyonepiano.cn/Music-search/?word=%E5%85%89%E5%B9%B4&come=web",
            "backsure": "",
            "backurl":  "",
            "go":       "",
            "__hash__": "f50fd4d722f5cdd28249de4287082733_b4742b81b5409fc617c45ad0caea2fbb",
        },
    )
    if err != nil {
        log.Fatal(err)
    }
    //visit := true
    c.OnHTML(".MusicIndexBox", func(e *colly.HTMLElement) {
        //if !visit {
        //    return
        //}
        no := e.DOM.Find(".MIMusicNO").Text() // 0014129
        // 是否存在
        o := orm.NewOrm()
        article := admin.Article{}
        qs := o.QueryTable(article)
        err = qs.Filter("no", no).One(&article)
        if err == nil || err != orm.ErrNoRows {
            return
        }

        if detail, ok := e.DOM.Find(".Title").Attr("href"); ok {
            if strings.HasPrefix(detail, "/") {
                detail = baseUrl + detail
            }
            c.Visit(detail)
            //visit = false
        }
    })

    c.OnHTML("html", func(e *colly.HTMLElement) {
        if strings.HasPrefix(e.Request.URL.Path, "/Music-") {
            fmt.Println(time.Now().String(), ", detail success:", e.Request.URL.Path)
            title := strings.Replace(e.DOM.Find("title").Text(), "【谱】", "", 1)
            tags := strings.Split(title, "-")
            no, _ := strconv.ParseUint(strings.Replace(strings.Split(e.Request.URL.Path, "-")[1], ".html", "", 1), 10, 0)
            o := orm.NewOrm()
            article := admin.Article{
                Title:    title,
                Tag:      tags[0] + ",乐谱",
                Desc:     tags[0] + "乐谱",
                Url:      "/",
                Status:   1,
                User:     &admin.User{Id: 1, Created: time.Now()},
                Category: &admin.Category{Id: 2},
                No:       fmt.Sprintf("%07d", no),
                Singer:   e.DOM.Find(".EOPReadInfoTxt").Find("a").Text(),
            }
            e.DOM.Find("#MusicInfoTxt2>p").Each(func(i int, s *goquery.Selection) {
                h, _ := s.Html()
                if !strings.Contains(h, ".html") && !strings.Contains(h, "和五线谱完全") {
                    h = strings.ReplaceAll(h, "EOP", "零度钢琴")
                    h = strings.Replace(h, "歌词下方", "上方", 1)
                    h = strings.ReplaceAll(h, "小编", "我")
                    article.Html = article.Html + "<p>" + h + "</p>\n"
                }
            })
            e.DOM.Find("meta").Each(func(i int, s *goquery.Selection) {
                if name, _ := s.Attr("name"); name == "description" {
                    description, _ := s.Attr("content")
                    article.Remark = description
                }
                if name, _ := s.Attr("name"); name == "keywords" {
                    keywords, _ := s.Attr("content")
                    keywords = strings.ReplaceAll(keywords, "EOP", "零度钢琴")
                    keywords = strings.ReplaceAll(keywords, "魔鬼训练", "训练")
                    article.Keywords = keywords
                }
            })
            coverUrl, _ := e.DOM.Find(".EOPReadPic").Attr("src")
            if strings.HasPrefix(coverUrl, "/") {
                coverUrl = baseUrl + coverUrl
            }
            midiUrl := baseUrl + strings.Replace(e.Request.URL.Path, "/Music-", "/Midi-", 1)
            id, _err := o.Insert(&article)
            if  _err == nil {
                cover := fmt.Sprintf("/static/pianomusic/%s/%07d-small.jpg", getPath(id), id)
                dir := fmt.Sprintf("./static/pianomusic/%s", getPath(id))
                DownloadFile(dir, "."+cover, coverUrl)
                //updateArt := admin.Article{Cover: cover}
                //o.Update(&updateArt, "Cover")
                _, _ = o.QueryTable(article).Filter("id", id).Update(orm.Params{
                    "cover": cover,
                })
                c.Visit(midiUrl)
            } else {
                fmt.Println(_err)
                panic(_err)
            }
        }
        if strings.HasPrefix(e.Request.URL.Path, "/Midi-") {
            midiUrl, ok := e.DOM.Find(".DownDiv>.btn-success").Attr("href")
            if ok && strings.HasPrefix(midiUrl, "/") {
                midiUrl = baseUrl + midiUrl
            }
            if midiUrl != "#" {
                //fmt.Println(midiUrl)
                c.Visit(midiUrl)
            }
        }
    })

    c.OnResponse(func(r *colly.Response) {
        //if strings.Contains(r.Headers.Get("Content-Type"), "application/force-download") {
        //fmt.Println(r.Request.URL.Path)
        if strings.HasPrefix(r.Request.URL.Path, "/Music/EopFile") {
            no, _ := strconv.ParseUint(strings.Split(r.Request.URL.Path, "/")[3], 10, 0)
            noStr := fmt.Sprintf("%07d", no)
            o := orm.NewOrm()
            article := admin.Article{}
            qs := o.QueryTable(article)
            err = qs.Filter("no", noStr).One(&article)
            if err == nil {
                artUuid := uuid.New().String()
                //updateArt := admin.Article{Uuid: artUuid}
                //o.Update(&updateArt, "Uuid")
                _, _ = o.QueryTable(article).Filter("id", article.Id).Update(orm.Params{
                    "uuid": artUuid,
                })

                path := fmt.Sprintf("./static/pianomusic/%s/%s.midi", getPath(int64(article.Id)), artUuid)
                dir := fmt.Sprintf("./static/pianomusic/%s", getPath(int64(article.Id)))
                err = r.Save(path)
                if err != nil {
                    fmt.Println("dowload video error", r.Request.URL.Path)
                    fmt.Println(err)
                    return
                }

                // export QT_QPA_PLATFORM=offscreen;mscore /tmp/0007445.midi -o /tmp/0007445.musicxml
                // export QT_QPA_PLATFORM=offscreen;mscore /tmp/0007445.musicxml -o /tmp/0007445.mxl
                musicXMLPath := fmt.Sprintf("%s/%s.musicxml", dir, artUuid)
                mxlPath := fmt.Sprintf("%s/%s.mxl", dir, artUuid)
                cmd1 := fmt.Sprintf("export QT_QPA_PLATFORM=offscreen;mscore %s -o %s", path, musicXMLPath)
                cmd := exec.Command("bash","-c", cmd1)
                err = cmd.Run()
                if err != nil {
                    fmt.Println(err)
                    return
                }

                fixMusicXML(musicXMLPath, article.Title, article.Singer)

                cmd2 := fmt.Sprintf("export QT_QPA_PLATFORM=offscreen;mscore %s -o %s", musicXMLPath, mxlPath)
                ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
                defer cancel()
                cmd = exec.CommandContext(ctx, "bash","-c", cmd2)
                err = cmd.Run()
                if err != nil || ctx.Err() == context.DeadlineExceeded {
                    fmt.Println("context DeadlineExceeded, err:", err, artUuid)
                    cmd2 = fmt.Sprintf("export QT_QPA_PLATFORM=offscreen;mscore %s -o %s", path, mxlPath)
                    cmd = exec.Command( "bash","-c", cmd2)
                    err = cmd.Run()
                    return
                }
            } else {
                fmt.Println(err, noStr)
            }
        }
    })

    // https://www.everyonepiano.cn/Music.html?&p=1413&canshu=cn_edittime&word=&author=&jianpu=&paixu=desc&username=
    for i := 2; i >= 2; i-- {
       visitUrl := fmt.Sprintf("https://www.everyonepiano.cn/Music.html?&p=%d&canshu=cn_edittime&word=&author=&jianpu=&paixu=desc&username=", i)
       c.Visit(visitUrl)
    }
    c.Visit(url)
}

func getPath(id int64) string {
    return fmt.Sprintf("%03d/%07d", (id/10000)+1, id)
}

func DownloadFile(dir, filepath string, url string) error {
    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if _, err = os.Stat(dir); os.IsNotExist(err) {
        // path/to/whatever does not exist
        err = os.MkdirAll(dir, os.ModePerm)
        if err != nil {
            return err
        }
    }

    // Create the file
    out, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    return err
}

func fixMusicXML(path, title, singer string) {
    doc := etree.NewDocument()
    if err := doc.ReadFromFile(path); err != nil {
        panic(err)
    }
    root := doc.SelectElement("score-partwise")
    workTitle := root.FindElement("./work/work-title")
    if workTitle == nil {
        work := etree.NewElement("work")
        workTitle = work.CreateElement("work-title")
        workTitle.CreateText(title)
        root.InsertChildAt(0, work)
    }
    identification := root.FindElement("./identification")
    if identification == nil {
        identification = etree.NewElement("identification")
    }
    creator := root.FindElement("./identification/creator")
    if creator == nil {
        creator = etree.NewElement("creator")
        creator.CreateText(singer)
        creator.CreateAttr("type", "composer")
        identification.InsertChildAt(0, creator)
    }

    _ = doc.WriteToFile(path)
}
