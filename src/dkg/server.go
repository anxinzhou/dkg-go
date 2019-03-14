package dkg

import (
	"errors"
	"log"
	"math/big"
	"time"
)

type DkgServer struct {
	D *Dkg
	C chan int
}

const (
	CLIENT_SEND_BUFFER = 96
	STAGE_BUFFER = 10
	CONNECTION_WAITING_TIME = 2*time.Second
	AFTER_DKG_WAITING_TIME = 2*time.Second
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

func NewDkgServer(d *Dkg) *DkgServer {
	return &DkgServer{
		D : d,
		C : make(chan int,STAGE_BUFFER),
	}
}

func(s *DkgServer) Start() {
	<-time.After(CONNECTION_WAITING_TIME)
	s.C <- SendShareStage1
}


func (s *DkgServer)StateTransition() {
	for {
		select {
		case state:= <-s.C:
			switch state {
			case SendShareStage1:
				stage1StartTime = time.Now()
				go s.D.SendStage1()
			case SendShareStage2:
				stage2StartTime = time.Now()
				log.Println("sending stage1 time:",stage2StartTime.Sub(stage1StartTime))
				go s.D.SendStage2()
			case EncrytionStage:
				log.Println("sending stage2 time:", time.Since(stage2StartTime))
				s.D.SetPublicKey()
				s.D.SetPrivateKey()
				log.Println("!!!!!! total dkg time:",time.Since(stage1StartTime))
				<-time.After(AFTER_DKG_WAITING_TIME)
				log.Println("----------------------------------")
				log.Println("start encryption and decryption ")
				encrytStartTime = time.Now()
				if s.D.Id == encryptionHost {
					ciphertext:=s.D.Encrypt(big.NewInt(encryptionMessage))
					encrytEndTime = time.Now()
					log.Println("encrytion time",encrytEndTime.Sub(encrytStartTime))
					s.D.Ciphertext = ciphertext
					go s.D.SendCiphertext(ciphertext)
					s.C<- DecryptionStage
				}
			case DecryptionStage:
				decryptStartTime = time.Now()
				log.Println("receiving encrption time:",decryptStartTime.Sub(encrytStartTime))
				decryptionShare := s.D.Decrypt(s.D.Ciphertext)
				decryptEndTime = time.Now()
				log.Println("decrption time:",decryptEndTime.Sub(decryptStartTime))
				s.D.AppendDecryptionShare(decryptionShare)
				go s.D.SendDecrptionShare(decryptionShare)
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

func(s *DkgServer) SendShareStage1(payload *ShareStage1Payload, reply *int) error {
	reply = nil
	if s.D.IsQualifiedPeerForStage1(payload) {
		length:=s.D.AppendQualifiedPeerShare(&PeerShare{
			Id: payload.Id,
			Share:payload.Share1,
		})
		if length == s.D.N {
			s.C <- SendShareStage2
		}
	} else {
		log.Fatal("invalid share in stage1")
		return errors.New("invalid share in stage1")
	}
	return nil
}

func(s* DkgServer) SendShareStage2(payload *ShareStage2Payload, reply *int) error {
	reply = nil
	if(s.D.IsQualifiedPeerForStage2(payload)) {
		length:=s.D.AppendQualifiedPeerPublicVal(&PeerPublicVal{
			Id:payload.Id,
			PublicVal:payload.PublicVals[0],
		})
		if length==s.D.N {
			s.C <- EncrytionStage
		}
	} else {
		log.Fatal("invalid share in stage2")
		return errors.New("invalid share for stage2")
	}
	return nil
}

func(s *DkgServer) SendCiphertext(payload *Ciphertext, reply *int) error {
	reply = nil
	if s.D.IsCiphertextValid(payload) {
		s.D.Ciphertext = payload
		s.C <- DecryptionStage
	} else {
		log.Fatal("invalid ciphertext")
		return errors.New("invalid ciphertext")
	}
	return nil
}

func(s *DkgServer) SendDecryptionShare(payload *DecryptionShare, reply *int) error {
	reply = nil
	if s.D.IsDecryptionShareValid(payload) {
		length:=s.D.AppendDecryptionShare(payload)
		if length == s.D.N  {
			s.C <- CombineShareStage
		}
	} else {
		log.Fatal("invalid decryption share")
		return errors.New("invalid decryption share")
	}
	return nil
}
