package goc_test

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/gob"
	jsonv1 "encoding/json"
	"encoding/json/v2"
	mathrand "math/rand/v2"
	"testing"

	"github.com/samborkent/gorpc/goc"
)

type Object struct {
	Num uint64
	Str string
	Map map[string]string
}

var object Object

func BenchmarkJSONV1Encode(b *testing.B) {
	buf := new(bytes.Buffer)
	encoder := jsonv1.NewEncoder(buf)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}
		b.StartTimer()

		if err := encoder.Encode(object); err != nil {
			b.Log("error: " + err.Error())
			return
		}
	}
}

func BenchmarkJSONV1Decode(b *testing.B) {
	buf := new(bytes.Buffer)
	encoder := jsonv1.NewEncoder(buf)
	decoder := jsonv1.NewDecoder(buf)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}

		if err := encoder.Encode(object); err != nil {
			b.Log("Encode error: " + err.Error())
			return
		}
		b.StartTimer()

		if err := decoder.Decode(&object); err != nil {
			b.Log("Decode error: " + err.Error())
			return
		}
	}
}

func BenchmarkJSONV2Encode(b *testing.B) {
	buf := new(bytes.Buffer)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}
		b.StartTimer()

		if err := json.MarshalWrite(buf, object); err != nil {
			b.Log("error: " + err.Error())
			return
		}
	}
}

func BenchmarkJSONV2Decode(b *testing.B) {
	buf := new(bytes.Buffer)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}

		if err := json.MarshalWrite(buf, object); err != nil {
			b.Log("MarshalWrite error: " + err.Error())
			return
		}
		b.StartTimer()

		if err := json.UnmarshalRead(buf, &object); err != nil {
			b.Log("UnmarshalRead error: " + err.Error())
			return
		}
	}
}

func BenchmarkGobEncode(b *testing.B) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}
		b.StartTimer()

		if err := encoder.Encode(object); err != nil {
			b.Log("error: " + err.Error())
			return
		}
	}
}

func BenchmarkGobDecode(b *testing.B) {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	decoder := gob.NewDecoder(buf)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}

		if err := encoder.Encode(object); err != nil {
			b.Log("Encode error: " + err.Error())
			return
		}
		b.StartTimer()

		if err := decoder.Decode(&object); err != nil {
			b.Log("Decode error: " + err.Error())
			return
		}
	}
}

func BenchmarkGocEncode(b *testing.B) {
	buf := new(bytes.Buffer)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}
		b.StartTimer()

		if err := goc.EncodeTo(buf, object); err != nil {
			b.Log("error: " + err.Error())
			return
		}
	}
}

func BenchmarkGocDecode(b *testing.B) {
	buf := new(bytes.Buffer)

	for b.Loop() {
		b.StopTimer()
		buf.Reset()

		object = Object{
			Num: mathrand.Uint64(),
			Str: cryptorand.Text(),
			Map: map[string]string{
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
				cryptorand.Text(): cryptorand.Text(),
			},
		}

		if err := goc.EncodeTo(buf, object); err != nil {
			b.Log("EncodeTo error: " + err.Error())
			return
		}
		b.StartTimer()

		o, err := goc.DecodeFrom[Object](buf)
		if err != nil {
			b.Log("DecodeFrom error: " + err.Error())
			return
		}

		object = o
	}
}
