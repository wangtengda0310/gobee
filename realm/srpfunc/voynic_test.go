package srpfunc

import (
	"crypto/subtle"
	"github.com/voynic/srp"
	"testing"
)

func TestName(t *testing.T) {
	// A client chooses an identifier and passphrase
	identifier := []byte("user123@example.com")
	passphrase := []byte("Password123!")

	// SRP creates a salt and verifier based on the client's identifier and passphrase
	s, v, err := srp.NewClient(identifier, passphrase)

	if err != nil {
		panic("Client creation failed!")
	}

	// The client's identifier and passphrase, registered with the server.

	// The client creates a public key "A" and a private key "a"
	A, a, err := srp.InitiateHandshake()

	if err != nil {
		panic("Handshake initiation failed!")
	}

	// The server receives a client's "identifier" and "A" value. Assume the
	// following variables are populated accordingly.

	// The server looks up a client's salt and verifier from the provided
	// identifier. Assume the following variables are populated accordingly.

	// Create a public key to share with the client, and compute the session key.
	B, s, K, err := srp.Handshake(A, v)

	if err != nil {
		panic("Handshake failed!")
	}
	// The client receives its salt "s" along with a public key "B" from the server.
	// Assume the following variables are populated accordingly.

	// Recall that the client has "A", "a", and "passphrase" variables from the
	// first step of session creation.

	// Compute the session key!
	K, err = srp.CompleteHandshake(A, a, identifier, passphrase, s, B)

	if err != nil {
		panic("Failed to complete the handshake!")
	}

	proof := srp.Hash(K)

	// The server received the client's proof, and assigns it to the variable below:
	var clientProof []byte = proof

	// Check if the client's proof is acceptable
	if subtle.ConstantTimeCompare(clientProof, srp.Hash(K)) != 1 {
		panic("Server does not accept client's proof!")
	}

	serverProof := srp.Hash(s, K)

	// The client received the server's proof, and assigns it to the variable below:

	if subtle.ConstantTimeCompare(serverProof, srp.Hash(s, K)) != 1 {
		panic("Client does not accept server's proof!")
	}

}
