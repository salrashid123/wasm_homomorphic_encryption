package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/ldsec/lattigo/v2/rlwe"
)

type Operands struct {
	A string
	B string
}

var (
	params    bfv.Parameters
	evaluator bfv.Evaluator
)

func gethandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func add(w http.ResponseWriter, r *http.Request) {
	fmt.Println("add() called")
	var p Operands
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//fmt.Printf("Operands: %+v\n", p)

	eaBytes, err := base64.StdEncoding.DecodeString(p.A)
	if err != nil {
		fmt.Printf("Error decoding ea to binary %v", err)
		return
	}
	ciphertextA := bfv.NewCiphertext(params, 1)
	err = ciphertextA.UnmarshalBinary(eaBytes)
	if err != nil {
		fmt.Printf("Error converting ea to binary %v", err)
		return
	}

	ebBytes, err := base64.StdEncoding.DecodeString(p.B)
	if err != nil {
		fmt.Printf("Error decoding eb to binary %v", err)
		return
	}
	ciphertextB := bfv.NewCiphertext(params, 1)
	err = ciphertextB.UnmarshalBinary(ebBytes)
	if err != nil {
		fmt.Printf("Error converting eb to binary %v", err)
		return
	}

	aPlusBciphertext := bfv.NewCiphertext(params, 1)

	evaluator.Add(ciphertextA, ciphertextB, aPlusBciphertext)

	resultBytes, err := aPlusBciphertext.MarshalBinary()
	if err != nil {
		fmt.Printf("Error converting aPlusBciphertext to binary %v", err)
		return
	}

	//result := encoder.DecodeUintNew(decryptorSk.DecryptNew(aPlusBciphertext))

	resp := simplejson.New()

	resp.Set("result", base64.StdEncoding.EncodeToString(resultBytes))

	payload, err := resp.MarshalJSON()
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func main() {
	// BFV parameters (128 bit security) with plaintext modulus 65929217
	paramDef := bfv.PN13QP218
	paramDef.T = 0x3ee0001
	var err error
	params, err = bfv.NewParametersFromLiteral(paramDef)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	evaluator = bfv.NewEvaluator(params, rlwe.EvaluationKey{})

	router := mux.NewRouter()
	router.Methods(http.MethodGet).Path("/get").HandlerFunc(gethandler)
	router.Methods(http.MethodPost).Path("/add").HandlerFunc(add)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", router)
	fmt.Println("Starting Server")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
