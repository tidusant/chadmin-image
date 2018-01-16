package main

import (
	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/log"
	rpsex "github.com/tidusant/chadmin-repo/session"

	"flag"
	"fmt"

	"net/http"
	"net/url"
	"os"
	"strconv"

	"net"
	"net/rpc"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/spf13/viper"
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

	logLevel := log.DebugLevel
	if !debug {
		logLevel = log.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	log.SetOutputFile(fmt.Sprintf("image-"+strconv.Itoa(port)), logLevel)
	defer log.CloseOutputFile()
	log.RedirectStdOut()

	log.Infof("running with port:" + strconv.Itoa(port))

	//init config

	router := gin.Default()
	// router.GET("/c/:ck", func(c *gin.Context) {

	// 	u, err := url.Parse(c.Request.Header.Get("Referer"))
	// 	checkError("get referer", err)
	// 	requestDomain := c3mcommon.CheckDomain("http://" + u.Host)
	// 	allowDomain := c3mcommon.CheckDomain(requestDomain)
	// 	strrt := ""
	// 	c.Header("Access-Control-Allow-Origin", "*")
	// 	if allowDomain != "" {
	// 		c.Header("Access-Control-Allow-Origin", allowDomain)
	// 		c.Header("Access-Control-Allow-Headers", "access-control-allow-origin, access-control-allow-headers,access-control-allow-credentials")
	// 		c.Header("Access-Control-Allow-Credentials", "true")

	// 		params := strings.Split(mycrypto.Decode(c.Param("ck")), "|")

	// 		if len(params) > 1 {
	// 			ck := params[0]
	// 			shopid := params[1]
	// 			log.Debugf("cookie:%s", ck)
	// 			if ck != "" {
	// 				if rpsex.CheckRequest(c.Request.URL.Path, c.Request.UserAgent(), c.Request.Referer(), c.Request.RemoteAddr, "POST") {
	// 					if rpsex.CheckSession(ck) {

	// 						//							expiration := time.Now().Add(365 * 24 * time.Hour)
	// 						//							cookie := http.Cookie{Name: "myc", Value: mycrypto.Encode(ck, 3), Expires: expiration, Path: "/", Domain: "192.168.1.221:8083"}
	// 						//							log.Debugf("set cookie myc:%s", ck)
	// 						//							http.SetCookie(c.Writer, &cookie)
	// 						//							cookie = http.Cookie{Name: "pos", Value: mycrypto.Encode(shopid, 4), Expires: expiration, Path: "/", Domain: "192.168.1.221:8083"}
	// 						//							log.Debugf("set cookie pos:%s", shopid)
	// 						//							http.SetCookie(c.Writer, &cookie)
	// 						strrt += "alert(1);var d = new Date();"
	// 						strrt += "d.setTime(d.getTime() + 24*60*60*1000);"
	// 						strrt += "var expires = \"expires=\"+ d.toUTCString();"
	// 						strrt += "document.cookie =\"myc=" + mycrypto.Encode(ck, 4) + ";\"+expires+\";path=/\";"
	// 						strrt += "document.cookie =\"pos=" + mycrypto.Encode(shopid, 3) + ";\"+expires+\";path=/\";"
	// 					} else {
	// 						log.Debugf("check session fail")
	// 					}
	// 				} else {
	// 					log.Debugf("check request fail")
	// 				}
	// 			}
	// 		}
	// 	} else {
	// 		log.Debugf("Not allow " + requestDomain)
	// 	}
	// 	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<script>"+strrt+"</script>"))
	// })

	router.GET("/:type/:filepath/*p", func(c *gin.Context) {
		log.Debugf("header:%v", c.Request.Header)
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
			mycookie, err := c.Request.Cookie("sex")
			checkError("get cookie myc", err)
			ck := mycookie.Value

			if ck != "" {
				log.Debugf("check request")
				log.Debugf("ck:%s", ck)
				if rpsex.CheckRequest(c.Request.URL.Path, c.Request.UserAgent(), c.Request.Referer(), c.Request.RemoteAddr, "GET") {
					log.Debugf("check sesion")
					if rpsex.CheckSession(ck) {
						log.Debugf("check aut")
						client, err := rpc.Dial("tcp", viper.GetString("RPCname.aut"))
						checkError("dial RPCAuth check login", err)
						reply := ""
						userIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
						autCall := client.Go("Arith.Run", ck+"|"+userIP+"|"+"aut", &reply, nil)
						autreplyCall := <-autCall.Done
						checkError("RPCAuth.Go", autreplyCall.Error)

						//RPC call
						if reply != "" {
							log.Debugf("get folder")
							info := strings.Split(reply, "[+]")
							//userid := info[0]
							shopid := info[1]

							//userid := reply
							filelocal := c.Param("type")
							uploadfolder := imagefolder + "/common/"
							filename := c.Param("filepath")
							if filelocal == "files" {
								uploadfolder = imagefolder + "/" + shopid
							} else {
								filename += "/" + c.Param("p")
							}

							if _, err := os.Stat(uploadfolder); err == nil {
								http.ServeFile(c.Writer, c.Request, uploadfolder+"/"+filename)
								return
							}

						} else {
							log.Debugf("check aut fail")
						}
					} else {
						log.Debugf("check sesion fail")
					}
				} else {
					log.Debugf("check request fail")
				}

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
		log.Errorf(msg+": ", err.Error())
		return false
	}
	return true
}
