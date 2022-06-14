package struct2csv

import (
	"archive/zip"
	"log"
	"os"
	"time"
)

func ExampleStruct2CSVFile() {
	data := []testStruct{{
		A1: 1,
		A2: true,
		B1: []int{2, 3},
		B2: []struct {
			B21 int
			B22 string
		}{
			{
				B21: 4,
				B22: "a",
			},
		},
		B3: []*struct {
			B31 []*int
		}{
			{
				B31: []*int{newInt(5)},
			},
		},
	},
	}

	conv, _ := NewStructConverter(NewHeaderAutoIncrementConv())
	result, err := conv.Convert(data)
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create("test.zip")
	if err != nil {
		log.Fatal(err)
	}

	zipW := zip.NewWriter(outFile)
	defer zipW.Close()

	f, _ := zipW.CreateHeader(&zip.FileHeader{
		Name:     "test.csv",
		Modified: time.Now(),
		Method:   zip.Deflate,
	})

	csvWriter := NewCSVWriter(f)
	if err := csvWriter.WriteCSV(result); err != nil {
		log.Fatal(err)
	}

	f2, _ := zipW.CreateHeader(&zip.FileHeader{
		Name:     "mapping.csv",
		Modified: time.Now(),
		Method:   zip.Deflate,
	})

	if err := NewCSVWriter(f2).WriteMapping(result); err != nil {
		log.Fatal(err)
	}

	// comment it
	os.Remove("test.zip")
}
