// get_whois_server project main.go
/*
	根据数据库中的后缀,爬取最新的域名后缀以及对应的whois server 信息
*/
package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"net/http"
	"net/url"
	"strings"

	"database/sql"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	suffix, err := get_suffix()
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range suffix {
		msg, err := get_site(v)
		if err != nil {
			fmt.Println(err)
		}
		server := Parsewhois(msg)
		server = strings.Replace(server, " ", "", -1)
		err = tosql(v, server)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func get_site(domain string) (result string, err error) {
	u, _ := url.Parse("https://www.iana.org/whois?q=" + domain)
	q := u.Query()

	u.RawQuery = q.Encode()
	fmt.Println(u.String())
	res, err := http.Get(u.String())
	if err != nil {

		return
	}
	re, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {

		return
	}
	result = string(re)
	return
}
func get_suffix() (suffixs []string, err error) {
	db, err := sql.Open("mysql", "root:root@/whois")
	if err != nil {
		return
	}
	defer db.Close()
	rows, err := db.Query("select suffix from conf_rootserver")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var temp string
		err = rows.Scan(&temp)
		suffixs = append(suffixs, temp)
	}
	return
}
func Parsewhois(whois string) string {
	return parserone(regexp.MustCompile(`(?i)whois:+(.*?)(\n|$)`), 1, whois)
}

//*******************

func parserone(re *regexp.Regexp, group int, data string) (one string) {
	var result []string
	found := re.FindAllStringSubmatch(data, 1)
	if len(found) == 0 {
		return
	}
	if len(found) > 0 {
		for _, one := range found {
			if len(one) >= 2 && len(one[group]) > 0 {
				result = appendIfMissing(result, one[group])

			}
		}
	}
	if len(result) == 0 {
		return
	}
	one = result[0]
	return
}
func appendIfMissing(slice []string, i string) []string {

	i = strings.ToLower(i)

	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}

	return append(slice, i)

}
func tosql(suf string, nserver string) (err error) {
	db, err := sql.Open("mysql", "root:root@/whois")
	if err != nil {
		return
	}
	defer db.Close()
	stmt, err := db.Prepare("update conf_rootserver set server=? where suffix=?")
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(nserver, suf)
	if err != nil {
		return
	}
	return
}
