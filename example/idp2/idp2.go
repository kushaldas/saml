// Package main contains an example identity provider implementation.
package main

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/zenazn/goji"
	"golang.org/x/crypto/bcrypt"

	"github.com/crewjam/saml/logger"
	"github.com/crewjam/saml/samlidp"
)

var key = func() crypto.PrivateKey {
	data, err := os.ReadFile("./idp.key")
	if err != nil {
		panic(err)
	}
	b, _ := pem.Decode(data)
	k, err := x509.ParsePKCS8PrivateKey(b.Bytes)

	if err != nil {
		panic(err)
	}
	return k
}()

var cert = func() *x509.Certificate {
	data, err := os.ReadFile("./idp.pem")
	if err != nil {
		panic(err)
	}
	b, _ := pem.Decode(data)
	c, _ := x509.ParseCertificate(b.Bytes)
	return c
}()

//var localca = func() *x509.Certificate {
//data, err := os.ReadFile("./rootCA.pem")
//if err != nil {
//panic(err)
//}
//b, _ := pem.Decode(data)
//c, _ := x509.ParseCertificate(b.Bytes)
//return c
//}()

func main() {

	logr := logger.DefaultLogger
	baseURLstr := flag.String("idp", "", "The URL to the IDP")
	flag.Parse()

	baseURL, err := url.Parse(*baseURLstr)
	if err != nil {
		logr.Fatalf("cannot parse base URL: %v", err)
	}

	idpServer, err := samlidp.New(samlidp.Options{
		URL:         *baseURL,
		Key:         key,
		Logger:      logr,
		Certificate: cert,
		Store:       &samlidp.MemoryStore{},
		//Intermediates: []*x509.Certificate{localca},
	})
	if err != nil {
		logr.Fatalf("%s", err)
	}

	userdata, err := os.ReadFile("./users.json")

	if err != nil {
		log.Println("Error when opening users file: ", err)
		log.Println("Now trying to read /etc/idp2/users.json")
		userdata, err = os.ReadFile("/etc/idp2/users.json")
		if err != nil {
			logr.Fatalf("%s", err)
		}

	}

	var users []samlidp.User
	err = json.Unmarshal(userdata, &users)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("hunter2"), bcrypt.DefaultCost)
	for _, user := range users {
		user.HashedPassword = hashedPassword
		err = idpServer.Store.Put(fmt.Sprintf("/users/%s", user.Name), user)
		if err != nil {
			logr.Fatalf("%s", err)
		}
	}

	goji.Handle("/*", idpServer)
	goji.Serve()
}
