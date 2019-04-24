package main

import (
	"bytes"
	"encoding/csv"
	"strings"
	"time"
	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/log"
	"github.com/tidusant/c3m-common/mycrypto"
	rpsex "github.com/tidusant/chadmin-repo/session"

	"flag"

	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {

	var port int
	var debug bool
	var imagefolder string
	//fmt.Println(mycrypto.Encode("abc,efc", 5))

	flag.IntVar(&port, "port", 8083, "help message for flagname")
	flag.BoolVar(&debug, "debug", false, "Indicates if debug messages should be printed in log files")
	flag.StringVar(&imagefolder, "imagefolder", "../upload/images", "Indicates if debug messages should be printed in log files")
	flag.Parse()
	//logLevel := log.DebugLevel
	if !debug {
		//logLevel = log.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	// log.SetOutputFile(fmt.Sprintf("image-"+strconv.Itoa(port)), logLevel)
	// defer log.CloseOutputFile()
	// log.RedirectStdOut()

	log.Infof("running with port:" + strconv.Itoa(port))

	//init config

	router := gin.Default()

	router.GET("/:type/:filepath/:p", func(c *gin.Context) {
		log.Debugf("header:%v", c.Request.Header)
		log.Debugf("Request:%v", c.Request)
		start := time.Now()
		u, err := url.Parse(c.Request.Header.Get("Referer"))
		checkError("get referer", err)
		log.Debugf("referer:%v", u)
		requestDomain := c3mcommon.CheckDomain("http://" + u.Host)
		allowDomain := c3mcommon.CheckDomain(requestDomain)
		strrt := "OK"
		c.Header("Access-Control-Allow-Origin", "*")
		if allowDomain != "" {
			c.Header("Access-Control-Allow-Origin", allowDomain)
			c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
			c.Header("Access-Control-Allow-Credentials", "true")
			//ck := mycrypto.Decode(c.PostForm("data"))
			// mycookie, err := c.Request.Cookie("sex")
			// checkError("get cookie myc", err)
			// ck := mycookie.Value
			// if ck != "" {
			log.Debugf("check request")

			log.Debugf("befor CheckRequest %s", time.Since(start).Nanoseconds())
			if rpsex.CheckRequest(c.Request.URL.Path, c.Request.UserAgent(), c.Request.Referer(), c.Request.RemoteAddr, "GET") {

				filelocal := c.Param("type")
				uploadfolder := imagefolder + "/common/"
				filename := c.Param("filepath")
				if filelocal == "files" {
					shopid := c.Param("p")
					shopid = mycrypto.DecodeA(shopid)
					uploadfolder = imagefolder + "/" + shopid
				} else if filelocal == "customer" {
					shopid := c.Param("p")
					shopid = mycrypto.Decode(shopid)
					session := mycrypto.Decode(filename)
					var data url.Values
					datastr := "cusexport|" + session + "|" + shopid
					log.Debugf("reques response %s", datastr)
					rs := c3mcommon.RequestService(mycrypto.Encode3(datastr), data)

					//write csv
					b := &bytes.Buffer{}
					w := csv.NewWriter(b)

					if err := w.Write([]string{"phone"}); err != nil {
						checkError("error writing record to csv:", err)
					}
					phones := strings.Split(rs, ",")
					for _, phone := range phones {
						if phone != "" {
							var record []string
							record = append(record, phone)
							if err := w.Write(record); err != nil {
								checkError("error writing record to csv:", err)
							}
						}
					}
					w.Flush()

					if err := w.Error(); err != nil {
						checkError("Error w.flush", err)
					}
					c.Header("Content-Description", "File Transfer")
					c.Header("Content-Disposition", "attachment; filename=contacts.csv")
					c.Data(http.StatusOK, "text/csv", b.Bytes())
					//c.String(http.StatusOK, rs)
					return
				} else {
					filename += "/" + c.Param("p")
				}
				log.Debugf("uploadfolder %s", uploadfolder+"/"+filename)
				if _, err := os.Stat(uploadfolder); err == nil {

					http.ServeFile(c.Writer, c.Request, uploadfolder+"/"+filename)
					return
				}
				log.Debugf("NOT ServeFile")

				// } else {
				// 	log.Debugf("check aut fail")
				// }
				// } else {
				// 	log.Debugf("check sesion fail")
				// }
				// } else {
				// 	log.Debugf("check request fail")
				// }

			} else {
				log.Debugf("check ck fail")
			}
		} else {
			log.Debugf("Not allow " + requestDomain)
		}

		c.String(http.StatusOK, strrt)
	})

	// router.GET("/common/template/:name/:file", func(c *gin.Context) {
	// 	uploadfolder := "../upload/images/common/"
	// 	name := c.Param("name")
	// 	file := c.Param("file")
	// 	http.ServeFile(c.Writer, c.Request, uploadfolder+"/"+name+"/"+file)
	// 	return
	// })

	router.Run(":" + strconv.Itoa(port))

}

func checkError(msg string, err error) bool {
	if err != nil {
		log.Debugf(msg+": ", err.Error())
		return false
	}
	return true
}
