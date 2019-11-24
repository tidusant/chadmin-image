package main

import (
	"encoding/json"
	"os"

	"gopkg.in/mgo.v2/bson"

	"github.com/spf13/viper"
	"github.com/tidusant/c3m-common/c3mcommon"
	"github.com/tidusant/c3m-common/log"
	"github.com/tidusant/chadmin-repo/models"
	rpimg "github.com/tidusant/chadmin-repo/vrsgim"

	//	"c3m/common/inflect"
	//	"c3m/log"v test dev

	"flag"
	"fmt"
	"net"
	"net/rpc"
	"strconv"
	"strings"
)

const (
	defaultcampaigncode string = "XVsdAZGVmY"
)

type ImageView struct {
	Key    string
	Value  string
	Status int
}

type Arith int

func (t *Arith) Run(data string, result *models.RequestResult) error {
	log.Debugf("calling with data:" + data)
	*result = models.RequestResult{}
	//parse args
	args := strings.Split(data, "|")

	if len(args) < 3 {
		return nil
	}

	var usex models.UserSession
	usex.Session = args[0]
	usex.Action = args[2]
	info := strings.Split(args[1], "[+]")
	usex.UserID = info[0]
	ShopID := info[1]
	usex.Params = ""
	if len(args) > 3 {
		usex.Params = args[3]
	}
	var shop models.Shop
	usex.Shop = shop
	usex.Shop.ID = bson.ObjectIdHex(ShopID)

	//	} else
	if usex.Action == "la" {

		*result = loadImageAlbum(usex)
	} else if usex.Action == "ri" {
		*result = doRemoveImage(usex)
	}

	return nil
}

func main() {
	var port int
	var debug bool
	flag.IntVar(&port, "port", 7877, "help message for flagname")
	flag.BoolVar(&debug, "debug", false, "Indicates if debug messages should be printed in log files")
	flag.Parse()

	logLevel := log.DebugLevel
	if !debug {
		logLevel = log.InfoLevel

	}

	log.SetOutputFile(fmt.Sprintf("adminDash-"+strconv.Itoa(port)), logLevel)
	defer log.CloseOutputFile()
	log.RedirectStdOut()

	//init db
	arith := new(Arith)
	rpc.Register(arith)
	log.Infof("running with port:" + strconv.Itoa(port))

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(port))
	c3mcommon.CheckError("rpc dail:", err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	c3mcommon.CheckError("rpc init listen", err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func doRemoveImage(usex models.UserSession) models.RequestResult {
	shopid := usex.Shop.ID.Hex()
	filename := usex.Params
	//get config

	uploadfolder := viper.GetString("config.imagefolder") + shopid
	//check folder exist
	if _, err := os.Stat(uploadfolder); os.IsNotExist(err) {
		return c3mcommon.ReturnJsonMessage("0", "folder not found", "", "")

	}
	if rpimg.RemoveImage(shopid, filename) {
		os.Remove(uploadfolder + "/" + filename)
		os.Remove(uploadfolder + "/thumb_" + filename)
	}

	return c3mcommon.ReturnJsonMessage("1", "", "", "")
}
func loadImageAlbum(usex models.UserSession) models.RequestResult {
	shopid := usex.Shop.ID.Hex()
	albumid := usex.Params
	log.Debugf("load image from shopid: " + shopid)
	//loop user directory
	images := rpimg.GetImages(shopid, albumid)
	var imgs []ImageView
	for _, img := range images {
		var viewimg ImageView
		viewimg.Key = img.Filename
		viewimg.Value = img.Filename
		viewimg.Status = 1
		imgs = append(imgs, viewimg)
	}
	str := "[]"
	if len(imgs) > 0 {
		b, _ := json.Marshal(imgs)
		str = string(b)
	}

	log.Debugf("loaded: " + str)

	return c3mcommon.ReturnJsonMessage("1", "", "", str)

}
