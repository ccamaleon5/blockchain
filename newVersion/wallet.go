/*
* Adrian Pareja
 */
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"

	"crypto/rand"
	"encoding/hex"
	
	"reflect"
	"runtime"
	"sort"
)

const (
	tableColumn     = "Movimientos"
	columnWallet    = "Wallet"
	columnTime      = "Time"
	columnAccountID = "Account"
	columnBusiness  = "Business"
	columnAmount    = "Amount"
	columnBalance   = "Balance"
	columnType      = "Type"
)

const (
	tableWalletColumn     = "Wallet"
	columnWalletId  = "Id"
	columnWalletAccountID = "Account"
	columnWalletBalance   = "Balance"
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
	Limit    float64 `json:"limit"`
}

//Wallet - Structure for products used in buy goods
type WalletBalance struct {
	WalletId  string  `json:"walletid"`
	Balance   float64 `json:"balance"`
}

//Movimiento - Structure for movements
type Movement struct {
	Time     int64   `json:"time"`
	WalletId string  `json:"walletid"`
	Business string  `json:"business"`
	Amount   float64 `json:"amount"`
	Balance  float64 `json:"balance"`
	Type     string  `json:"type"`
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
			&shim.ColumnDefinition{Name: columnWallet, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnTime, Type: shim.ColumnDefinition_INT64, Key: true},
		&shim.ColumnDefinition{Name: columnBusiness, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnAmount, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnBalance, Type: shim.ColumnDefinition_STRING, Key: false},
		&shim.ColumnDefinition{Name: columnType, Type: shim.ColumnDefinition_STRING, Key: false},
	})
	
	stub.CreateTable(tableWalletColumn, []*shim.ColumnDefinition{
		&shim.ColumnDefinition{Name: columnWalletId, Type: shim.ColumnDefinition_STRING, Key: true},	
		&shim.ColumnDefinition{Name: columnAccountID, Type: shim.ColumnDefinition_STRING, Key: true},
		&shim.ColumnDefinition{Name: columnBalance, Type: shim.ColumnDefinition_STRING, Key: false},
	})
	
	fmt.Printf("Iniciandooo Job de reinicio de limite")
	
	s := NewScheduler()
	s.Every(60).Seconds().Do(resetLimit(stub))
	<- s.Start()

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
			} else {
				if function == "debitbalance" {
					return t.debitBalance(stub, args)
				} else {
					if function == "puttotalcoin" {
						return t.putTotalCoin(stub, args)
					} else if function == "debittotalcoin" {
						return t.debitTotalCoin(stub, args)
					}
				}
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
	} else {
		if function == "gettotalcoin" {
			return t.getTotalCoin(stub, args)
		} else {
			if function == "getmovimientos" {
				return t.getMovimientos(stub, args)
			} else { 
				if function == "getdatos" {
					return t.getDatos(stub, args)
				} else if function == "getwallets" {
					return t.getWallets(stub, args)
				} 
			}
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
	
	bytesWallet, _ := stub.GetState(args[0])
	if bytesWallet != nil {
		fmt.Println("Ya existe el wallet con id %s",args[0])
		return nil,errors.New("El wallet ya existe")
	}

	walletId := NewV4()
	fmt.Printf("UUIDv4: %s\n", walletId)
	amt, err := strconv.ParseFloat("0", 64)

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
		Limit:    100,
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

	a := makeTimestamp()

	fmt.Printf("Time: %d \n", a)

	col1Val := args[0]
	col2Val := "Create"
	col3Val := "0"
	col5Val := "W"

	var columns []*shim.Column
	col6 := shim.Column{Value: &shim.Column_String_{String_: "Movement"}}
	col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
	col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col4 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
	columns = append(columns, &col6)
	columns = append(columns, &col0)
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
	
	//Se inserta el Wallet en la Tabla de Wallets
	fmt.Printf("Insertando Wallet en la Tabla")

	col11Val := args[0]

	var columns1 []*shim.Column
	col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	col11 := shim.Column{Value: &shim.Column_String_{String_: col11Val}}
	col12 := shim.Column{Value: &shim.Column_String_{String_: "0.00"}}
	
	columns1 = append(columns1, &col10)
	columns1 = append(columns1, &col11)
	columns1 = append(columns1, &col12)

	row1 := shim.Row{Columns: columns1}
	ok1, err1 := stub.InsertRow("Wallet", row1)
	if err1 != nil {
		return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err1)
	}
	if !ok1 {
		return nil, errors.New("Fallo insertar Row with given key already exists")
	}
	
	//Se envia un evento de exito
	err = stub.SetEvent("createWallet", []byte("createWallet:OK"))
	if err != nil {
		return nil, errors.New("Fallo enviar el evento de Crear Wallet")
	}

	return []byte(`{"code":0,"response":null}`), nil
}

// putBalance - invocar esta funcion incrementar los coins en el balance
func (t *SimpleChaincode) putBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion PutBalance---")
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 3 para putBalance")
	}

	fmt.Printf("WalletId 1: %s\n", args[0])
	fmt.Printf("Business: %s\n", args[1])
	fmt.Printf("Monto: %s\n", args[2])

	bytesWallet1, err1 := stub.GetState(args[0])

	walletReceiver := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletReceiver)

	fmt.Println(walletReceiver)
	if err1 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}

	amt, err := strconv.ParseFloat(args[2], 64)

	walletReceiver.Amount = walletReceiver.Amount + amt //carga coins al balance

	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet

	if err != nil {
		return nil, err
	}

	a := makeTimestamp()

	fmt.Printf("Time: %d \n", a)

	col1Val := args[0]
	col2Val := args[1]
	col3Val := strconv.FormatFloat(amt, 'f', 6, 64)
	col4Val := strconv.FormatFloat(walletReceiver.Amount, 'f', 6, 64)
	col5Val := "C"

	var columns []*shim.Column
	col6 := shim.Column{Value: &shim.Column_String_{String_: "Movement"}}
	col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
	col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
	columns = append(columns, &col6)
	columns = append(columns, &col0)
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
	
	//Se actualiza el row de Wallet
	var columns1 []*shim.Column
	col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	col11 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col12 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	
	columns1 = append(columns1, &col10)
	columns1 = append(columns1, &col11)
	columns1 = append(columns1, &col12)

	row1 := shim.Row{Columns: columns1}
	ok2, err2 := stub.ReplaceRow("Wallet",row1)
	if err2 != nil {
		return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err2)
	}
	if !ok2 {
		return nil, errors.New("Fallo insertar Row Wallet with given key already exists")
	}
	
	//Se envia un evento de exito
	err = stub.SetEvent("debitEvent", []byte("debitEvent:4"))
	if err != nil {
		return nil, errors.New("Fallo enviar el evento de debito")
	}

	return []byte(`{"code":0,"response":null}`), nil
}

// debitBalance - invocar esta funcion debitar coins del balance
func (t *SimpleChaincode) debitBalance(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call---Funcion DebitBalance---")
	if len(args) != 3 {
		return nil, errors.New("Numero incorrecto de argumentos.Se espera 2 para debitBalance")
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

	amt, err := strconv.ParseFloat(args[2], 64)

	walletReceiver.Amount = walletReceiver.Amount - amt //debita coins del balance
	walletReceiver.Limit = walletReceiver.Limit - amt

	walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
	err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet

	if err != nil {
		return nil, err
	}

	a := makeTimestamp()

	fmt.Printf("Time: %d \n", a)

	col1Val := args[0]
	col2Val := args[1]
	col3Val := strconv.FormatFloat(amt, 'f', 6, 64)
	col4Val := strconv.FormatFloat(walletReceiver.Amount, 'f', 6, 64)
	col5Val := "D"

	var columns []*shim.Column
	col6 := shim.Column{Value: &shim.Column_String_{String_: "Movement"}}
	col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
	col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
	col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
	col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
	columns = append(columns, &col6)
	columns = append(columns, &col0)
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
	
	//Se actualiza el row de Wallet
	var columns1 []*shim.Column
	col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	col11 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
	col12 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	
	columns1 = append(columns1, &col10)
	columns1 = append(columns1, &col11)
	columns1 = append(columns1, &col12)

	row1 := shim.Row{Columns: columns1}
	ok3, err3 := stub.ReplaceRow("Wallet",row1)
	if err3 != nil {
		return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err3)
	}
	if !ok3 {
		return nil, errors.New("Fallo insertar Row Wallet with given key already exists")
	}
	
	//Se actualiza el balance global de coin
	
	coinBalance, err2 := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err2 != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}
	
	newCoinBalance,_ := strconv.ParseFloat(string(coinBalance), 64)
	newCoinBalance = newCoinBalance + amt
	
	err = stub.PutState("coinBalance", []byte(strconv.FormatFloat(newCoinBalance, 'f', 6, 64)))
	if err != nil {
		fmt.Println("Error setting new coinBalance")
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

	fmt.Printf("WalletId 1: %s\n", args[1])
	fmt.Printf("WalletId 2: %s\n", args[0])
	fmt.Printf("Monto: %s\n", args[2])

	bytesWallet1, err1 := stub.GetState(args[1])

	walletSender := Wallet{}
	err := json.Unmarshal(bytesWallet1, &walletSender)

	if err1 != nil {
		fmt.Println("Error retrieving " + args[1])
		return nil, errors.New("Error retrieving " + args[1])
	}
	fmt.Println(walletSender)

	bytesWallet2, err2 := stub.GetState(args[0])
	walletReceiver := Wallet{}
	err = json.Unmarshal(bytesWallet2, &walletReceiver)
	if err2 != nil {
		fmt.Println("Error retrieving " + args[0])
		return nil, errors.New("Error retrieving " + args[0])
	}
	fmt.Println(walletReceiver)

	amt, err := strconv.ParseFloat(args[2], 64)

	if amt <= walletSender.Amount {
		walletSender.Amount = walletSender.Amount - amt     //debita el monto
		walletReceiver.Amount = walletReceiver.Amount + amt //carga el monto

		walletSenderJSONasBytes, _ := json.Marshal(walletSender)
		err = stub.PutState(args[1], walletSenderJSONasBytes) //rewrite the wallet

		if err != nil {
			fmt.Println("Error guardar el Sender")
			return nil, err
		}

		walletReceiverJSONasBytes, _ := json.Marshal(walletReceiver)
		err = stub.PutState(args[0], walletReceiverJSONasBytes) //rewrite the wallet
		if err != nil {
			fmt.Println("Error guardar el Recceiver")
			return nil, err
		}

		a := makeTimestamp()

		fmt.Printf("Time: %d \n", a)

		col1Val := args[0]
		col2Val := args[1]
		col3Val := strconv.FormatFloat(amt, 'f', 6, 64)
		col4Val := strconv.FormatFloat(walletReceiver.Amount, 'f', 6, 64)
		col5Val := "C"

		var columns []*shim.Column
		var columns2 []*shim.Column
		col6 := shim.Column{Value: &shim.Column_String_{String_: "Movement"}}
		col0 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col1 := shim.Column{Value: &shim.Column_Int64{Int64: a}}
		col2 := shim.Column{Value: &shim.Column_String_{String_: col2Val}}
		col3 := shim.Column{Value: &shim.Column_String_{String_: col3Val}}
		col4 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
		col5 := shim.Column{Value: &shim.Column_String_{String_: col5Val}}
		columns = append(columns, &col6)
		columns = append(columns, &col0)
		columns = append(columns, &col1)
		columns = append(columns, &col2)
		columns = append(columns, &col3)
		columns = append(columns, &col4)
		columns = append(columns, &col5)
		
		row := shim.Row{Columns: columns}
		ok, err := stub.InsertRow("Movimientos", row)
		if err != nil {
			fmt.Println("Error al insertar la fila de sender")
			return nil, fmt.Errorf("Insert Row Movimientos operation failed. %s", err)
		}
		if !ok {
			return nil, errors.New("Fallo insertar Row with given key already exists")
		}

		fmt.Println("Inserto Fila de Sender")
		
		//Se actualiza el row de Wallet Sender
		var columns1 []*shim.Column
		col10 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
		col11 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col12 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	
		columns1 = append(columns1, &col10)
		columns1 = append(columns1, &col11)
		columns1 = append(columns1, &col12)
	
		row1 := shim.Row{Columns: columns1}
		ok3, err3 := stub.ReplaceRow("Wallet",row1)
		if err3 != nil {
			return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err3)
		}
		if !ok3 {
			return nil, errors.New("Fallo insertar Row Wallet with given key already exists")
		}

		//Se inserta fila de Receiver
		b := a+1
		
		fmt.Printf("Time: %d \n", b)
		
		col1Val = args[1]
		col2Val = args[0]
		col3Val = strconv.FormatFloat(amt, 'f', 6, 64)
		col4Val = strconv.FormatFloat(walletSender.Amount, 'f', 6, 64)
		col5Val = "D"

		col0 = shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col1 = shim.Column{Value: &shim.Column_Int64{Int64: b}}
		col2 = shim.Column{Value: &shim.Column_String_{String_: col2Val}}
		col3 = shim.Column{Value: &shim.Column_String_{String_: col3Val}}
		col4 = shim.Column{Value: &shim.Column_String_{String_: col4Val}}
		col5 = shim.Column{Value: &shim.Column_String_{String_: col5Val}}
		columns2 = append(columns2, &col6)
		columns2 = append(columns2, &col0)
		columns2 = append(columns2, &col1)
		columns2 = append(columns2, &col2)
		columns2 = append(columns2, &col3)
		columns2 = append(columns2, &col4)
		columns2 = append(columns2, &col5)

		row2 := shim.Row{Columns: columns2}
		ok2, err2 := stub.InsertRow("Movimientos", row2)
		if err2 != nil {
			fmt.Println("Error al insertar la fila de receiver")
			return nil, fmt.Errorf("Insert Row2 Movimientos operation failed. %s", err2)
		}
		if !ok2 {
			return nil, errors.New("Fallo insertar Row2 with given key already exists")
		}

		fmt.Println("Inserto fila de receiver")
		
		//Se actualiza el row de Wallet Receiver
		var columns3 []*shim.Column
		col20 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
		col21 := shim.Column{Value: &shim.Column_String_{String_: col1Val}}
		col22 := shim.Column{Value: &shim.Column_String_{String_: col4Val}}
	
		columns3 = append(columns3, &col20)
		columns3 = append(columns3, &col21)
		columns3 = append(columns3, &col22)
	
		row3 := shim.Row{Columns: columns3}
		ok4, err4 := stub.ReplaceRow("Wallet",row3)
		if err4 != nil {
			return nil, fmt.Errorf("Insert Row Wallet operation failed. %s", err4)
		}
		if !ok4 {
			return nil, errors.New("Fallo insertar Row Wallet with given key already exists")
		}

		return []byte(`{"code":0,"response":null}`), nil
	} else {
		return nil, errors.New("No cuentas con suficientes coins para esta transferencia")
	}
}

//Obtener el balance de un wallet
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

	return []byte(fmt.Sprintf(`{"code":0,"balance":"%s","limit":"%s"}`, strconv.FormatFloat(wallet.Amount, 'f', 6, 64), strconv.FormatFloat(wallet.Limit, 'f', 6, 64))), nil
}

//Funcion que obtiene el total de Coins en el sistema
func (t *SimpleChaincode) getTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getTotalCoin() is running----")

	coinBalance, err := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, coinBalance)), nil
}

//Funcion que otorga coins a los negocios
func (t *SimpleChaincode) debitTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----debitTotalCoin() is running----")
	
	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	coinBalance, err := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}
	
	newCoinBalance,_ := strconv.ParseFloat(string(coinBalance), 64)
	amount,_ := strconv.ParseFloat(args[0], 64)
	newCoinBalance = newCoinBalance - amount
	
	err = stub.PutState("coinBalance", []byte(strconv.FormatFloat(newCoinBalance, 'f', 6, 64)))
	if err != nil {
		fmt.Println("Error setting new coinBalance")
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, coinBalance)), nil
}

//Funcion que devuelve coins a los negocios
func (t *SimpleChaincode) putTotalCoin(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----putTotalCoin() is running----")
	
	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	coinBalance, err := stub.GetState("coinBalance")
	fmt.Println(coinBalance)
	if err != nil {
		fmt.Println("Error retrieving coinBalance")
		return nil, errors.New("Error retrieving coinBalance")
	}
	
	newCoinBalance,_ := strconv.ParseFloat(string(coinBalance), 64)
	amount,_ := strconv.ParseFloat(args[0], 64)
	
	newCoinBalance = newCoinBalance + amount
	
	err = stub.PutState("coinBalance", []byte(strconv.FormatFloat(newCoinBalance, 'f', 6, 64)))
	if err != nil {
		fmt.Println("Error setting new coinBalance")
		return nil, err
	}

	return []byte(fmt.Sprintf(`{"code":0,"response":"%s"}`, coinBalance)), nil
}

//Funcion que obtiene todos los movimientos de un usuario
func (t *SimpleChaincode) getMovimientos(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getMovimientos() is running----")

	if len(args) < 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	var columns []shim.Column
	col0 := shim.Column{Value: &shim.Column_String_{String_: "Movement"}}
	columns = append(columns, col0)
	if len(args) == 2 {
		walletId := args[1] // wallet id
		fmt.Println("wallet id is ")
		fmt.Println(walletId)
		col1 := shim.Column{Value: &shim.Column_String_{String_: walletId}}
		columns = append(columns, col1)
	}
	
	rowChannel, err := stub.GetRows("Movimientos", columns)
	if err != nil {
		return nil, fmt.Errorf("getRowTableOne operation failed. %s", err)
	}

	movimientos := []Movement{}
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				columnas := row.GetColumns()
				amountRow, _ := strconv.ParseFloat(columnas[4].GetString_(), 64)
				balanceRow, _ := strconv.ParseFloat(columnas[5].GetString_(), 64)
				movimiento := Movement{Time: columnas[2].GetInt64(), WalletId: columnas[1].GetString_(), Business: columnas[3].GetString_(), Amount: amountRow, Balance: balanceRow, Type: columnas[6].GetString_()}

				movimientos = append(movimientos, movimiento)
			}
		}
		if rowChannel == nil {
			break
		}
	}

	jsonRows, err := json.Marshal(movimientos)
	if err != nil {
		return nil, fmt.Errorf("getRows Movimientos operation failed. Error marshaling JSON: %s", err)
	}

	return jsonRows, nil
}

//Funcion que obtiene todos los wallets ubicados en la blockchain
func (t *SimpleChaincode) getWallets(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getWallets() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: "Wallet"}}
	columns = append(columns, col1)

	rowChannel, err := stub.GetRows("Wallet", columns)
	if err != nil {
		return nil, fmt.Errorf("getRow TableWallet operation failed. %s", err)
	}

	wallets := []WalletBalance{}
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				columnas := row.GetColumns()
				balanceRow, _ := strconv.ParseFloat(columnas[2].GetString_(), 64)
				wallet := WalletBalance{WalletId: columnas[1].GetString_(), Balance: balanceRow}

				wallets = append(wallets, wallet)
			}
		}
		if rowChannel == nil {
			break
		}
	}

	jsonRows, err := json.Marshal(wallets)
	if err != nil {
		return nil, fmt.Errorf("getRows Wallet operation failed. Error marshaling JSON: %s", err)
	}

	return jsonRows, nil
}

//getData - Obtiene los datos generales del usuario
func (t *SimpleChaincode) getDatos(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Call----getDatos() is running----")

	if len(args) != 1 {
		return nil, errors.New("Incorrecto numero de argumentos. Se esperaba 1")
	}

	walletId := args[0] // wallet id
	fmt.Println("wallet id is")
	fmt.Println(walletId)
	bytes, err := stub.GetState(args[0])
	if err != nil {
		fmt.Println("Error retrieving " + walletId)
		return nil, errors.New("Error retrieving " + walletId)
	}
	wallet := Wallet{}
	err1 := json.Unmarshal(bytes, &wallet)

	fmt.Println(wallet)
	if err1 != nil {
		fmt.Println("Error parseando a Json" + args[0])
		return nil, errors.New("Error retrieving Balance" + args[0])
	}

	return bytes, nil
}

//resetea los limites de los wallets
func resetLimit(stub shim.ChaincodeStubInterface) bool{
	fmt.Println("--------LLamando JOBBBB----------")
	//var params = []string{"Wallet"}
	bytes, err := stub.GetState("42928586")
	if err != nil {
		fmt.Println("Error retrieving")
	}
	wallet := Wallet{}
	err1 := json.Unmarshal(bytes, &wallet)
	if err1 != nil {
		fmt.Println("Error retrieving")
	}
	fmt.Println("Retorno: %s",wallet.Id)
	
	return true
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

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}





// Time location, default set by the time.Local (*time.Location)
var loc = time.Local

// Change the time location
func ChangeLoc(newLocation *time.Location) {
	loc = newLocation
}

// Max number of jobs, hack it if you need.
const MAXJOBNUM = 10000

type Job struct {

	// pause interval * unit bettween runs
	interval uint64

	// the job jobFunc to run, func[jobFunc]
	jobFunc string
	// time units, ,e.g. 'minutes', 'hours'...
	unit string
	// optional time at which this job runs
	atTime string

	// datetime of last run
	lastRun time.Time
	// datetime of next run
	nextRun time.Time
	// cache the period between last an next run
	period time.Duration

	// Specific day of the week to start on
	startDay time.Weekday

	// Map for the function task store
	funcs map[string]interface{}

	// Map for function and  params of function
	fparams map[string]([]interface{})
}

// Create a new job with the time interval.
func NewJob(intervel uint64) *Job {
	return &Job{
		intervel,
		"", "", "",
		time.Unix(0, 0),
		time.Unix(0, 0), 0,
		time.Sunday,
		make(map[string]interface{}),
		make(map[string]([]interface{})),
	}
}

// True if the job should be run now
func (j *Job) shouldRun() bool {
	return time.Now().After(j.nextRun)
}

//Run the job and immdiately reschedulei it
func (j *Job) run() (result []reflect.Value, err error) {
	f := reflect.ValueOf(j.funcs[j.jobFunc])
	params := j.fparams[j.jobFunc]
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of param is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	j.lastRun = time.Now()
	j.scheduleNextRun()
	return
}

// for given function fn , get the name of funciton.
func getFunctionName(fn interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf((fn)).Pointer()).Name()
}

// Specifies the jobFunc that should be called every time the job runs
//
func (j *Job) Do(jobFun interface{}, params ...interface{}) {
	typ := reflect.TypeOf(jobFun)
	if typ.Kind() != reflect.Func {
		panic("only function can be schedule into the job queue.")
	}

	fname := getFunctionName(jobFun)
	j.funcs[fname] = jobFun
	j.fparams[fname] = params
	j.jobFunc = fname
	//schedule the next run
	j.scheduleNextRun()
}

//	s.Every(1).Day().At("10:30").Do(task)
//	s.Every(1).Monday().At("10:30").Do(task)
func (j *Job) At(t string) *Job {
	hour := int((t[0]-'0')*10 + (t[1] - '0'))
	min := int((t[3]-'0')*10 + (t[4] - '0'))
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		panic("time format error.")
	}
	// time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	mock := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), int(hour), int(min), 0, 0, loc)

	if j.unit == "days" {
		if time.Now().After(mock) {
			j.lastRun = mock
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, hour, min, 0, 0, loc)
		}
	} else if j.unit == "weeks" {
		if time.Now().After(mock) {
			i := mock.Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), hour, min, 0, 0, loc)
		} else {
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-7, hour, min, 0, 0, loc)
		}
	}
	return j
}

//Compute the instant when this job should run next
func (j *Job) scheduleNextRun() {
	if j.lastRun == time.Unix(0, 0) {
		if j.unit == "weeks" {
			i := time.Now().Weekday() - j.startDay
			if i < 0 {
				i = 7 + i
			}
			j.lastRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-int(i), 0, 0, 0, 0, loc)

		} else {
			j.lastRun = time.Now()
		}
	}

	if j.period != 0 {
		// translate all the units to the Seconds
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	} else {
		switch j.unit {
		case "minutes":
			j.period = time.Duration(j.interval * 60)
			break
		case "hours":
			j.period = time.Duration(j.interval * 60 * 60)
			break
		case "days":
			j.period = time.Duration(j.interval * 60 * 60 * 24)
			break
		case "weeks":
			j.period = time.Duration(j.interval * 60 * 60 * 24 * 7)
			break
		case "seconds":
			j.period = time.Duration(j.interval)
		}
		j.nextRun = j.lastRun.Add(j.period * time.Second)
	}
}

// the follow functions set the job's unit with seconds,minutes,hours...

// Set the unit with second
func (j *Job) Second() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Seconds()
	return
}

// Set the unit with seconds
func (j *Job) Seconds() (job *Job) {
	j.unit = "seconds"
	return j
}

// Set the unit  with minute, which interval is 1
func (j *Job) Minute() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Minutes()
	return
}

//set the unit with minute
func (j *Job) Minutes() (job *Job) {
	j.unit = "minutes"
	return j
}

//set the unit with hour, which interval is 1
func (j *Job) Hour() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Hours()
	return
}

// Set the unit with hours
func (j *Job) Hours() (job *Job) {
	j.unit = "hours"
	return j
}

// Set the job's unit with day, which interval is 1
func (j *Job) Day() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Days()
	return
}

// Set the job's unit with days
func (j *Job) Days() *Job {
	j.unit = "days"
	return j
}

/*
// Set the unit with week, which the interval is 1
func (j *Job) Week() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	job = j.Weeks()
	return
}

*/

// s.Every(1).Monday().Do(task)
// Set the start day with Monday
func (j *Job) Monday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 1
	job = j.Weeks()
	return
}

// Set the start day with Tuesday
func (j *Job) Tuesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 2
	job = j.Weeks()
	return
}

// Set the start day woth Wednesday
func (j *Job) Wednesday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 3
	job = j.Weeks()
	return
}

// Set the start day with thursday
func (j *Job) Thursday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 4
	job = j.Weeks()
	return
}

// Set the start day with friday
func (j *Job) Friday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 5
	job = j.Weeks()
	return
}

// Set the start day with saturday
func (j *Job) Saturday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 6
	job = j.Weeks()
	return
}

// Set the start day with sunday
func (j *Job) Sunday() (job *Job) {
	if j.interval != 1 {
		panic("")
	}
	j.startDay = 0
	job = j.Weeks()
	return
}

//Set the units as weeks
func (j *Job) Weeks() *Job {
	j.unit = "weeks"
	return j
}

// Class Scheduler, the only data member is the list of jobs.
type Scheduler struct {
	// Array store jobs
	jobs [MAXJOBNUM]*Job

	// Size of jobs which jobs holding.
	size int
}

// Scheduler implements the sort.Interface{} for sorting jobs, by the time nextRun

func (s *Scheduler) Len() int {
	return s.size
}

func (s *Scheduler) Swap(i, j int) {
	s.jobs[i], s.jobs[j] = s.jobs[j], s.jobs[i]
}

func (s *Scheduler) Less(i, j int) bool {
	return s.jobs[j].nextRun.After(s.jobs[i].nextRun)
}

// Create a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{[MAXJOBNUM]*Job{}, 0}
}

// Get the current runnable jobs, which shouldRun is True
func (s *Scheduler) getRunnableJobs() (running_jobs [MAXJOBNUM]*Job, n int) {
	runnableJobs := [MAXJOBNUM]*Job{}
	n = 0
	sort.Sort(s)
	for i := 0; i < s.size; i++ {
		if s.jobs[i].shouldRun() {

			runnableJobs[n] = s.jobs[i]
			//fmt.Println(runnableJobs)
			n++
		} else {
			break
		}
	}
	return runnableJobs, n
}

// Datetime when the next job should run.
func (s *Scheduler) NextRun() (*Job, time.Time) {
	if s.size <= 0 {
		return nil, time.Now()
	}
	sort.Sort(s)
	return s.jobs[0], s.jobs[0].nextRun
}

// Schedule a new periodic job
func (s *Scheduler) Every(interval uint64) *Job {
	job := NewJob(interval)
	s.jobs[s.size] = job
	s.size++
	return job
}

// Run all the jobs that are scheduled to run.
func (s *Scheduler) RunPending() {
	runnableJobs, n := s.getRunnableJobs()

	if n != 0 {
		for i := 0; i < n; i++ {
			runnableJobs[i].run()
		}
	}
}

// Run all jobs regardless if they are scheduled to run or not
func (s *Scheduler) RunAll() {
	for i := 0; i < s.size; i++ {
		s.jobs[i].run()
	}
}

// Run all jobs with delay seconds
func (s *Scheduler) RunAllwithDelay(d int) {
	for i := 0; i < s.size; i++ {
		s.jobs[i].run()
		time.Sleep(time.Duration(d))
	}
}

// Remove specific job j
func (s *Scheduler) Remove(j interface{}) {
	i := 0
	for ; i < s.size; i++ {
		if s.jobs[i].jobFunc == getFunctionName(j) {
			break
		}
	}

	for j := (i + 1); j < s.size; j++ {
		s.jobs[i] = s.jobs[j]
		i++
	}
	s.size = s.size - 1
}

// Delete all scheduled jobs
func (s *Scheduler) Clear() {
	for i := 0; i < s.size; i++ {
		s.jobs[i] = nil
	}
	s.size = 0
}

// Start all the pending jobs
// Add seconds ticker
func (s *Scheduler) Start() chan bool {
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				s.RunPending()
			case <-stopped:
				return
			}
		}
	}()

	return stopped
}

// The following methods are shortcuts for not having to
// create a Schduler instance

var defaultScheduler = NewScheduler()
var jobs = defaultScheduler.jobs

// Schedule a new periodic job
func Every(interval uint64) *Job {
	return defaultScheduler.Every(interval)
}

// Run all jobs that are scheduled to run
//
// Please note that it is *intended behavior that run_pending()
// does not run missed jobs*. For example, if you've registered a job
// that should run every minute and you only call run_pending()
// in one hour increments then your job won't be run 60 times in
// between but only once.
func RunPending() {
	defaultScheduler.RunPending()
}

// Run all jobs regardless if they are scheduled to run or not.
func RunAll() {
	defaultScheduler.RunAll()
}

// Run all the jobs with a delay in seconds
//
// A delay of `delay` seconds is added between each job. This can help
// to distribute the system load generated by the jobs more evenly over
// time.
func RunAllwithDelay(d int) {
	defaultScheduler.RunAllwithDelay(d)
}

// Run all jobs that are scheduled to run
func Start() chan bool {
	return defaultScheduler.Start()
}

// Clear
func Clear() {
	defaultScheduler.Clear()
}

// Remove
func Remove(j interface{}) {
	defaultScheduler.Remove(j)
}

// NextRun gets the next running time
func NextRun() (job *Job, time time.Time) {
	return defaultScheduler.NextRun()
}
