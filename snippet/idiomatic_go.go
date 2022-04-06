package snippet

import (
	"github.com/pkg/errors"
	"log"
	"time"
)

// Source: https://go-talks.appspot.com/github.com/matryer/present/idiomatic-go-tricks/main.slide#1
// 1. Return teardown functions
func StartTimer(name string) func() {
	t := time.Now()
	log.Println(name, "started")
	return func() {
		d := time.Now().Sub(t)
		log.Println(name, "took", d)
	}
}

func FunkyFunc() {
	stop := StartTimer("FunkyFunc")
	defer stop()

	time.Sleep(1 * time.Second)
}

// 2.Other ways to implement the interface
type SizeFunc func() int64

func (s SizeFunc) Size() int64 {
	return s()
}

type Size int64

func (s Size) Size() int64 {
	return int64(s)
}

// 3. OK pattern
type Valid interface {
	OK() error
}
type Person struct {
	Name string
}

func (p Person) OK() error {
	if p.Name == "" {
		return errors.New("name required")
	}
	return nil
}
func Decode(r io.Reader, v interface{}) error {
	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		return err
	}
	obj, ok := v.(Valid)
	if !ok {
		return nil // no OK method
	}
	err = obj.OK()
	if err != nil {
		return err
	}
	return nil
}

// 4. retry
type Func func(attempt int) (retry bool, err error)

func Try(fn Func) error {
	var err error
	var cont bool
	attempt := 1
	for {
		cont, err = fn(attempt)
		if !cont || err == nil {
			break
		}
		attempt++
		if attempt > MaxRetries {
			return errMaxRetriesReached
		}
	}
	return err
}

//Delay between retries
func Test() {
	var value string
	err := Try(func(attempt int) (bool, error) {
		var err error
		value, err = SomeFunction()
		if err != nil {
			time.Sleep(1 * time.Minute) // wait a minute
		}
		return attempt < 5, err
	})
	if err != nil {
		log.Fatalln("error:", err)
	}
}

// 5. empty structure
type Codec interface {
	Encode(w io.Writer, v interface{}) error
	Decode(r io.Reader, v interface{}) error
}

type jsonCodec struct{}

func (jsonCodec) Encode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
func (jsonCodec) Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

var JSON Codec = jsonCodec{}

/*
Empty struct{} to group methods together
Methods don't capture the receiver
JSON variable doesn't expose the jsonCodec type
Very simple API
*/
