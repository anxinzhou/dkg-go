package dkg

import (
	"crypto/sha256"
	"log"
	"math/big"
	"math/rand"
	"net/rpc"
	"sync"
	"time"
)

// state machine
const (
	InitialStage = iota
	SendShareStage1
	SendShareStage2
	EncrytionStage
	DecryptionStage
	CombineShareStage
	encryptionHost = 1
)

const (
	START_CONNECTION_WAITING_TIME = 10*time.Second
	START_ENCRYPT_WAITING_TIME = 14*time.Second
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

type ShareStage1Payload struct {
	Id               int
	Share1             *big.Int
	Share2             *big.Int
	CombinedPublicVals []*big.Int
}

type ShareStage2Payload struct {
	Id       int        `json:"id"`
	Share      *big.Int   `json:"share"`
	PublicVals []*big.Int `json:"publicVals"`
}

type Ciphertext struct {
	C  *big.Int `json:"c"`
	U  *big.Int `json:"u"`
	U_ *big.Int `json:"u_"`
	E  *big.Int `json:"e"`
	F  *big.Int `json:"f"`
}

type DecryptionShare struct {
	Id int      `json:"id"`
	U  *big.Int `json:"u"`
	CiphertextU *big.Int `json:"ciphertextU"`
	E  *big.Int `json:"e"`
	F  *big.Int `json:"f"`
	H *big.Int `json:"h"`
}

type PeerShare struct {
	Id int `json:"id"`
	Share *big.Int `json:"share"`
}

type PeerPublicVal struct {
	Id int `json:"id"`
	PublicVal *big.Int `json:"publicVal"`
}

type Client struct {
	C *rpc.Client
	Mutex sync.Mutex
	ConnectedCount int
}

type Dkg struct {
	G                  *big.Int
	G_                 *big.Int
	H                  *big.Int
	P                  *big.Int
	Q   			   *big.Int
	Id                 int
	T                  int
	N                  int
	Servers            []string
	Shares1            []*big.Int
	Shares2            []*big.Int
	PublicVals1        []*big.Int
	CombinedPublicVals []*big.Int

	shareMutex          *sync.Mutex
	QualifiedPeerShares []*PeerShare

	publicValMutex          *sync.Mutex
	QualifiedPeerPublicVals []*PeerPublicVal

	decryptionShareMutex *sync.Mutex
	DecryptionShares     []*DecryptionShare
	Ciphertext		   *Ciphertext

	PublicKey          *big.Int
	PrivateKey         *big.Int

	RPCClients   []*Client
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func NewDkg(g *big.Int,g_ *big.Int, h *big.Int, p *big.Int, q *big.Int, t int, n int, id int, servers []string) *Dkg {
	d := &Dkg{
		Id:                   id,
		G:                    g,
		G_:                   g_,
		H:                    h,
		P:                    p,
		Q:					  q,
		T:                    t,
		N:                    n,
		Servers:              servers,
		shareMutex:           &sync.Mutex{},
		publicValMutex:       &sync.Mutex{},
		decryptionShareMutex: &sync.Mutex{},
	}

	paras1 := generateRandomParas(t+1)
	paras2 := generateRandomParas(t+1)

	d.Shares1 = computeShares(func(z *big.Int) *big.Int {
		return polynomial(paras1, z, q)
	}, n)

	d.Shares2 = computeShares(func(z *big.Int) *big.Int {
		return polynomial(paras2, z, q)
	}, n)

	d.PublicVals1 = computePublicVals(paras1, g, t, p)
	d.CombinedPublicVals = d.combinePublicVals(d.PublicVals1, computePublicVals(paras2, h, t, p))

	d.QualifiedPeerShares = make([]*PeerShare, 1, n)
	d.QualifiedPeerShares[0] = &PeerShare{
		Id: id,
		Share:d.Shares1[id-1],
	}
	d.QualifiedPeerPublicVals = make([]*PeerPublicVal, 1, n)
	d.QualifiedPeerPublicVals[0] = &PeerPublicVal{
		Id:id,
		PublicVal:d.PublicVals1[0],
	}

	d.RPCClients = make([]*Client,d.N,d.N)
	for i,_:=range d.RPCClients {
		d.RPCClients[i]=&Client{
			Mutex: sync.Mutex{},
		}
	}

	return d
}

func (d *Dkg) AppendDecryptionShare(decryptionShare *DecryptionShare) int {
	d.decryptionShareMutex.Lock()
	defer d.decryptionShareMutex.Unlock()
	d.DecryptionShares = append(d.DecryptionShares, decryptionShare)
	return len(d.DecryptionShares)
}

func (d *Dkg) AppendQualifiedPeerShare(share *PeerShare) int {
	d.shareMutex.Lock()
	defer d.shareMutex.Unlock()
	d.QualifiedPeerShares = append(d.QualifiedPeerShares, share)
	return len(d.QualifiedPeerShares)
}

func (d *Dkg) AppendQualifiedPeerPublicVal(publicVal *PeerPublicVal) int {
	d.publicValMutex.Lock()
	defer d.publicValMutex.Unlock()
	d.QualifiedPeerPublicVals = append(d.QualifiedPeerPublicVals, publicVal)
	return len(d.QualifiedPeerPublicVals)
}

func (d *Dkg) IsQualifiedPeerForStage1(payload *ShareStage1Payload) bool {
	share1:= payload.Share1
	share2:= payload.Share2
	combinedPublicVals:=payload.CombinedPublicVals

	if len(combinedPublicVals) != d.T+1 {
		log.Println("len of combined public vals is not equal to t+1")
		return false
	}

	gsij:= new(big.Int).Exp(d.G,share1,d.P)
	hsij:= new(big.Int).Exp(d.H,share2,d.P)
	gMulh := new(big.Int).Mod(new(big.Int).Mul(gsij,hsij),d.P)
	product := d.computePublicValsProduct(combinedPublicVals)
	if gMulh.Cmp(product) == 0 {
		return true
	} else {
		return false
	}
}

func (d *Dkg) IsQualifiedPeerForStage2(payload *ShareStage2Payload) bool {
	share:= payload.Share
	publicVal:= payload.PublicVals
	if len(publicVal) != d.T+1 {
		log.Println("len of public vals is not equal to t+1")
		return false
	}

	gsij:= new(big.Int).Exp(d.G,share,d.P)
	if gsij.Cmp(d.computePublicValsProduct(publicVal)) == 0 {
		return true
	} else {
		return false
	}
}

func (d *Dkg) Connect(server string, id int) {
	if d.RPCClients[id].C!=nil {
		return
	}

	d.RPCClients[id].Mutex.Lock()
	defer d.RPCClients[id].Mutex.Unlock()
	if d.RPCClients[id].C != nil {
		return
	}
	for {
		client,err:= rpc.DialHTTP("tcp",server)
		if err==nil {
			if client==nil {
				log.Println("lost client")
				continue
			}
			d.RPCClients[id].C = client
			d.RPCClients[id].ConnectedCount +=1
			break;
		} else {
			log.Println(server,"not open")
		}
	}
	if d.RPCClients[id].ConnectedCount >1 {
		log.Println("abnormal connection, id",id,"connect count", d.RPCClients[id].ConnectedCount)
	}
	//log.Println("connected")
	//wg.Done()
}

func (d *Dkg) Broadcast(handler func(i int)) {
	for i, v := range d.Servers {
		if i+1 == d.Id {
			continue
		}

		if d.RPCClients[i].C!=nil {
			handler(i)
		} else {
			go func(v string, i int, handler func(i int)) {
				d.Connect(v, i)
				handler(i)
			}(v,i,handler)
		}
	}
}

func (d *Dkg) SendStage1() {
	d.Broadcast(func(i int){
		var reply int
		d.RPCClients[i].C.Go("DkgServer.SendShareStage1",&ShareStage1Payload{
			Id:               d.Id,
			Share1:             d.Shares1[i],
			Share2:             d.Shares2[i],
			CombinedPublicVals: d.CombinedPublicVals,
		},&reply,nil)
	})
}

func (d *Dkg) SendStage2() {
	d.Broadcast(func(i int) {
		var reply int
		d.RPCClients[i].C.Go("DkgServer.SendShareStage2",&ShareStage2Payload{
			Id: d.Id,
			Share: d.Shares1[i],
			PublicVals: d.PublicVals1,
		},&reply,nil)
	})
}

func (d *Dkg) SendCiphertext(ciphertext *Ciphertext) {
	d.Broadcast(func(i int) {
		var reply int
		d.RPCClients[i].C.Go("DkgServer.SendCiphertext",ciphertext,&reply,nil)
	})
}

func (d *Dkg) SendDecrptionShare(decryptionShare *DecryptionShare) {
	d.Broadcast(func(i int) {
		var reply int
		d.RPCClients[i].C.Go("DkgServer.SendDecryptionShare",decryptionShare,&reply,nil)
	})
}

func (d *Dkg) SetPublicKey() {
	d.PublicKey = big.NewInt(1)
	for _, v := range d.QualifiedPeerPublicVals {
		d.PublicKey = new(big.Int).Mod(new(big.Int).Mul(d.PublicKey, v.PublicVal),d.P)
	}
}

func (d *Dkg) SetPrivateKey() {
	d.PrivateKey = big.NewInt(0)
	for _, v := range d.QualifiedPeerShares {
		d.PrivateKey.Add(d.PrivateKey, v.Share)
	}
	d.PrivateKey.Mod(d.PrivateKey, d.Q)
}

func (d *Dkg) Encrypt(m *big.Int) *Ciphertext {


	// encryption
	r := getRandomBigInt()
	s := getRandomBigInt()
	hr := new(big.Int).Exp(d.PublicKey, r, d.P)
	hashOfhr := new(big.Int).SetBytes(d.hash(sha256.New(), hr.Bytes()))

	c := new(big.Int).Xor(hashOfhr, m)
	u := new(big.Int).Exp(d.G, r, d.P)
	w := new(big.Int).Exp(d.G, s, d.P)
	u_ := new(big.Int).Exp(d.G_, r, d.P)
	w_ := new(big.Int).Exp(d.G_, s, d.P)
	e := new(big.Int).SetBytes(d.hash(sha256.New(), c.Bytes(), u.Bytes(), w.Bytes(), u_.Bytes(), w_.Bytes()))
	f := new(big.Int).Mod(new(big.Int).Add(s, new(big.Int).Mul(r, e)),d.P)

	return &Ciphertext{
		C:  c,
		U:  u,
		U_: u_,
		E:  e,
		F:  f,
	}
}

func (d *Dkg) Decrypt(ciphertext *Ciphertext) *DecryptionShare {
	u := ciphertext.U
	g := d.G
	xi := d.PrivateKey
	si := getRandomBigInt()

	ui := new(big.Int).Exp(u, xi, d.P)
	ui_ := new(big.Int).Exp(u, si, d.P)
	hi_ := new(big.Int).Exp(g, si, d.P)
	ei := new(big.Int).SetBytes(d.hash(sha256.New(), ui.Bytes(), ui_.Bytes(), hi_.Bytes()))
	fi := new(big.Int).Mod(new(big.Int).Add(si, new(big.Int).Mul(xi, ei)), d.P)
	hi:= new(big.Int).Exp(d.G,xi,d.P)

	return &DecryptionShare{
		Id: d.Id,
		U:  ui,
		E:  ei,
		F:  fi,
		H: hi,
		CiphertextU:ciphertext.U,
	}
}

func (d *Dkg) CombineShares() *big.Int {

	shares:= d.DecryptionShares[:d.T+1]

	productU:= big.NewInt(1)
	for _,v:= range shares {
		interpolation:= d.getInterpolationCoefficients(shares,v.Id)
		tmp := new(big.Int).Exp(v.U,interpolation,d.P)
		productU.Mul(productU,tmp)
		productU.Mod(productU,d.P)
	}

	hOfProductU:= new(big.Int).SetBytes(d.hash(sha256.New(), productU.Bytes()))
	m:=new(big.Int).Xor(hOfProductU,d.Ciphertext.C)
	return m
}

func (d *Dkg) IsDecryptionShareValid(decryptionShare *DecryptionShare) bool {
	ei := decryptionShare.E
	ui := decryptionShare.U
	fi := decryptionShare.F
	hi := decryptionShare.H


	ufi:= new(big.Int).Exp(decryptionShare.CiphertextU,fi,d.P)
	uiei:= new(big.Int).Exp(ui,ei,d.P)
	ui_:= new(big.Int).Mod(new(big.Int).Mul(ufi,new(big.Int).ModInverse(uiei,d.P)),d.P)

	gfi:= new(big.Int).Exp(d.G,fi,d.P)
	hiei:= new(big.Int).Exp(hi,ei,d.P)
	hi_:= new(big.Int).Mod(new(big.Int).Mul(gfi,new(big.Int).ModInverse(hiei,d.P)),d.P)

	hashR:= new(big.Int).SetBytes(d.hash(sha256.New(),ui.Bytes(),ui_.Bytes(),hi_.Bytes()))

	if ei.Cmp(hashR) == 0 {
		return true
	} else {
		return false
	}
}

func (d *Dkg) IsCiphertextValid(ciphertext *Ciphertext) bool {
	c := ciphertext.C
	u := ciphertext.U
	u_ := ciphertext.U_
	e := ciphertext.E
	f := ciphertext.F

	gf := new(big.Int).Exp(d.G, f, d.P)
	ue := new(big.Int).ModInverse(new(big.Int).Exp(u, e, d.P), d.P)
	w := new(big.Int).Mod(new(big.Int).Mul(gf, ue), d.P)

	_gf := new(big.Int).Exp(d.G_, f, d.P)
	_ue := new(big.Int).ModInverse(new(big.Int).Exp(u_, e, d.P), d.P)
	w_ := new(big.Int).Mod(new(big.Int).Mul(_gf, _ue), d.P)

	hashR := new(big.Int).SetBytes(d.hash(sha256.New(), c.Bytes(), u.Bytes(), w.Bytes(), u_.Bytes(), w_.Bytes()))

	if e.Cmp(hashR) == 0 {
		return true
	} else {
		return false
	}
}

func timeSince(startTime string, waitTime time.Duration) time.Duration {
	t,err:=time.Parse(time.UnixDate,startTime)
	if err!=nil {
		panic(err)
	}
	t.Add(waitTime)
	t2:=time.Now().UTC()
	diff:=t.Sub(t2)
	log.Println("wait time:",diff)
	return diff
}


func StateTransition(s *DkgServer,startTime string) {
	<-time.After(timeSince(startTime,START_CONNECTION_WAITING_TIME))
	// send stage 1
	stage1StartTime = time.Now()
	go s.D.SendStage1()
	for {
		select {
		case state:= <-s.C:
			switch state {
			case SendShareStage2:
				stage2StartTime = time.Now()
				log.Println("receive stage1 time:",stage2StartTime.Sub(stage1StartTime))
				go s.D.SendStage2()
			case EncrytionStage:
				log.Println("receive stage2 time:", time.Since(stage2StartTime))
				s.D.SetPublicKey()
				s.D.SetPrivateKey()
				log.Println("!!!!!! total dkg time:",time.Since(stage1StartTime))
				<-time.After(timeSince(startTime,START_ENCRYPT_WAITING_TIME))
				log.Println("----------------------------------")
				log.Println("start encryption and decryption ")
				encrytStartTime = time.Now()
				if s.D.Id == encryptionHost {
					ciphertext:=s.D.Encrypt(big.NewInt(encryptionMessage))
					encrytEndTime = time.Now()
					log.Println("encrytion time",encrytEndTime.Sub(encrytStartTime))
					go s.D.SendCiphertext(ciphertext)
					s.D.Ciphertext = ciphertext
					s.C<- DecryptionStage
				}
			case DecryptionStage:
				decryptStartTime = time.Now()
				log.Println("receiving encrption time:",decryptStartTime.Sub(encrytStartTime))
				decryptionShare := s.D.Decrypt(s.D.Ciphertext)
				decryptEndTime = time.Now()
				log.Println("decrption time:",decryptEndTime.Sub(decryptStartTime))
				go s.D.SendDecrptionShare(decryptionShare)
				s.D.AppendDecryptionShare(decryptionShare)
			case CombineShareStage:
				combineShareStartTime = time.Now()
				log.Println("receiving share time:",combineShareStartTime.Sub(decryptEndTime))
				m:= s.D.CombineShares()
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
