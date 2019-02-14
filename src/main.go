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
	"runtime"
	"sync"
	"time"
)

const (
	dkgConfig = "etc/dkgConfig.json"
	peerConfig = "etc/peerConfig.json"
	productPeerConfig = "etc/productPeerConfig.json"
	encryptionHost = 1
	encryptionMessage = 20424
)



var (
	startTime time.Time
	stage1StartTime time.Time
	stage2StartTime time.Time
	encrytStartTime time.Time
	encrytEndTime time.Time
	decryptStartTime time.Time
	decryptEndTime time.Time
	combineShareStartTime time.Time
	combineShareEndTime time.Time
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

func stateTransition(d *dkg.Dkg,c chan int) {
	for {
		select {
		case state:= <-c:
			//log.Println("server ",d.Id," :current state:",state)
			switch state {
			case dkg.SendShareStage1:
				stage1StartTime = time.Now()
				log.Println("connection time:", stage1StartTime.Sub(startTime))
				go d.SendStage1()
			case dkg.SendShareStage2:
				stage2StartTime = time.Now()
				log.Println("sending stage1 time:",stage2StartTime.Sub(stage1StartTime))
				go d.SendStage2()
			case dkg.EncrytionStage:
				log.Println("sending stage2 time:", time.Since(stage2StartTime))
				d.SetPublicKey()
				d.SetPrivateKey()
				log.Println("!!!!!! total dkg time:",time.Since(startTime))
				<-time.After(2*time.Second)
				log.Println("----------------------------------")
				log.Println("start encryption and decryption ")
				encrytStartTime = time.Now()
				if d.Id == encryptionHost {
					ciphertext:=d.Encrypt(big.NewInt(encryptionMessage))
					encrytEndTime = time.Now()
					log.Println("encrytion time",encrytEndTime.Sub(encrytStartTime))
					d.Ciphertext = ciphertext
					go d.SendCiphertext(ciphertext)
					c<- dkg.DecryptionStage
				}
			case dkg.DecryptionStage:
				decryptStartTime = time.Now()
				log.Println("receiving encrption time:",decryptStartTime.Sub(encrytStartTime))
				decryptionShare := d.Decrypt(d.Ciphertext)
				decryptEndTime = time.Now()
				log.Println("decrption time:",decryptEndTime.Sub(decryptStartTime))
				d.AppendDecryptionShare(decryptionShare)
				go d.SendDecrptionShare(decryptionShare)
			case dkg.CombineShareStage:
				combineShareStartTime = time.Now()
				log.Println("receiving share time:",combineShareStartTime.Sub(decryptEndTime))
				m:=d.CombineShares()
				combineShareEndTime= time.Now()
				log.Println("combine share time:",combineShareEndTime.Sub(combineShareStartTime))
				log.Println("!!!!!! decryption total time:",combineShareEndTime.Sub(encrytStartTime))
				if m.Cmp(big.NewInt(encryptionMessage))!=0 {
					panic("can not pass text")
				}
			}
		}
	}
}

func connect(d *dkg.Dkg,connected map[int]bool, server string, id int, wg *sync.WaitGroup) {
	client,err:= rpc.DialHTTP("tcp",server)
	if err==nil {
		if client==nil {
			panic("lost client")
		}
		d.RPCClients[id] = client
		connected[id] = true
	} else {
		log.Println(server,"not open")
	}
	wg.Done()
}

func waitAndStart(c chan<- int,d *dkg.Dkg, servers []string) {
	<-time.After(2*time.Second)
	startTime= time.Now()
	connected:=make(map[int]bool)
	var wg sync.WaitGroup
	wg.Add(len(servers)-1)

	for i,v:=range servers {
		if i+1==d.Id || connected[i] {
			continue
		}
		go connect(d,connected,v,i,&wg)
	}

	wg.Wait()
	log.Println("all connected")
	c <- dkg.SendShareStage1
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
	runtime.GOMAXPROCS(1)
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

	s := &dkg.DkgServer{
		D : loadDkg(dkgConfig, index, servers),
		C : make(chan int,1),
	}

	go stateTransition(s.D,s.C)
	go waitAndStart(s.C,s.D,servers)

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

