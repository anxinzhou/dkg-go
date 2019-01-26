package dkg

import (
	"errors"
	"log"
)

type DkgServer struct {
	D *Dkg
	C chan int
}

func(s *DkgServer) SendShareStage1(payload *ShareStage1Payload, reply *int) error {
	log.Println("receive a share")
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
