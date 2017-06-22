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

const (
	tableColumn       = "Movimientos"
	columnAccountID   = "Account"
	columnBusiness   = "Business"
	columnAmount = "Amount"
	columnBalance = "Balance"
	columnType = "Type"
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
	Id       string  `json:"id"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Document string  `json:"document"`
	Password string  `json:"password"`
	Amount   float64 `json:"amount"`
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
	
	stub.CreateTable(tableColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnBusiness, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnBalance, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnType, Type: shim.ColumnDefinition_STRING, Key: false},
	})

	return nil, nil
}

// Invoke Punto de entrada a cualquier función del ledger
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running..FUNCTION:" + function)

	if function == "createwallet" {
		return t.createWallet(stub, args)
	} else {
		if function == "transfer" {
			return t.transfer(stub, args)
		} else {
			if function == "putbalance" {
				return t.putBalance(stub, args)
			} else if function == "debitbalance" {
				return t.debitBalance(stub, args)
			}
		}
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
	} else{ 
		if function == "gettotalcoin" {
			return t.getTotalCoin(stub, args)
		} else if function == "getmovimientos"{
			return t.getMovimientos(stub, args)
		}
	}
	fmt.Println("query no encuentra la funcion: " + function)

	return nil, errors.New("Funcion invocada desconocida: " + function)
}

// createWallet - invocar esta funcion para crear un wallet con saldo inicial
func (t *SimpleChaincode) createWallet(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion createWallet---")
	if len(args) != 6 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 6 para createWallet")
	}

	walletId := NewV4()
	fmt.Printf("UUIDv4: %s\n", walletId)
	amt, err := strconv.ParseFloat(args[5], 64)

	if err != nil {
		fmt.Println("Error Float parsing")
		return nil, errors.New("Error marshaling wallet")
	}

	wallet := Wallet{
		Id:       args[0],
		Email:    args[1],
		Phone:    args[2],
		Document: args[3],
		Password: args[4],
		Amount:   amt,
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
	
	col1Val := args[0]
	col2Val := args[1]
	col3Val := args[2]
	col4Val := args[3]
	col5Val := args[4]
	
	var columns []*shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
		col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
		col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
		col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
		columns = append(columns, &col1)
		columns = append(columns, &col2)
		columns = append(columns, &col3)
		columns = append(columns, &col4)
		columns = append(columns, &col5)

		row := shim.Row{Columns: columns}
		ok, err := stub.InsertRow("Movimientos", row)
		if err != nil {
			return nil, fmt.Errorf("Insert Row Movimientos operation failed. %s", err)
		}
		if !ok {
			return nil, errors.New("Fallo insertar Row with given key already exists")
		}
	
	return []byte(`{"code":0,"response":null}`), nil
}

// putBalance - invocar esta funcion incrementar los coins en el balance
func (t *SimpleChaincode) putBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion PutBalance---")
	if len(args) != 2 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 2 para createWallet")
	}

	fmt.Printf("WalletId 1: %s\n", args[0])
	fmt.Printf("Monto: %s\n", args[1])

	bytesWallet1, err1 := stub.GetState(args[0])

	walletReceiver := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletReceiver)

	fmt.Println(walletReceiver)
	if err1 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}

	amt, err := strconv.ParseFloat(args[1], 64)

	walletReceiver.Amount = walletReceiver.Amount + amt //carga coins al balance

	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet

	if err != nil {
		return nil, err
	}

	return []byte(`{"code":0,"response":null}`), nil
}

// debitBalance - invocar esta funcion debitar coins del balance
func (t *SimpleChaincode) debitBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion DebitBalance---")
	if len(args) != 2 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 2 para createWallet")
	}

	fmt.Printf("WalletId 1: %s\n", args[0])
	fmt.Printf("Monto: %s\n", args[1])

	bytesWallet1, err1 := stub.GetState(args[0])

	walletReceiver := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletReceiver)

	fmt.Println(walletReceiver)
	if err1 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}

	amt, err := strconv.ParseFloat(args[1], 64)

	walletReceiver.Amount = walletReceiver.Amount - amt //carga coins al balance

	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet

	if err != nil {
		return nil, err
	}

	return []byte(`{"code":0,"response":null}`), nil
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

	if amt <= walletSender.Amount {

		walletSender.Amount = walletSender.Amount - amt     //debita el monto
		walletReceiver.Amount = walletReceiver.Amount + amt //carga el monto

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

		return []byte(`{"code":0,"response":null}`), nil
	} else {
		return nil,errors.New("No cuentas con suficientes coins para esta transferencia")
	}
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

	fmt.Println(wallet.Amount)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, strconv.FormatFloat(wallet.Amount, 'f', 6, 64))), nil
}

func (t *SimpleChaincode) getTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getTotalCoin() is running----")

	bytes, err := stub.GetState("coinBalance")
	fmt.Println(bytes)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, bytes)), nil
}

func (t *SimpleChaincode) getMovimientos(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getMovimientos() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("wallet id is ")
	fmt.Println(walletId)
		var columns []shim.Column
		col1 := shim.Column{Value: &shim.Column_String_{String_: walletId}}
		columns = append(columns, col1)

		rowChannel, err := stub.GetRows("Movimientos", columns)
		if err != nil {
			return nil, fmt.Errorf("getRowTableOne operation failed. %s", err)
		}

		var rows []shim.Row
		for {
			select {
			case row, ok := <-rowChannel:
				if !ok {
					rowChannel = nil
				} else {
					rows = append(rows, row)
				}
			}
			if rowChannel == nil {
				break
			}
		}

		jsonRows, err := json.Marshal(rows)
		if err != nil {
			return nil, fmt.Errorf("getRowsTableTwo operation failed. Error marshaling JSON: %s", err)
		}

		return jsonRows, nil
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
