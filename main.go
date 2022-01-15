package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"syscall/js"

	"github.com/ldsec/lattigo/v2/bfv"
	"github.com/ldsec/lattigo/v2/rlwe"
)

var (
	a, b        int
	r           string
	params      bfv.Parameters
	encoder     bfv.Encoder
	encryptorPk bfv.Encryptor
	decryptorSk bfv.Decryptor
	evaluator   bfv.Evaluator
)

func main() {
	fmt.Println("============================================")
	fmt.Println("init keys")
	fmt.Println("============================================")
	fmt.Println()

	// BFV parameters (128 bit security) with plaintext modulus 65929217
	paramDef := bfv.PN13QP218
	paramDef.T = 0x3ee0001
	var err error
	params, err = bfv.NewParametersFromLiteral(paramDef)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	encoder = bfv.NewEncoder(params)
	kgen := bfv.NewKeyGenerator(params)

	sk, pk := kgen.GenKeyPair()

	encryptorPk = bfv.NewEncryptor(params, pk)
	decryptorSk = bfv.NewDecryptor(params, sk)

	evaluator = bfv.NewEvaluator(params, rlwe.EvaluationKey{})

	keyParameters := fmt.Sprintf("Parameters : N=%d, T=%d, Q = %d bits, sigma = %f \n",
		1<<params.LogN(), params.T(), params.LogQP(), params.Sigma())
	fmt.Println(keyParameters)
	js.Global().Get("document").Call("getElementById", "parameters").Set("innerHTML", keyParameters)
	fmt.Println()

	document := js.Global().Get("document")
	document.Call("getElementById", "a").Set("oninput", updateEncrypter(&a))
	document.Call("getElementById", "b").Set("oninput", updateEncrypter(&b))
	js.Global().Set("decrypt", js.FuncOf(Decrypt))

	<-make(chan bool)
}

//export updater
func updateEncrypter(n *int) js.Func {
	fmt.Printf("callback for operands %d\n", *n)
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		*n, _ = strconv.Atoi(this.Get("value").String())
		Encrypt()
		return nil
	})
}

//export result
func Decrypt(this js.Value, p []js.Value) interface{} {
	fmt.Println("result callback")

	aPlusBciphertextEncoded := p[0].String()

	aPlusBciphertextEncodedBytes, err := base64.StdEncoding.DecodeString(aPlusBciphertextEncoded)
	if err != nil {
		fmt.Printf("Error converting aPlusBciphertextEncoded from binary %v", err)
		return nil
	}

	aPlusBciphertext := bfv.NewCiphertext(params, 1)
	err = aPlusBciphertext.UnmarshalBinary(aPlusBciphertextEncodedBytes)
	if err != nil {
		fmt.Printf("Error unmarshalling aPlusBciphertextEncodedBytes from binary %v", err)
		return nil
	}

	result := encoder.DecodeUintNew(decryptorSk.DecryptNew(aPlusBciphertext))
	var computedDist uint64

	for i := 0; i < len(result); i++ {
		computedDist = computedDist + result[i]
	}

	fmt.Printf("Decrypted: %d\n", computedDist)

	return js.ValueOf(computedDist)
}

func Encrypt() {

	fmt.Println("Adding numbers")

	plaintextA := bfv.NewPlaintext(params)

	r1 := make([]uint64, 1<<params.LogN())
	r1[0] = uint64(a)
	encoder.EncodeUint(r1, plaintextA)

	plaintextB := bfv.NewPlaintext(params)
	r2 := make([]uint64, 1<<params.LogN())
	r2[1] = uint64(b)
	encoder.EncodeUint(r2, plaintextB)

	ciphertextA := encryptorPk.EncryptNew(plaintextA)
	ciphertextB := encryptorPk.EncryptNew(plaintextB)

	eaBytes, err := ciphertextA.MarshalBinary()
	if err != nil {
		fmt.Printf("Error converting ea to binary %v", err)
		return
	}

	ebBytes, err := ciphertextB.MarshalBinary()
	if err != nil {
		fmt.Printf("Error converting eb to binary %v", err)
		return
	}
	eaEncoded := base64.StdEncoding.EncodeToString(eaBytes)
	ebEncoded := base64.StdEncoding.EncodeToString(ebBytes)

	eah := sha256.Sum256(eaBytes)
	ebh := sha256.Sum256(ebBytes)

	js.Global().Get("document").Call("getElementById", "ea").Set("innerHTML", fmt.Sprintf("%s", base64.StdEncoding.EncodeToString((eah[:]))))
	js.Global().Get("document").Call("getElementById", "eb").Set("innerHTML", fmt.Sprintf("%s", base64.StdEncoding.EncodeToString((ebh[:]))))
	js.Global().Call("add", eaEncoded, ebEncoded)
}
