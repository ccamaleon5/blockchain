/*
* Adrian Pareja
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"crypto/rand"
	"encoding/hex"
)

// UUID layout variants.
const (
	VariantNCS = iota
	VariantRFC4122
	VariantMicrosoft
	VariantFuture
)

// Used in string method conversion
const dash byte = '-'

// UUID representation compliant with specification
// described in RFC 4122.
type UUID [16]byte

//Wallet - Structure for products used in buy goods
type Wallet struct {
	Id        string    `json:"id"`
	email     string  `json:"email"`
	phone     string  `json:"phone"`
	document  string `json:"document"`
	password  string `json:"password"` 
	amount    float64 `json:"amount"` 
}

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	fmt.Printf("Iniciandooo Contrato Wallet....")
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error Iniciando Wallet Smart Contract: %s", err)
	}
}

// Init reinicia los estados del ledger
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Número de Argumentos incorrecto. Se esperaba 1 argumento")
	}

	err := stub.PutState("coinBalance", []byte(args[0]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke Punto de entrada a cualquier función del ledger
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running..FUNCTION:" + function)

	if function == "createwallet" {
		return t.createWallet(stub, args)
	} else if function == "transfer" {
		return t.transfer(stub, args)
	}
	fmt.Println("invoke no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// Query es nuestro punto de entrada de querys
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running FUNCTION:" + function)

	// Manejar diferentes funciones
	if function == "getbalance" {
		return t.getBalance(stub, args)
	}else if function == "gettotalcoin"{
		return t.getTotalCoin(stub, args)
	}
	fmt.Println("query no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// createWallet - invocar esta funcion para crear un wallet con saldo inicial
func (t *SimpleChaincode) createWallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion createWallet---")
	if len(args) != 5 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 5 para createWallet")
	}

	walletId := NewV4()
	fmt.Printf("UUIDv4: %s\n", walletId)
	amt, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		fmt.Println("Error Float parsing")
		return nil, errors.New("Error marshaling wallet")
	}

	wallet := Wallet{
		Id:        walletId.String(),
		email:     args[0],
		phone:     args[1],
		document:  args[2],
		password:  args[3],
		amount:    amt,
	}

	bytes, err := json.Marshal(wallet)
	if err != nil {
		fmt.Println("Error marshaling wallet")
		return nil, errors.New("Error marshaling wallet")
	}

	err = stub.PutState(wallet.Id, bytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// transfer - invocar esta funcion para transferir coins de un wallet a otro
func (t *SimpleChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion Transfer---")
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 3 para createWallet")
	}
	
	fmt.Printf("WalletId 1: %s\n", args[0])
	fmt.Printf("WalletId 2: %s\n", args[1])
	fmt.Printf("Monto: %s\n", args[2])

	bytesWallet1, err1 := stub.GetState(args[0])
	
	walletSender := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletSender)
	
	fmt.Println(walletSender)
	if err1 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}

	bytesWallet2, err2 := stub.GetState(args[1])
	walletReceiver := Wallet{}
	err = json.Unmarshal(bytesWallet2, &walletReceiver)
	if err2 != nil {
		fmt.Println("Error retrieving " + args[1])
		return nil, errors.New("Error retrieving " + args[1])
	}
	
	amt, err := strconv.ParseFloat(args[2], 64)
	
	walletSender.amount = walletSender.amount-amt //debita el monto
	walletReceiver.amount = walletReceiver.amount+amt //carga el monto

	walletSenderJSONasBytes, _ := json.Marshal(walletSender)
	err = stub.PutState(args[0], walletSenderJSONasBytes) //rewrite the wallet
	
	if err != nil {
		return nil, err
	}
	
	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[1], walletReceiverJSONasBytes) //rewrite the wallet
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

func (t *SimpleChaincode) getBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getBalance() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("wallet id is ")
	fmt.Println(walletId)
	bytes, err := stub.GetState(args[0])
	if err != nil {
		fmt.Println("Error retrieving " + walletId)
		return nil, errors.New("Error retrieving " + walletId)
	}
	wallet := Wallet{}
	err1 := json.Unmarshal(bytes, &wallet)
	
	fmt.Println(wallet.amount)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}
	
	return []byte(strconv.FormatFloat(wallet.amount,'f',6,64)), nil
}

func (t *SimpleChaincode) getTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getTotalCoin() is running----")

	bytes, err := stub.GetState("coinBalance")
	fmt.Println(bytes)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}
	
	return bytes, nil
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

// SetVersion sets version bits.
func (u *UUID) SetVersion(v byte) {
	u[6] = (u[6] & 0x0f) | (v << 4)
}

// SetVariant sets variant bits as described in RFC 4122.
func (u *UUID) SetVariant() {
	u[8] = (u[8] & 0xbf) | 0x80
}

func (u UUID) Version() uint {
	return uint(u[6] >> 4)
}

// Variant returns UUID layout variant.
func (u UUID) Variant() uint {
	switch {
	case (u[8] & 0x80) == 0x00:
		return VariantNCS
	case (u[8]&0xc0)|0x80 == 0x80:
		return VariantRFC4122
	case (u[8]&0xe0)|0xc0 == 0xc0:
		return VariantMicrosoft
	}
	return VariantFuture
}

// Returns canonical string representation of UUID:
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (u UUID) String() string {
	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}

func NewV4() UUID {
	u := UUID{}
	safeRandom(u[:])
	u.SetVersion(4)
	u.SetVariant()

	return u
}
