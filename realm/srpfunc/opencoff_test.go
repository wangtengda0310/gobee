package srpfunc

import (
	"crypto/subtle"
	"fmt"
	"github.com/opencoff/go-srp"
	"math/big"
)

type server struct {
	ih, vh string
	srv    *srp.Server
	k      []byte
	s      *srp.SRP
	c      *srp.Client
}

var bits = 1024

func (srv *server) register(i, pass []byte) {

	ss, err := srp.New(bits)
	if err != nil {
		panic(err)
	}
	srv.s = ss

	//

	v, err := srv.s.Verifier(i, pass)
	if err != nil {
		panic(err)
	}

	srv.ih, srv.vh = v.Encode()

	// Store ih, vh in durable storage
	fmt.Printf("Verifier Store:\n   %s => %s\n", srv.ih, srv.vh)
}

func (srv *server) f1(A *big.Int) string {

	// Now, pretend to lookup the user db using "I" as the key and
	// fetch salt, verifier etc.
	s, v, err := srp.MakeSRPVerifier(srv.vh)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Server Begin; <v, A>:\n   %s\n   %x\n", srv.vh, A.Bytes())
	srv1, err := s.NewServer(v, A)
	if err != nil {
		panic(err)
	}
	srv.srv = srv1

	// Generate the credentials to send to client
	creds := srv.srv.Credentials()

	// Send the server public key and salt to server
	fmt.Printf("Server Begin; <s, B> --> client:\n   %s\n", creds)

	return creds
}

func (srv *server) proof(cauth string) string {

	// Receive the proof of authentication from client
	proof, ok := srv.srv.ClientOk(cauth)
	if !ok {
		panic("client auth failed")
	}

	srv.k = srv.srv.RawKey()

	// Send proof to the client
	fmt.Printf("Server Authenticator: M' --> Server\n   %s\n", proof)
	return proof
}

type client struct {
	ih string
	k  []byte

	s *srp.SRP
	c *srp.Client
}

func (cli *client) f1(i []byte, pass []byte) (A *big.Int) {

	ss, err := srp.New(bits)
	if err != nil {
		panic(err)
	}
	cli.s = ss

	cc, err := cli.s.NewClient(i, pass)
	if err != nil {
		panic(err)
	}
	cli.c = cc

	//

	// client credentials (public key and identity) to send to server
	creds := cli.c.Credentials()

	fmt.Printf("Client Begin; <I, A> --> server:\n   %s\n", creds)

	// Begin the server by parsing the client public key and identity.
	ih, A, err := srp.ServerBegin(creds)
	if err != nil {
		panic(err)
	}
	cli.ih = ih
	return A
}

func (cli *client) proof(creds string) string {

	// client processes the server creds and generates
	// a mutual authenticator; the authenticator is sent
	// to the server as proof that the client derived its keys.
	cauth, err := cli.c.Generate(creds)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Client Authenticator: M --> Server\n   %s\n", cauth)

	return cauth
}

func (cli *client) key(proof string) []byte {

	// Verify the server's proof
	if !cli.c.ServerOk(proof) {
		panic("server auth failed")
	}

	// Now, we have successfully authenticated the client to the
	// server and vice versa.

	kc := cli.c.RawKey()
	cli.k = kc
	return kc
}

func srpfunc_opencoff(iden, password string) bool {
	pass := []byte(password)
	i := []byte(iden)

	srv := &server{}
	cli := &client{}

	srv.register(i, pass)
	A := cli.f1(i, pass)
	creds := srv.f1(A)

	cauth := cli.proof(creds)
	proof := srv.proof(cauth)

	cli.key(proof)

	fmt.Printf("Client Key: %x\nServer Key: %x\n", cli.k, srv.k)

	return subtle.ConstantTimeCompare(srv.k, cli.k) == 1
}
