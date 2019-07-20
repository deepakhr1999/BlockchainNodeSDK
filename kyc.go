package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

/*
	Unlike the previous chaincode, this is user centred.
	This stores data about what banks the user has opted for KYC.
	*Algorithms are dead in this code*
	To do:
		- finish writing basic chaincode
		- write query function to return multiple users, filtered by banks
*/


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Response struct{
	Flag   bool `json:"flag"`
	Message string `json:"message"`
}

type User struct{
	Hash 	string `json:"hash"`
	Name 	string `json:"name"`
	Banks 	[]Flag `json:"banks"`
	Requests []string `json:"requests"`
}

type Flag struct{
	ID 			string `json:"bank"`
	Pending 	bool `json:"pending"`
	Approved 	bool `json:"approved"`
}

func res(flag bool, message string) []byte{
	result := Response{Flag: flag, Message: message,}
	resbytes, _ := json.Marshal(result)
	return resbytes
}


func (t *SimpleChaincode) Accept(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	result := Response{Flag: false}
	if(len(args) != 4){
		return shim.Success(res(false, "Expecting 4 args"))
	} 

	userID, userHash, bank, decision := args[0], args[1], args[2], args[3]

	result = t.Auth(stub, userID, userHash)
	if(!result.Flag){
		resultbytes, _ := json.Marshal(result) 
		return shim.Success(resultbytes)
	}

	// We did not bail, so user credentials are valid
	userbytes, _ := stub.GetState(userID)
	if userbytes==nil{
		return shim.Success(res(false, "User does not exist"))
	}
	var User User
	json.Unmarshal(userbytes, &User)

	// flow maintains the fact that requests array contains 'bank'
	result.Flag = false
	for i, b := range User.Requests{
		if(b == bank){
			result.Flag = true
			n := len(User.Requests)
			User.Requests[i] = User.Requests[n-1]
			User.Requests = User.Requests[:n-1]
			break
		}
	}

	if(! result.Flag){
		result.Message = "Bank not found in requests"
		resultbytes, _ := json.Marshal(result) 
		return shim.Success(resultbytes)
	}

	//If decision is "No" then just YEET after deleting
	if(decision == "No"){
		userbytes, _ = json.Marshal(User)
		_ = stub.PutState(userID, userbytes)
		return shim.Success(res(true, "Removed "+bank+" from Requests"))
	}

	//append the bank in user flags and save
	flag := Flag{ ID: bank, Pending: true, Approved: false}
	User.Banks = append(User.Banks, flag)
	userbytes, _ = json.Marshal(User)
	_ = stub.PutState(userID, userbytes)
	return shim.Success(res(true,"Added "+bank+" to pending lists"))
}

func (t *SimpleChaincode) Endorse(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	result := Response{Flag: false}
	if(len(args) != 4){
		return shim.Success(res(false, "Expecting 4 args"))
	} 

	bankID, bankHash, userID, decision := args[0], args[1], args[2], args[3]

	result = t.Auth(stub, bankID, bankHash)
	if(!result.Flag){
		resultbytes, _ := json.Marshal(result) 
		return shim.Success(resultbytes)
	}

	// bank credentials are valid
	userbytes, _ := stub.GetState(userID)
	var User User
	json.Unmarshal(userbytes, &User)

	/* Endorse KYC for user by bank*/
	for _, flag:= range User.Banks{
		if(flag.ID 	== bankID){
			flag.Pending = false
			if decision!="No"{
				flag.Approved = true
			}else {
				flag.Approved =false
			}
		}
	}
	//save changes
	userbytes, _ = json.Marshal(User)
	_ = stub.PutState(userID, userbytes)
	return shim.Success(res(decision!="No", "KYC for "+userID+" is done"))
}

func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response{
	return shim.Success(nil)
}

func (t *SimpleChaincode) Request(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	result := Response{Flag: false}
	if(len(args) != 3){
		return shim.Success(res(false, "Expecting 3 args"))
	} 

	bankID, bankHash, userID := args[0], args[1], args[2]

	result = t.Auth(stub, bankID, bankHash)
	if(!result.Flag){
		resultbytes, _ := json.Marshal(result) 
		return shim.Success(resultbytes)
	}

	//validated credentials
	state_b, err := stub.GetState(userID)
	if err != nil {
		return shim.Success(res(false, "User does not exist"))
	}

	var User User
	json.Unmarshal(state_b, &User)
	User.Requests = append(User.Requests, bankID)
	state_b, _ = json.Marshal(User)
	stub.PutState(userID, state_b)
	return shim.Success(res(true, "Request has been sent to "+userID))
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// fmt.Println("ex02 Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "Query" {
		return t.Query(stub, args)
	}else if function == "Join"{
		return t.Join(stub, args)
	}else if function == "Endorse"{
		return t.Endorse(stub, args)
	}else if function == "Delete"{
		return t.Delete(stub, args)
	}else if function == "QueryPrivate"{
		return t.QueryPrivate(stub, args)
	}else if function == "Request"{
		return t.Request(stub, args)
	}else if function == "Accept"{
		return t.Accept(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting Query, Accept, Request, Join, Endorse, Delete, QueryPrivate")
}

func (t *SimpleChaincode) Auth(stub shim.ChaincodeStubInterface, ID string, hash string) Response{
	var result Response
	result.Flag = false

	state_b, err := stub.GetState(ID)
	if err != nil {
		result.Message = "User does not exist"
		return result
	}

	var person User
	json.Unmarshal(state_b, &person)

	if(person.Hash == hash){
		result.Flag = true
		return result
	}else{
		result.Message = "Password does not match"
		return result
	}
}
/*Query the public state of the ledger*/
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	result := Response{Flag: false}
	if(len(args)!= 1){
		result.Message = "Expecting 1 arg"
		resultbytes, _ := json.Marshal(result) 
		return shim.Success(resultbytes)
	}
	
	var name = args[0]

	// Get the state from the ledger
	state_b, err := stub.GetState(name)
	if err != nil {
		return shim.Error("Could not get state")
	}

	//return the state as bytes
	return shim.Success(state_b)
}

func (t *SimpleChaincode) Join(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	result := Response{Flag: false}
	if(len(args)!=3){
		result.Message = "Expecting 3 args"
		resultbytes, _ := json.Marshal(result) 
		return shim.Success(resultbytes)
	}
	//assemble data
	userID, userHash, username := args[0], args[1], args[2]

	//check if user exists?!
	state, _ := stub.GetState(userID)
	if state != nil{
		return shim.Success(res(false, "Given user already exists"))
	}
	User := User{Hash: userHash, Name: username,}
	userbytes, _ := json.Marshal(User)
	_ = stub.PutState(userID, userbytes)
	return shim.Success(res(true, userID+" has joined the network"))
}


// Gets data from private data collection
func (t *SimpleChaincode) QueryPrivate(stub shim.ChaincodeStubInterface, args []string) pb.Response{
	if len(args)!=2{
		return shim.Error("Incorrect arguments, expecting 2")
	}

	username := args[0]
	coll := args[1]
	//check if the state under username has been deleted
	state_b, err := stub.GetState(username)
	if state_b == nil {
		return shim.Error("User does not exist")
	}

	private_b, err := stub.GetPrivateData(coll, username) 
     if err != nil {
             return shim.Error("Failed to get private details for "+username)
     } else if private_b == nil {
             return shim.Error("Private details do not exist for "+username)
     }
	return shim.Success(private_b)
}

// Deletes an entity from state
func (t *SimpleChaincode) Delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	username := args[0]

	// Delete the key from the state in ledger
	err := stub.DelState(username)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
