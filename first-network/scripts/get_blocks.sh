#!/bin/bash
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
CHANNEL_NAME=mychannel
i=0

while [ $i -lt $1 ]; do
	peer channel fetch $i block_$i.pb -o orderer.example.com:7050 -c $CHANNEL_NAME --tls --cafile $ORDERER_CA
	configtxlator proto_decode --input block_$i.pb --type common.Block | jq . > block_$i.json
	rm block_$i.pb
	i=$[$i+1]
done
