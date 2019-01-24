package dkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"hash"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
)

func (d *Dkg) getInterpolationCoefficients(id int) *big.Int {
	topHalf:= big.NewInt(1)
	for _,v:= range d.DecryptionShares {
		if id!=v.Id {
			topHalf.Mul(topHalf,big.NewInt(int64(v.Id)))
		}
	}
	bottomHalf:= big.NewInt(1)
	for _,v:= range d.DecryptionShares {
		if id!=v.Id {
			bottomHalf.Mul(bottomHalf,big.NewInt(int64(v.Id-id)))
		}
	}
	return new(big.Int).Mod(new(big.Int).Div(topHalf,bottomHalf),d.P)
}


func getRandomBigInt() *big.Int {
	min := math.MaxInt8
	max := math.MaxInt16
	return big.NewInt(int64(min + rand.Intn(max-min)))
}

func (d *Dkg) hash(h hash.Hash, paras ...[]byte) []byte {
	data := new(bytes.Buffer)
	for _, v := range paras {
		data.Write(v)
	}
	h.Write(data.Bytes())
	result := h.Sum(nil)
	h.Reset()
	return new(big.Int).Mod(new(big.Int).SetBytes(result), d.P).Bytes()
}

func send(payload interface{}, url string) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if resp.StatusCode != 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err!=nil {
			log.Println(err.Error())
			return
		}
		log.Println(errors.New(string(data)))
		return
	}

	defer resp.Body.Close()
}

func (d *Dkg) computePublicValsProduct(publicVals []*big.Int) *big.Int {
	product := big.NewInt(1)
	for i, v := range publicVals {
		jk := new(big.Int).Exp(big.NewInt(int64(d.Id)), big.NewInt(int64(i)), d.P)
		product= new(big.Int).Mod(product.Mul(product, new(big.Int).Exp(v, jk, d.P)),d.P)
	}
	return product.Mod(product,d.P)
}

func (d *Dkg)combinePublicVals(pb1 []*big.Int, pb2 []*big.Int) []*big.Int {
	combinedPb := make([]*big.Int, len(pb1), len(pb1))
	for i, v := range pb1 {
		combinedPb[i] = new(big.Int).Mod(new(big.Int).Mul(v, pb2[i]),d.P)
	}
	return combinedPb
}

func polynomial(paras []*big.Int, z *big.Int, p *big.Int) *big.Int {
	sum := big.NewInt(0)
	for i, v := range paras {
		tmp := new(big.Int).Exp(z, big.NewInt(int64(i)), p)
		sum.Add(sum, tmp.Mul(tmp, v))
	}
	return sum.Mod(sum, p)
}

func computeShares(f func(*big.Int) *big.Int, n int) []*big.Int {
	shares := make([]*big.Int, n, n)
	for i := 1; i <= n; i++ {
		shares[i-1] = f(big.NewInt(int64(i)))
	}
	return shares
}

func computePublicVals(paras []*big.Int, generator *big.Int, t int, p *big.Int) []*big.Int {
	publicVals := make([]*big.Int, t+1, t+1)
	for i := 0; i <= t; i++ {
		publicVals[i] = new(big.Int).Exp(generator, paras[i], p)
	}
	return publicVals
}

func generateRandomParas(n int) []*big.Int {
	paras := make([]*big.Int, n, n)
	for i, _ := range paras {
		paras[i] = getRandomBigInt()
	}
	return paras
}