package snippet

// Source: https://medium.com/picus-security-engineering/custom-json-marshaller-in-go-and-common-pitfalls-c43fa774db05
import (
	"encoding/json"
	"fmt"
)

type Status int

const (
	Waiting Status = iota
	Queued
	Completed
	Error
)

func (s Status) String() string {
	statuses := [...]string{"Waiting", "Queued", "Completed", "Error"}
	if len(statuses) < int(s) {
		return ""
	}
	return statuses[s]
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}
	*s = ToStatus(statusStr)
	return nil
}

func ToStatus(b string) Status {
	switch b {
	case "Waiting":
		return Waiting
	case "Queued":
		return Queued
	case "Error":
		return Error
	case "Completed":
		return Completed
	default:
		return -1
	}
}

type Foo struct {
	FooID     string `json:"fooId"`
	FooStatus Status `json:"fooStatus"`
	FooBars   []Bar  `json:"fooBars"`
}

func (f *Foo) MarshalJSON() ([]byte, error) {
	type FooAlias Foo
	type NewFoo struct {
		FooStatus string `json:"fooStatus"`
		*FooAlias
	}

	newFoo := &NewFoo{FooStatus: f.FooStatus.String(), FooAlias: (*FooAlias)(f)}

	return json.Marshal(newFoo)
}

func (f *Foo) UnmarshalJSON(data []byte) error {
	type Alias Foo
	aux := &struct {
		FooStatus string `json:"fooStatus"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	f.FooStatus = ToStatus(aux.FooStatus)
	return nil
}

type Bar struct {
	BarID     string `json:"barId"`
	BarStatus Status `json:"barStatus"`
}

func (b *Bar) MarshalJSON() ([]byte, error) {
	type Alias Bar
	return json.Marshal(&struct {
		BarStatus string `json:"barStatus"`
		*Alias
	}{
		BarStatus: b.BarStatus.String(),
		Alias:     (*Alias)(b),
	})
}

func (b *Bar) UnmarshalJSON(data []byte) error {
	type Alias Bar
	aux := &struct {
		BarStatus string `json:"barStatus"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.BarStatus = ToStatus(aux.BarStatus)
	return nil
}

func main() {

	bar1 := Bar{
		BarID:     "this is bar-1",
		BarStatus: Queued,
	}
	bar2 := Bar{
		BarID:     "this is bar-2",
		BarStatus: Completed,
	}
	foo := Foo{
		FooID:     "this is the foo",
		FooStatus: Queued,
		FooBars:   []Bar{bar1, bar2},
	}
	marshalled, err := json.Marshal(&foo)
	if err != nil {
		fmt.Printf("Oops! Error when marshalling %+v \n", err)
	}

	fmt.Printf("This is marshalled = %s\n", marshalled)
	var unmarshalled Foo
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		fmt.Printf("Oops! Error when unmarshalling %v\n", err)
	}
	fmt.Printf("Success! This is unmarshalled = %s\n", unmarshalled)
}
