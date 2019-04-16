package main

import (
	"crypto/sha256"
	"flag"
	"log"
	"math/big"
	"math/rand"
	"time"
)

const hashSize = 32
const checkPoint = 10000

// 2^224 -1

var timeStamp = time.Now().UTC().UnixNano()
var r = rand.Int63()

func hash(data []byte) [hashSize]byte {
	return sha256.Sum256(data)
}

func getDifficultyFromLambda(lambda int) *big.Int {
	diffB := make([]byte, lambda)
	for i, _ := range diffB {
		diffB[i] = '1'
	}
	diff, _ := new(big.Int).SetString(string(diffB), 2)
	return diff
}

func cal(difficulty *big.Int, duration time.Duration) (int64,time.Duration) {
	// generate public key
	var pubKey *big.Int
	for {
		pk := make([]byte, hashSize)
		_, err := rand.Read(pk)
		if err != nil {
			panic(err)
		}
		pubKey = new(big.Int).SetBytes(pk)
		if pubKey.BitLen() == hashSize*8 {
			break
		}
	}
	//log.Println("public key", pubKey)
	// other part
	prefix := new(big.Int).Xor(big.NewInt(r|timeStamp), pubKey)

	// calculate nonce
	startTime := time.Now()
	targetTime := startTime.Add(duration)
	var nonce int64 = 0
	for {
		nb := big.NewInt(nonce)
		nb.Xor(nb, prefix)
		h := hash(nb.Bytes())
		v := new(big.Int).SetBytes(h[:])
		if v.Cmp(difficulty) < 0 {
			break;
		}
		if (nonce+1)%checkPoint == 0 {
			if time.Now().After(targetTime) {
				log.Println("time out", "current nonce", nonce)
				nonce = -1
				break;
			}
		}
		nonce = nonce + 1
	}
	endTime := time.Now()
	return nonce,endTime.Sub(startTime)
}

var t int
var clientNum int
var lambda int

func init() {
	flag.IntVar(&t, "duration", 60, "time of computing")
	flag.IntVar(&clientNum, "num", 1, "number of committee")
	flag.IntVar(&lambda,"lambda",229,"difficulty of mining")
	flag.Parse()
}

func main() {
	difficulty := getDifficultyFromLambda(lambda)
	duration := time.Duration(t) * time.Second
	nonce,cost:=cal(difficulty,duration)
	if nonce!=-1 {
		log.Println("pass, time:",cost)
	} else {
		log.Println("fail, time:",cost)
	}
}
