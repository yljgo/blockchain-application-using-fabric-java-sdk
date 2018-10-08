package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type EvidenceChaincode struct{}

type Evidence struct {
	ID      string `json:"id"`
	IDCard  string `json:"id_card"`
	Name    string `json:name`
	Hash    string `json:hash`
	Content string `json:content`
}

type ResInfo struct {
	Status bool   `json:"status"`
	Msg    string `json:"msg"`
}

func (t *ResInfo) error(msg string) {
	t.Status = false
	t.Msg = msg
}
func (t *ResInfo) ok(msg string) {
	t.Status = true
	t.Msg = msg
}

func (t *ResInfo) response() pb.Response {
	resJson, err := json.Marshal(t)
	if err != nil {
		return shim.Error("Failed to generate json result " + err.Error())
	}
	return shim.Success(resJson)
}

func main() {
	err := shim.Start(new(EvidenceChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

type process func(shim.ChaincodeStubInterface, []string) *ResInfo

// Init initializes chaincode
// ===========================
func (t *EvidenceChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *EvidenceChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "newEvidence" { //create a new evidence
		return t.newEvidence(stub, args)
	} else if function == "queryEvidence" {
		return t.queryEvidence(stub, args)
	}
	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// 写入证据
func (e *EvidenceChaincode) newEvidence(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return e.handleProcess(stub, args, 5, func(shim.ChaincodeStubInterface, []string) *ResInfo {
		ri := &ResInfo{true, ""}
		_id := args[0]
		_idcard := args[1]
		_name := args[2]
		_hash := args[3]
		_content := args[4]
		_evidence := &Evidence{_id, _idcard, _name, _hash, _content}
		_ejson, err := json.Marshal(_evidence)

		if err != nil {
			ri.error(err.Error())
		} else {
			_old, err := stub.GetState(_id)
			if err != nil {
				ri.error(err.Error())
			} else if _old != nil {
				ri.error("the evidence has exists")
			} else {
				err := stub.PutState(_id, _ejson)
				if err != nil {
					ri.error(err.Error())
				} else {
					ri.ok("")
				}
			}
		}
		return ri
	})
}

func (e *EvidenceChaincode) queryEvidence(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return e.handleProcess(stub, args, 1, func(shim.ChaincodeStubInterface, []string) *ResInfo {
		ri := &ResInfo{true, ""}
		queryString := args[0]
		queryResults, err := getQueryResultForQueryString(stub, queryString)
		if err != nil {
			ri.error(err.Error())
		} else {
			ri.ok(string(queryResults))
		}
		return ri
	})
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

func (t *EvidenceChaincode) handleProcess(stub shim.ChaincodeStubInterface, args []string, expectNum int, f process) pb.Response {
	res := &ResInfo{false, ""}
	err := t.checkArgs(args, expectNum)
	if err != nil {
		res.error(err.Error())
	} else {
		res = f(stub, args)
	}
	return res.response()
}

func (t *EvidenceChaincode) checkArgs(args []string, expectNum int) error {
	if len(args) != expectNum {
		return fmt.Errorf("Incorrect number of arguments. Expecting  " + strconv.Itoa(expectNum))
	}
	for p := 0; p < len(args); p++ {
		if len(args[p]) <= 0 {
			return fmt.Errorf(strconv.Itoa(p+1) + "nd argument must be a non-empty string")
		}
	}
	return nil
}
