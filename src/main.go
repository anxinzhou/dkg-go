package main

import (
	"dkg"
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"runtime"
	"time"
)

const (
	urlShareStage1 = "/shareStage1"
	urlShareStage2 = "/shareStage2"
	urlCiphertext = "/ciphertext"
	urlDecryptionShare = "/decryptionShare"
	serverConfig = "etc/serverConfig.json"
	dkgConfig = "etc/dkgConfig.json"
	peerConfig = "etc/peerConfig.json"
	encryptionHost = 1
	encryptionMessage = 20424
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

func postShareStage1(d *dkg.Dkg, c chan<- int)  func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload dkg.ShareStage1Payload
		data,err:= ioutil.ReadAll(r.Body)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
			return
		}
		err =json.Unmarshal(data,&payload)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
		}

		if d.IsQualifiedPeerForStage1(&payload) {
			length:=d.AppendQualifiedPeerShare(&dkg.PeerShare{
				Id: payload.Id,
				Share:payload.Share1,
			})
			if length == d.N {
				c <- dkg.SendShareStage2
			}
		} else {
			log.Println("invalid")
			http.Error(w,"vals is not qualified for stage1", http.StatusBadRequest)
		}
	}
}

func postShareStage2(d *dkg.Dkg, c chan<- int) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload dkg.ShareStage2Payload
		data,err:= ioutil.ReadAll(r.Body)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
			return
		}
		err =json.Unmarshal(data,&payload)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
		}

		Id:= payload.Id
		publicVals:= payload.PublicVals
		if(d.IsQualifiedPeerForStage2(&payload)) {
			length:=d.AppendQualifiedPeerPublicVal(&dkg.PeerPublicVal{
				Id:Id,
				PublicVal:publicVals[0],
			})
			if length==d.N {
				c <- dkg.EncrytionStage
			}
		} else {
			http.Error(w,"vals is not qualified for stage2", http.StatusBadRequest)
		}
	}
}

func postCiphertext(d *dkg.Dkg, c chan <- int) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload dkg.Ciphertext
		data,err:= ioutil.ReadAll(r.Body)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
			return
		}
		err =json.Unmarshal(data,&payload)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
		}

		if d.IsCiphertextValid(&payload) {
			d.Ciphertext = &payload
			c <- dkg.DecryptionStage
		} else {
			http.Error(w,"invalid ciphertext", http.StatusBadRequest)
		}
	}
}

func postDecryptionShare(d *dkg.Dkg, c chan<- int) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload dkg.DecryptionShare
		data,err:= ioutil.ReadAll(r.Body)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w,err.Error(),http.StatusBadRequest)
			return
		}
		err =json.Unmarshal(data,&payload)
		if err!=nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if d.IsDecryptionShareValid(&payload) {
			length:=d.AppendDecryptionShare(&payload)
			if length == d.N  {
				c <- dkg.CombineShareStage
			}
		} else {
			http.Error(w,"invalid decryption share", http.StatusBadRequest)
		}
	}
}

func ping(w http.ResponseWriter, r *http.Request) {

}

func loadServerConfig(config string) (host,port string) {
	type serverConfig struct {
		Host string `json:"host"`
		Port string `json:"port"`
	}

	var sc serverConfig
	data,err := ioutil.ReadFile(config)
	if err!=nil {
		panic(err)
	}
	err =json.Unmarshal(data,&sc)
	if err!=nil {
		panic(err)
	}
	return sc.Host, sc.Port
}

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
				go d.SendStage1(urlShareStage1)
			case dkg.SendShareStage2:
				stage2StartTime = time.Now()
				log.Println("sending stage1 time:",stage2StartTime.Sub(stage1StartTime))
				go d.SendStage2(urlShareStage2)
			case dkg.EncrytionStage:
				encrytStartTime = time.Now()
				log.Println("sending stage2 time:", encrytStartTime.Sub(stage2StartTime))
				d.SetPublicKey()
				d.SetPrivateKey()
				log.Println("total dkg time:",time.Since(stage1StartTime))

				if d.Id == encryptionHost {
					ciphertext:=d.Encrypt(big.NewInt(encryptionMessage))
					encrytEndTime = time.Now()
					log.Println("encrytion time",encrytEndTime.Sub(encrytStartTime))
					d.Ciphertext = ciphertext
					go d.SendCiphertext(ciphertext,urlCiphertext)
					c<- dkg.DecryptionStage
				}
			case dkg.DecryptionStage:
				decryptStartTime = time.Now()
				if d.Id == encryptionHost {
					log.Println("receiving encrption time:",decryptStartTime.Sub(encrytEndTime))
				}
				decryptionShare := d.Decrypt(d.Ciphertext)
				decryptEndTime = time.Now()
				log.Println("decrption time:",decryptEndTime.Sub(decryptStartTime))
				d.AppendDecryptionShare(decryptionShare)
				go d.SendDecrptionShare(decryptionShare,urlDecryptionShare)
			case dkg.CombineShareStage:
				combineShareStartTime = time.Now()
				log.Println("receiving share time:",combineShareStartTime.Sub(decryptEndTime))
				m:=d.CombineShares()
				combineShareEndTime= time.Now()
				log.Println("combine share time:",combineShareEndTime.Sub(combineShareStartTime))
				log.Println("total time:",combineShareEndTime.Sub(stage1StartTime))
				if m.Cmp(big.NewInt(encryptionMessage))!=0 {
					panic("can not pass text")
				}
			}
		}
	}
}

func waitAndStart(c chan<- int,d *dkg.Dkg, servers []string) {
	connected:=make(map[int]bool)

	for{
		if len(connected) == len(servers) {
			break;
		}
		for i,v:=range servers {
			if connected[i] {
				continue
			}
			_, err := http.Get(v)
			if err ==nil {
				connected[i] = true
			}
		}
		runtime.Gosched()
	}


	c <- dkg.SendShareStage1
}

var (
	host string
	port string
)

func init() {
	flag.StringVar(&host,"host","127.0.0.1","http host(default 127.0.0.1)")
	flag.StringVar(&port,"port","4001","http port (default 4000)")
	flag.Parse()
}

func main() {
	// log

	log.SetFlags(log.LstdFlags | log.Lshortfile)



	uri:= "http://"+host+":"+port

	index,servers:=loadPeers(uri,peerConfig)
	d:= loadDkg(dkgConfig, index, servers)
	c:=make(chan int,1)
	go stateTransition(d,c)
	go waitAndStart(c,d,servers)

	r:=mux.NewRouter()
	r.HandleFunc(urlShareStage1,postShareStage1(d,c)).Methods("POST")
	r.HandleFunc(urlShareStage2,postShareStage2(d,c)).Methods("POST")
	r.HandleFunc(urlCiphertext,postCiphertext(d,c)).Methods("POST")
	r.HandleFunc(urlDecryptionShare,postDecryptionShare(d,c)).Methods("POST")
	r.HandleFunc("/ping",ping).Methods("GET")
	err:=http.ListenAndServe(host+":"+port,r)
	if err!=nil {
		panic(err)
	}
}

