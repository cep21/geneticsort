package record

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"github.com/cep21/geneticsort/genetic"
)

type Record struct {
	Algorithm     genetic.Algorithm
	BestCandidate genetic.Individual
	//Config map[string]string
}

func mustWrite(_ int, err error) {
	if err != nil {
		panic(err)
	}
}

func (r *Record) Hash() string {
	h := sha256.New()
	mustWrite(io.WriteString(h, r.BestCandidate.String()))
	return r.Algorithm.Factory.Family() + ":" + base64.StdEncoding.EncodeToString(h.Sum(nil))
}

type Recorder interface {
	// Record a record (lol english)
	Record(ctx context.Context, r Record) error
}
