//package main
//
//import (
//	"crypto/sha256"
//	"flag"
//	"log"
//	"math/big"
//	"math/rand"
//	"time"
//)
//
//const hashSize = 32
//const checkPoint = 100000
//
//// 2^224 -1
//
//var timeStamp = time.Now().UTC().UnixNano()
//var r = rand.Int63()
//
//func hash(data []byte) [hashSize]byte {
//	return sha256.Sum256(data)
//}
//
//func getDifficultyFromLambda(lambda int) *big.Int {
//	diffB := make([]byte, lambda)
//	for i, _ := range diffB {
//		diffB[i] = '1'
//	}
//	diff, _ := new(big.Int).SetString(string(diffB), 2)
//	return diff
//}
//
//func cal(difficulty *big.Int, duration time.Duration, nonces chan int64) {
//	// generate public key
//	var pubKey *big.Int
//	for {
//		pk := make([]byte, hashSize)
//		_, err := rand.Read(pk)
//		if err != nil {
//			panic(err)
//		}
//		pubKey = new(big.Int).SetBytes(pk)
//		if pubKey.BitLen() == hashSize*8 {
//			break
//		}
//	}
//	//log.Println("public key", pubKey)
//	// other part
//	prefix := new(big.Int).Xor(big.NewInt(r|timeStamp), pubKey)
//
//	// calculate nonce
//	startTime := time.Now()
//	//targetTime := startTime.Add(duration)
//	var nonce int64 = 0
//	for {
//		nb := big.NewInt(nonce)
//		nb.Xor(nb, prefix)
//		h := hash(nb.Bytes())
//		v := new(big.Int).SetBytes(h[:])
//		if v.Cmp(difficulty) < 0 {
//			break;
//		}
//		//if (nonce+1)%checkPoint == 0 {
//		//	//if time.Now().After(targetTime) {
//		//	//	log.Println("time out", "current nonce", nonce)
//		//	//	nonce = -1
//		//	//	break;
//		//	//}
//		//	log.Println(nonce)
//		//}
//		nonce = nonce + 1
//	}
//	endTime := time.Now()
//	log.Println("finish time:", endTime.Sub(startTime),"nonces",nonce)
//	nonces <- nonce
//}
//
//var t int
//var clientNum int
//var lambda int
//
//func init() {
//	flag.IntVar(&t, "duration", 60, "time of computing")
//	flag.IntVar(&clientNum, "num", 1, "number of committee")
//	flag.IntVar(&lambda,"lambda",229,"")
//}
//
//func main() {
//	log.Println("timeStamp,", timeStamp)
//	lambda := 229
//	difficulty := getDifficultyFromLambda(lambda)
//	duration := time.Duration(t) * time.Second
//	nonces := make(chan int64, clientNum)
//	for i := 0; i < clientNum; i++ {
//		go cal(difficulty, duration, nonces)
//	}
//
//	successCount := 0
//outer:
//	for {
//		select {
//		case <-nonces:
//			successCount += 1
//		case <-time.After(duration):
//			log.Println("time out")
//			break outer
//		}
//	}
//	if successCount>=1 {
//		log.Println("pass, count:",successCount)
//	}
//}
