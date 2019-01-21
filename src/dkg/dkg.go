package dkg

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"hash"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

type ShareStage1Payload struct {
	From int `json:"from"`
	Share1 *big.Int `json:"share1"`
	Share2 *big.Int `json:"share2"`
	CombinedPublicVals []*big.Int `json:"combinedPublicVals"`
}

type ShareStage2Payload struct {
	From int `json:"from"`
	Share *big.Int `json:"share"`
	PublicVals []*big.Int `json:"publicVals"`
}

type Dkg struct {
	G *big.Int
	H *big.Int
	P *big.Int
	Id int
	T int
	N int
	Servers []string
	SecretShare *big.Int
	PublicVal *big.Int
	Shares1 []*big.Int
	Shares2 []*big.Int
	PublicVals1 []*big.Int
	CombinedPublicVals []*big.Int

	shareMutex *sync.Mutex
	publicValMutex *sync.Mutex
	QualifiedPeerShares []*big.Int
	QualifiedPeerPublicVals []*big.Int
}

func NewDkg(g*big.Int, h *big.Int, p *big.Int, t int, n int, id int, servers []string) *Dkg {
	d:= &Dkg{
		Id:id,
		G:g,
		H:h,
		P:p,
		T:t,
		N:n,
		Servers:servers,
		shareMutex:&sync.Mutex{},
		publicValMutex:&sync.Mutex{},
	}

	paras1:= generateRandomParas(n)
	paras2:= generateRandomParas(n)

	d.Shares1 = computeShares(func(z *big.Int) *big.Int {
		return polynomial(paras1,z,p)
	}, n)

	d.Shares2 = computeShares(func(z *big.Int) *big.Int {
		return polynomial(paras2,z,p)
	}, n)

	d.PublicVals1 = computePublicVals(paras1,g,t,p)
	d.CombinedPublicVals = combinePublicVals(d.PublicVals1, computePublicVals(paras2,h,t,p))

	d.QualifiedPeerShares = make([]*big.Int,1,n)
	d.QualifiedPeerShares[0] = d.Shares1[id-1]
	d.QualifiedPeerPublicVals = make([]*big.Int,1,n)
	d.QualifiedPeerPublicVals[0] = d.CombinedPublicVals[0]

	return d
}

func generateRandomParas(n int) []*big.Int {
	min:=math.MaxInt8
	max:=math.MaxInt16
	paras:=make([]*big.Int,n,n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i,_:=range paras {
		paras[i] = big.NewInt(int64(min+rand.Intn(max-min)))
	}
	return paras
}

func (d *Dkg) AppendQualifiedPeerShare(share *big.Int) {
	d.shareMutex.Lock()
	defer d.shareMutex.Unlock()
	d.QualifiedPeerShares=append(d.QualifiedPeerShares, share)
}

func (d *Dkg) AppendQualifiedPeerPublicVal(publicVal *big.Int) {
	d.publicValMutex.Lock()
	defer d.publicValMutex.Unlock()
	d.QualifiedPeerPublicVals = append(d.QualifiedPeerPublicVals, publicVal)
}

func polynomial(paras []*big.Int, z *big.Int, p *big.Int) *big.Int {
	sum:= big.NewInt(0)
	for i,v:= range paras {
		tmp:= new(big.Int).Exp(z,big.NewInt(int64(i)),p)
		sum.Add(sum,tmp.Mul(tmp,v))
	}
	return sum.Mod(sum, p)
}

func computeShares(f func(*big.Int) *big.Int, n int) []*big.Int {
	shares:= make([]*big.Int,n,n)
	for i:=1;i<=n;i++ {
		shares[i]=f(big.NewInt(int64(i)))
	}
	return shares
}

func computePublicVals(paras []*big.Int, generator *big.Int, t int, p *big.Int) []*big.Int {
	publicVals:=make([]*big.Int,t+1,t+1)
	for i:=0;i<=t;i++ {
		publicVals[i] = new(big.Int).Exp(generator,paras[i],p)
	}
	return publicVals
}

func combinePublicVals(pb1 []*big.Int, pb2 []*big.Int) []*big.Int {
	combinedPb:=make([]*big.Int, len(pb1),len(pb1))
	for i,v:=range pb1 {
		combinedPb[i] = new(big.Int).Mul(v, pb2[i])
	}
	return combinedPb
}

func (d *Dkg)IsQualifiedPeerForStage1(peerId int, share1 *big.Int, share2 *big.Int, combinedPublicVals []*big.Int) bool {
	if len(combinedPublicVals)!=d.T+1 {
		log.Println("len of combined public vals is not equal to t+1")
		return false
	}
	gMulh:= big.NewInt(0).Mul(share1,share2)
	product:= d.computePublicValsProduct(peerId, combinedPublicVals)
	if gMulh.Cmp(product) ==0 {
		return true
	} else {
		return false
	}
}

func (d *Dkg) computePublicValsProduct(peerId int,publicVals []*big.Int) *big.Int {
	product:=big.NewInt(0)
	for i,v:= range publicVals {
		jk:= big.NewInt(0).Exp(big.NewInt(int64(peerId)),big.NewInt(int64(i)),d.P)
		product.Add(product,big.NewInt(0).Exp(v,jk,d.P))
	}
	return product
}

func (d *Dkg)IsQualifiedPeerForStage2(peerId int, share *big.Int, publicVal []*big.Int) bool {
	if len(publicVal)!=d.T+1 {
		log.Println("len of public vals is not equal to t+1")
		return false
	}
	if share.Cmp(d.computePublicValsProduct(peerId,publicVal))==0 {
		return true
	} else {
		return false
	}
}

func (d *Dkg) sendStage1(share1 *big.Int,share2 *big.Int, url string) {
	payload:= &ShareStage1Payload{
		From: d.Id,
		Share1: share1,
		Share2: share2,
		CombinedPublicVals: d.CombinedPublicVals,
	}
	data,err:= json.Marshal(payload)
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	req,err:=http.NewRequest("POST",url, bytes.NewBuffer(data))
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	client:=&http.Client{}
	resp,err:=client.Do(req)
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	defer resp.Body.Close()
}

func (d *Dkg) SendStage1(url string) {
	for i,v:= range d.Servers {
		if i+1 == d.Id {
			continue
		}
		go d.sendStage1(d.Shares1[i],d.Shares2[i],v+url)
	}
}

func (d *Dkg) sendStage2(share1 *big.Int, url string) {
	payload := &ShareStage2Payload{
		From: d.Id,
		Share: share1,
		PublicVals: d.PublicVals1,
	}

	data,err:= json.Marshal(payload)
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	req,err:=http.NewRequest("POST",url, bytes.NewBuffer(data))
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	client:=&http.Client{}
	resp,err:=client.Do(req)
	if err!=nil {
		log.Println(err.Error())
		panic(err)
	}

	defer resp.Body.Close()
}

func (d *Dkg) SendStage2(url string) {
	for i,v:= range d.Servers {
		if i+1 == d.Id {
			continue
		}
		go d.sendStage2(d.Shares1[i],v+url)
	}
}

func (d *Dkg) SetShare() {
	d.SecretShare = big.NewInt(0)
	for _,v:=range d.QualifiedPeerShares {
		d.SecretShare.Add(d.SecretShare,v)
	}
	d.SecretShare.Mod(d.SecretShare,d.P)
}

func (d *Dkg) SetPublicVal() {
	d.PublicVal = big.NewInt(0)
	for _,v:=range d.QualifiedPeerPublicVals {
		d.PublicVal.Add(d.PublicVal,v)
	}
	d.PublicVal.Mod(d.PublicVal,d.P)
}

func (d *Dkg) hash(h hash.Hash,paras ...[]byte) []byte {
	data:= new(bytes.Buffer)
	for _,v:=range paras {
		data.Write(v)
	}
	h.Write(data.Bytes())
	result:=h.Sum(nil)
	h.Reset()
	return new(big.Int).Mod(new(big.Int).SetBytes(result),d.P).Bytes()
}

func (d *Dkg)encrypt(m *big.Int) (*big.Int,*big.Int,*big.Int,*big.Int,*big.Int) {
	min:=math.MaxInt8
	max:=math.MaxInt16
	rand.Seed(time.Now().UTC().UnixNano())
	hfunc:= sha256.New()

	// encryption
	r:=big.NewInt(int64(min+rand.Intn(max-min)))
	s:=big.NewInt(int64(min+rand.Intn(max-min)))
	_g:=big.NewInt(3)
	hr:= new(big.Int).Exp(d.PublicVal,r,d.P)
	hashOfhr:=d.hash(hfunc,hr.Bytes())

	c:= new(big.Int).Mod(new(big.Int).Xor(new(big.Int).SetBytes(hashOfhr),m),d.P)
	u:= new(big.Int).Exp(d.G,r,d.P)
	w:= new(big.Int).Exp(d.G,s,d.P)
	_u:= new(big.Int).Exp(_g,r,d.P)
	_w:= new(big.Int).Exp(_g,s,d.P)
	e:= new(big.Int).SetBytes(d.hash(hfunc,c.Bytes(),u.Bytes(),w.Bytes(),_u.Bytes(),_w.Bytes()))
	f:= new(big.Int).Add(s,new(big.Int).Mul(r,e))
	return c,u,_u,e,f
}


func (d *Dkg) TestEncyptionAndDecryption() {
	m:=big.NewInt(20)
	log.Println("message ",m)
	c,u,_u,e,f:=d.encrypt(m)
	// decryption
}
