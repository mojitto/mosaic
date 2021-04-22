package main

import(
	"os"
	"mosaic/abe"
	"bufio"
	"encoding/json"
	"fmt"
	"crypto/aes"
    "crypto/sha256"
    "crypto/cipher"
	"flag"
//    "crypto/rand"
//	"encoding/hex"
    "errors"
    "bytes"
   // "io"
)

func main(){
	//so we could make it read username and attributes from file as parameter, so far it's hardcoded 
	//tu myślę że można dodać parametryzację przy uruchamianiu czyli -user marcello.paris@gmail.com -keyspath .../marcello.paris@gmail.com.json
	//ale to do ustalenia
	var keysPath string
    flag.StringVar(&keysPath, "keys", "", "path to keys")
	flag.Parse()
    if len(keysPath) == 0 {
        panic("-keys is required")
    }
	file, _ := os.Open("new_files/ciphertext.json")
	reader := bufio.NewReader(file)
	message_pack_str, _:=reader.ReadString('\n')
	message_pack := new(abe.Enc_and_Key)
	json.Unmarshal([]byte(message_pack_str), message_pack)

	ciphertextStr:=message_pack.Enc_key
	//fmt.Printf("%s", ciphertextStr)
	ciphertext := new(abe.Ciphertext)
	json.Unmarshal([]byte(ciphertextStr), ciphertext)
	ciphertext.OfJsonObj()
	//so we restored ciphertext into object from json file
	//otwierzyliśmy ciphertext
	file2, _ := os.Open(keysPath)
	reader2 := bufio.NewReader(file2)
	userattrsStr, _:=reader2.ReadString('\n')
	
	userattrs := new(abe.UserAttrs)
	json.Unmarshal([]byte(userattrsStr), userattrs)
	userattrs.OfJsonObj()
	//same for keys
	//otworzyliśmy klucze dla atrybutów
	userattrs.SelectUserAttrs(userattrs.User, ciphertext.Policy)//okrajamy tylko do tych, które się przydadzą
	//we choose only those keys which are required for decryption	
	secret_dec := abe.Decrypt(ciphertext, userattrs)//deszyfracja //dec
	secret_dec_hash := abe.SecretHash(secret_dec)//hash
	
	//fmt.Printf("%s", secret_dec_hash)//wyświetlić, porównać z tym z decrypta
	//it should print same hash as encrypt has done, compare, the end
	enc_msg:=[]byte(abe.Decode(message_pack.Enc_msg))
	key := []byte(abe.Decode(secret_dec_hash))
	block, err := aes.NewCipher(key)
	if err != nil {
			panic(err)
	}
	
	iv := enc_msg[:aes.BlockSize]
	enc_msg = enc_msg[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(enc_msg, enc_msg)
	
	enc_msg, _ = pkcs7Unpad(enc_msg, 16)
	fmt.Printf("\n%s", string(enc_msg))
	hash:=sha256.Sum256([]byte(enc_msg))
	if (abe.Encode(string(hash[:]))==message_pack.Plaintext_hash){
		fmt.Printf("\nPlaintext hashes equal\n")
	}
	
	

}	



var (
	// ErrInvalidBlockSize indicates hash blocksize <= 0.
	ErrInvalidBlockSize = errors.New("invalid blocksize")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails to bad input.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func pkcs7Pad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	n := blocksize - (len(b) % blocksize)
	pb := make([]byte, len(b)+n)
	copy(pb, b)
	copy(pb[len(b):], bytes.Repeat([]byte{byte(n)}, n))
	return pb, nil
}

// pkcs7Unpad validates and unpads data from the given bytes slice.
// The returned value will be 1 to n bytes smaller depending on the
// amount of padding, where n is the block size.
func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, ErrInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, ErrInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}