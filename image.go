package main

import (
	"fmt"
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
				log.Debugf("check sesion")
				//if rpsex.CheckSession(ck) {
				// log.Debugf("check aut")
				// client, err := rpc.Dial("tcp", viper.GetString("RPCname.aut"))
				// checkError("dial RPCAuth check login", err)
				// reply := ""
				// userIP, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
				// autCall := client.Go("Arith.Run", ck+"|"+userIP+"|"+"aut", &reply, nil)
				// autreplyCall := <-autCall.Done
				// checkError("RPCAuth.Go", autreplyCall.Error)
				// client.Close()

				// //RPC call
				// if reply != "" {
				//log.Debugf("get folder")
				//info := strings.Split(reply, "[+]")
				//userid := info[0]
				//shopid := info[1]

				//userid := reply
				log.Debugf("before decode %s", time.Since(start).Nanoseconds())
				shopid := c.Param("p")
				shopid = mycrypto.DecodeA(shopid)
				filelocal := c.Param("type")
				uploadfolder := imagefolder + "/common/"
				filename := c.Param("filepath")
				if filelocal == "files" {
					uploadfolder = imagefolder + "/" + shopid
				} else {
					filename += "/" + c.Param("p")
				}
				log.Debugf("after decode %s", time.Since(start).Nanoseconds())
				log.Debugf("type %s, filepath %s, p %s", filelocal, filename, c.Param("p"))
				log.Debugf("uploadfolder %s", uploadfolder)
				if _, err := os.Stat(uploadfolder); err == nil {
					log.Debugf("ServeFile")
					log.Debugf("after checkfolder %s", time.Since(start).Nanoseconds())
					http.ServeFile(c.Writer, c.Request, uploadfolder+"/"+filename)
					log.Debugf("after ServeFile %s", time.Since(start).Nanoseconds())
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
