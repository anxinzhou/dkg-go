package main

import (
	"dkg"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/rpc"
	"runtime"
	"time"
)

var (
	stage1StartTime time.Time
	stage2StartTime time.Time
	encrytStartTime time.Time
	encrytEndTime time.Time
	decryptStartTime time.Time
	decryptEndTime time.Time
	combineShareStartTime time.Time
	combineShareEndTime time.Time
)

const (
	dkgConfig = "etc/dkgConfig.json"
	peerConfig = "etc/peerConfig.json"
	productPeerConfig = "etc/productPeerConfig.json"
)

func loadDkg(config string,id int, servers []string) (*dkg.Dkg) {
	type dkgConfig struct {
		G_ *big.Int `json:"g_"`
		G *big.Int `json:"g"`
		H *big.Int `json:"h"`
		P *big.Int `json:"p"`
		Q *big.Int `json:"q"`
	}

	var dc dkgConfig
	data,err:= ioutil.ReadFile(config)
	if err !=nil {
		panic(err)
	}
	err =json.Unmarshal(data,&dc)
	if err!=nil {
		panic(err)
	}

	n:= len(servers)
	t:= int(math.Ceil(float64(n)/3))

	return dkg.NewDkg(dc.G,dc.G_,dc.H,dc.P,dc.Q,t,n,id,servers)
}

func loadPeers(hostAddress string ,config string,num int) (int,[]string) {
	type peerConfig struct {
		Servers []string `json:"servers"`
	}

	var pc peerConfig
	data,err:= ioutil.ReadFile(config)
	if err!=nil {
		panic(err)
	}
	err =json.Unmarshal(data,&pc)
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	if len(pc.Servers)<num {
		panic(errors.New("not enough server"))
	}

	var index int
	for i,v:=range pc.Servers {
		if hostAddress ==v {
			index = i+1
			break
		}
	}
	return index, pc.Servers
}

var (
	host string
	port string
	isProduct bool
	num int
)

func init() {
	flag.StringVar(&host,"host","127.0.0.1","http host(default 127.0.0.1)")
	flag.StringVar(&port,"port","4001","http port (default 4000)")
	flag.BoolVar(&isProduct,"p",false,"product mode(default false)")
	flag.IntVar(&num,"num",4,"number of servers")
	flag.Parse()
}

func main() {
	// log
	runtime.GOMAXPROCS(1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)


	uri:= host+":"+port

	var index int
	var servers []string
	if isProduct{
		log.Println("in product mode")
		index,servers =loadPeers(uri,productPeerConfig,num)
	} else {
		log.Println("in test mode")
		index,servers =loadPeers(uri,peerConfig,num)
	}

	s:= dkg.NewDkgServer(loadDkg(dkgConfig, index, servers))

	go dkg.StateTransition(s)
	go dkg.Start(s)

	err:=rpc.Register(s)
	if err!=nil {
		log.Fatal(err.Error())
	}
	rpc.HandleHTTP()
	err = http.ListenAndServe(uri, nil)
	if err!=nil {
		log.Fatal(err.Error())
	}
}

