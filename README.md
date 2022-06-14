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

## speciality
- support a struct or map slice convert to csv
- supports header mapping to custom types

## how to use
```go
type struct User {
    ID   string
    Name string
}

func Converter(input []User) error {
    conv, err := struct2csv.NewStructConverter(struct2csv.NewHeaderOriginalStringConv(), struct2csv.WithResultCap(len(input)))
    if err != nil {
        return err
    }

    results, err := csvConv.Convert(input)
    if err != nil {
        return error
    }
    defer results.Reset()

    f, err := os.Create("test.csv")
    if err != nil {
        return err
    }
    defer f.Close()
	
    return struct2csv.NewCSVWriter(f).WriteCSV(results)
}
```

## License
[MIT][1]

[1]: http://opensource.org/licenses/MIT