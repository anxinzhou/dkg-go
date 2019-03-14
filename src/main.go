package main

import (
	"dkg"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"net/rpc"
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

func loadPeers(hostAddress string ,config string) (int,[]string) {
	type peerConfig struct {
		Servers []string `json:"servers"`
	}

	var pc peerConfig
	data,err:= ioutil.ReadFile(config)
	if err!=nil {
		panic(err)
	}
	json.Unmarshal(data,&pc)
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
)

func init() {
	flag.StringVar(&host,"host","127.0.0.1","http host(default 127.0.0.1)")
	flag.StringVar(&port,"port","4001","http port (default 4000)")
	flag.BoolVar(&isProduct,"p",false,"product mode(default false)")
	flag.Parse()
}

func main() {
	// log
	//runtime.GOMAXPROCS(1)
	log.SetFlags(log.LstdFlags | log.Lshortfile)


	uri:= host+":"+port

	var index int
	var servers []string
	if isProduct{
		log.Println("in product mode")
		index,servers =loadPeers(uri,productPeerConfig)
	} else {
		log.Println("in test mode")
		index,servers =loadPeers(uri,peerConfig)
	}

	s:= dkg.NewDkgServer(loadDkg(dkgConfig, index, servers))

	go s.StateTransition()
	go s.Start()

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

