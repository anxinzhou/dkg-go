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

func loadPeers(hostAddress string ,config string,num int) ([]string) {
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

	return pc.Servers[:num]
}

var (
	host string
	port string
	isProduct bool
	num int
	index int
	startTime string
)

func init() {
	flag.StringVar(&host,"host","127.0.0.1","http host(default 127.0.0.1)")
	flag.StringVar(&port,"port","4001","http port (default 4000)")
	flag.BoolVar(&isProduct,"p",false,"product mode(default false)")
	flag.IntVar(&num,"num",4,"number of servers")
	flag.IntVar(&index,"index",1,"index of this server")
	flag.StringVar(&startTime,"startTime","","startTime of program")
	flag.Parse()
}

func start() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	uri:= host+":"+port

	var servers []string
	if isProduct{
		log.Println("in product mode")
		servers =loadPeers(uri,productPeerConfig,num)
	} else {
		log.Println("in test mode")
		servers =loadPeers(uri,peerConfig,num)
	}

	s:= dkg.NewDkgServer(loadDkg(dkgConfig, index, servers))

	go dkg.StateTransition(s,startTime)

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


func main() {
	start()
}



