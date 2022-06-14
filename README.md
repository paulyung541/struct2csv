# struct2csv
a lib convert go struct to csv file

the struct like this
```go
type testStruct struct {
	A1 int
	A2 bool
	B1 []int
	B2 []struct {
		B21 int
		B22 string
	}
	B3 []*struct {
		B31 []*int
	}
}
```
after the transformation, the header like follow
```text
"/A1"
"/A2"
"/B1/0"
"/B1/1"
"/B2/0/B21"
"/B2/0/B22"
"/B3/0/B31/0"
"/B3/0/B31/1"
```