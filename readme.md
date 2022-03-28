#### Go Blog 

一个基于Beego开发的社交商城博客系统

[![Go Blog Stargazers over time](https://starchart.cc/1920853199/go-blog.svg)](https://starchart.cc/1920853199/go-blog)

[![GitHub stars](https://img.shields.io/github/stars/1920853199/go-blog)](https://github.com/1920853199/go-blog/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/1920853199/go-blog)](https://github.com/1920853199/go-blog/network)
[![GitHub license](https://img.shields.io/github/license/1920853199/go-blog)](https://github.com/1920853199/go-blog/master/LICENSE)

> 前台演示站点(社区) https://nihongdengxia.com/ 
> 陈立个人博客 http://clblog.club/ 
> Goooooooogle 搜索摸鱼 http://gooooooooogle.cn/ 
> 
![image](https://user-images.githubusercontent.com/28426364/128692732-d92d8de8-8174-4447-9be9-57ba14a8de2a.png)
![image](https://user-images.githubusercontent.com/28426364/129147083-29c15731-f90e-4146-8bcb-aaa42c92bab3.png)

> 前台演示站点(社区) https://nihongdengxia.com/ 


### 系统功能
![系统功能](/系统功能.png "系统功能")

### Install 

#### 方式1 docker 安装（推荐）
1. 先安装`docker`以及`docker-compose`
2. 把根目录下的`docker-compose.yml`赋值到你需要运行的`Go Blog`项目的目录下，执行`docker-compose up -d`.（会报找不到数据库的错误，忽略，在步骤3导入数据后就正常了）
3. 登录`docker`启动的`mysql`，新建数据库`go-blog`,导入`go-blog/database/blog-mysql.sql`数据。
4. 访问url`http://127.0.0.0:8080`,后台url`http://127.0.0.0:8080/admin`,默认账户:`user`,密码:`123456`

#### 方式二 源码安装
1. 把Go Blog项目拉到本地

```
https://github.com/1920853199/go-blog.git
```

2. 新建数据库，导入数据库文件，数据库文件/database/blog.sql

3. 修改项目配置信息

```
#conf/app.conf

appname = go-blog
httpport = 8088
runmode = dev
EnableAdmin = false
sessionon = true
url = 127.0.0.1:8088

[db]
dbType = mysql
dbUser = root
dbPass = root
dbHost = 127.0.0.1
dbPort = 3306
dbName = blog

[redis]
rHost = 127.0.0.1
rPort = 6379


```

4. 在bo-blog 根目录下执行bee run ，访问 http://127.0.0.1:8888 即可

5. 守护进程模式运行 可以了解PM2的相关信息，配置可查看start.sh 文件

6. nginx代理示例

```
server {
        listen 80;
        server_name go-blog.cn;
        root    /home/data/go-blog;

        location ~ \.(txt|xml)$ {
                root /home/data/go-blog;
        }

        location / {
            proxy_pass http://127.0.0.1:8889;
            #proxy_redirect off;
            proxy_http_version    1.1;
            proxy_cache_bypass    $http_upgrade;
            proxy_set_header Upgrade            $http_upgrade;
            proxy_set_header Connection         "upgrade";
            proxy_set_header Host               $host;
            proxy_set_header X-Real-IP          $remote_addr;
            proxy_set_header X-Forwarded-For    $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto  $scheme;
            proxy_set_header X-Forwarded-Host   $host;
            proxy_set_header X-Forwarded-Port   $server_port;
        }

        access_log    /home/wwwlogs/go-blog.access.log;
}

```

### 互动交流
#### 与作者对话
该项目是利用业余时间进行开发的，开发思路主要是源于自己的项目积累及个人思考，如果您有更好的想法和建议请与我进行沟通，一起探讨，畅聊技术人生，相互学习，一起进步。我非常期待！下面是我的微信二维码（如果此项目对您提供了帮助也可以请作者喝杯咖啡 (*￣︶￣)，聊表心意，一起星巴克「续杯」~嘿嘿 ）：

<div>

<img style="display: block;float: left;padding-right: 20px;" src="https://cdn.learnku.com/uploads/images/202012/23/43046/Fzyua3mXnY.jpeg!large" width="256" alt="we-pay" />
</div>

## END
感谢您关注此项目 : )，如果有好的想法欢迎 Issue or PR。
