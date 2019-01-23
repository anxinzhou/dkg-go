package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"dkg"
)

const (
	url1 = "/shareStage1"
	url2 = "/shareStage2"
	url3 = "/ciphertext"
	url4 = "/decryptionShare"
	serverConfig = "etc/serverConfig.json"
	dkgConfig = "etc/dkgConfig.json"
	peerConfig = "etc/peerConfig.json"
)

const (
	t = 3
	n = 5
)

func postShareStage1(d *dkg.Dkg)  func (http.ResponseWriter, *http.Request) {
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

		from:= payload.From
		share1:= payload.Share1
		share2:= payload.Share2
		combinedPublicVals:=payload.CombinedPublicVals
		if(d.IsQualifiedPeerForStage1(from,share1,share2,combinedPublicVals)) {
			d.AppendQualifiedPeerShare(share1)
			if len(d.QualifiedPeerShares) == d.N {
				d.SetShare()
				go d.SendStage2(url2)
			}
		} else {
			http.Error(w,"vals is not qualified for stage1", http.StatusBadRequest)
		}
	}
}

func postShareStage2(d *dkg.Dkg) func (http.ResponseWriter, *http.Request) {
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

		from:= payload.From
		share:= payload.Share
		publicVals:= payload.PublicVals
		if(d.IsQualifiedPeerForStage2(from,share,publicVals)) {
			d.AppendQualifiedPeerPublicVal(publicVals[0])
			if len(d.QualifiedPeerPublicVals)==d.N {
				d.SetPublicVal()
				if(d.NeedEncrpyt) {
					go d.SendCiphertext()
				}
			}
		} else {
			http.Error(w,"vals is not qualified for stage2", http.StatusBadRequest)
		}
	}
}

func postCyphertext(d *dkg.Dkg) func (http.ResponseWriter, *http.Request) {
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
			d.SendDecrptionShare(&payload)
		} else {
			http.Error(w,"invalid ciphertext", http.StatusBadRequest)
		}
	}
}

func postDecryptionShare(d *dkg.Dkg) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload dkg.DecreptionShare
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

		d.AppendDecreptionShare(&payload)
		if len(d.DecryptionShares) == d.T+1 {
			go d.
		}
	}
}

func loadServerConfig(config string) (address string) {
	type serverConfig struct {
		Host string `json:"host"`
		Port int `json:"port"`
	}

	var sc serverConfig
	data,err := ioutil.ReadFile(config)
	if err!=nil {
		panic(err)
	}
	json.Unmarshal(data,&sc)
	return sc.Host + string(sc.Port)
}

func loadDkg(config string,id int, servers []string) (*dkg.Dkg) {
	type dkgConfig struct {
		G *big.Int `json:"g"`
		H *big.Int `json:"h"`
		P *big.Int `json:"p"`
	}

	var dc dkgConfig
	data,err:= ioutil.ReadFile(config)
	if err !=nil {
		panic(err)
	}
	json.Unmarshal(data,&dc)
	return dkg.NewDkg(dc.G,dc.H,dc.P,t,n,id,servers)
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

func main() {
	// log

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	host:=loadServerConfig(serverConfig)
	index,servers:=loadPeers(host,peerConfig)
	d:= loadDkg(dkgConfig, index, servers)
	go d.SendStage1(url1)

	r:=mux.NewRouter()

	r.HandleFunc(url1,postShareStage1(d)).Methods("POST")
	r.HandleFunc(url2,postShareStage2(d)).Methods("POST")
	r.HandleFunc(url3,postCyphertext(d)).Methods("POST")
	r.HandleFunc(url4,postDecryptionShare(d)).Methods("POST")
	http.ListenAndServe(loadServerConfig(serverConfig),r)
}

