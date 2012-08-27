package main

import (
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"errors"
)

var blogs = make(map[string]*Blog)

func main() {
	blog := &Blog{Url: "blog.zhaojie.me"}
	Spider(blog)
	//插入数据库
	fmt.Println("开始插入数据库")
	InsertMysql()
	time.Sleep(3000*time.Second)
}
func Spider(blog *Blog) {
	if len(blogs) >= 50 {
		return
	}
	//在这里就写个页面抓取方法
	blog, err := GetUrl(blog)
	if err != nil {
		return
	}
	log.Println(len(blogs),blog.Url)
	blogs[blog.Url] = blog
	SpiderFollow(blog, blog.Url)
}
func SpiderFollow(blog *Blog, fan string) {
	for _, v := range blog.Follow {
		if vv, ok := blogs[v]; ok {
			b := true
			for _, f := range vv.Fans {
				if f == fan {
					b = false
					break
				}
			}
			if b {
				vv.Fans = append(vv.Fans, fan)
			}
			break
		} else {
			newBlog := &Blog{Url: v, Fans: []string{fan}}
			go Spider(newBlog)
		}

	}
}
func GetUrl(blog *Blog) (*Blog, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/", blog.Url))
	if err != nil {
		return blog, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return blog, err
	}
	title := regexp.MustCompile(`<title>([\S\s\t]*?)</title>`).FindAllString(string(body), -1)
	var tmpTitle = ""
	if len(title) > 0 {
		tmpTitle = title[0][7 : len(title[0])-8]
	}
	tmpTitle = strings.Replace(tmpTitle, "\n", "", -1)
	tmpTitle = strings.Replace(tmpTitle, "\r", "", -1)
	tmpTitle = strings.Trim(tmpTitle, " ")
	if !IsChinese(tmpTitle) {
		return blog,errors.New("err")
	}
	follow := regexp.MustCompile(`http://\w{1,30}.\w{1,30}.\w{2,3}/"|http://\w{1,30}.\w{1,30}.\w{2,3}"|http://\w{1,30}.\w{2,3}/"|http://\w{1,30}.\w{2,3}"`).FindAllString(string(body), -1)

	blog.Title = tmpTitle
	newFollow := make([]string, 0)
	mapFollow := make(map[string]string)
	for _, s := range follow {
		s = s[7:]
		s = strings.Replace(s, `/`, "", -1)
		s = strings.Replace(s, `"`, "", -1)
		if s != blog.Url {
			mapFollow[s] = s
			newFollow = append(newFollow, s)
		}
	}
	for k, _ := range mapFollow {
		newFollow = append(newFollow, k)
	}

	blog.Follow = newFollow
	return blog, nil
}

type Blog struct {
	Url    string
	Title  string
	Follow []string
	Fans   []string
}
type Blogs map[string]*Blog

func InsertMysql() {
	db := mysql.New("tcp", "", "127.0.0.1:3306", "root", "", "blog")
	db.Register("set names utf8")
	err := db.Connect()
	if err != nil {
		panic(err)
	}
	stmt, err := db.Prepare("INSERT INTO `goodblog`(`url`, `title`, `follownum`,`fansnum`,`follow`, `fans`) VALUES (?,?,?,?,?,?)")
	for _, v := range blogs {
		_, err = stmt.Run(v.Url, v.Title, len(v.Follow),len(v.Fans),strings.Join(v.Follow, ","), strings.Join(v.Fans, ","))
		log.Println(v.Url)
	}
	fmt.Println("insert mysql over")
}
func IsChinese(s string) bool{
	for _,v:=range []byte(s) {
		if v>127{
			return true
		}
	}
	return false
}
