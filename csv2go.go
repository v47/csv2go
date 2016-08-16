package main

import (
	"encoding/csv"
	"bufio"
	"log"
	"os"
	"text/template"
	"errors"
	"regexp"
	"strings"
	"flag"
	"fmt"
)

type GenStruct struct {
	Name    string
	GenVars []GenVar
}

type GenVar struct {
	Name string
	Type string
}

type GenMap struct {
	Name         string
	IndexType    string
	ValType      string
	GenMapValues []GenMapValue
}
type GenMapValue struct {
	Key     string
	ValType string
	Val     []GenVal
}

type GenVal struct {
	GenValType string
	Val        string
}

const headerTemp = `//Generated code by csv2go
package {{.}}
`
const dataStruct = `
type {{.Name}} struct { {{range $index, $val := .GenVars}}
        {{$val.Name}} {{ $val.Type}}{{end}}
}

`

const dataMap = `
var {{.Name}} = map[{{.IndexType}}]{{.ValType}}{ {{range $index, $val := .GenMapValues}}
        "{{$val.Key}}": {{.ValType}} { {{range $i, $v := $val.Val}}`+"`"+`{{$v.Val}}`+"`"+`, {{end}} },{{end}}
}

`

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of csv2go:")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  csv2go -hashKey=<hash key name> -hashKeyType=<dataType> -inputFile=<PathToCsvFile>")
		fmt.Fprintln(os.Stderr, "")
		flag.PrintDefaults()
	}

	flag.CommandLine.Init("", flag.ExitOnError)
}

func main() {
	hashIndex := *flag.String("hashKey", "", "Determines name of index")
	hashIndexType := *flag.String("hashKeyType", "string", "Type of hash key")
	dataStrName := *flag.String("dataStrName", "", "Data structure for rows")
	mapName := *flag.String("varName", "", "Name of var with data from csv")
	inputFile := *flag.String("inputFile", "", "Path to input csv file")

	flag.Parse()

	hashIndex = "ISO3"
	dataStrName = "countrie"
	mapName = "countriesList"
	inputFile="/Users/viktor/work/src/github.com/fraugster/utils/csv2go/countries_list.ignore"

	dataStrName=strings.ToUpper((dataStrName)[:1])+strings.ToLower(dataStrName)[1:]
	packageName:=strings.ToLower(mapName)
	mapName=strings.ToUpper(mapName[:1])+packageName[1:]


	f, _ := os.Open(inputFile)
	r := csv.NewReader(bufio.NewReader(f))
	r.Comma = '	'

	firstLine, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}


	var GenRec GenStruct
	GenRec.Name = dataStrName
	hashIndexCol := -1
	for i, v := range firstLine {
		if v != hashIndex {
			GenRec.GenVars = append(GenRec.GenVars, GenVar{nameNormalization(v), "string"})
		} else {
			hashIndexCol = i
		}
	}
	if hashIndexCol == -1 {
		panic(errors.New("Can't find index element"))
	}


	records, err := r.ReadAll()

	genMap := GenMap{Name: mapName, IndexType: hashIndexType, ValType: dataStrName}

	for _, row := range records {
		genVal := GenMapValue{ValType: genMap.ValType}
		for inx, col := range row {
			if inx != hashIndexCol {
				genVal.Val = append(genVal.Val, GenVal{"string", col})
			} else {
				genVal.Key = col
			}
		}
		genMap.GenMapValues = append(genMap.GenMapValues, genVal)
	}

	outPutFile, err := os.Create(packageName+".go")

	header := template.Must(template.New("headerTemp").Parse(headerTemp))

	if err = header.Execute(outPutFile, packageName); err != nil {
		log.Fatal("Can't create header")
	}

	dataStr := template.Must(template.New("dataStruct").Parse(dataStruct))

	if err = dataStr.Execute(outPutFile, GenRec); err != nil {
		log.Fatal("Can't define structure")
	}


	valTemp := template.Must(template.New("dataMap").Parse(dataMap))
	err = valTemp.Execute(outPutFile, genMap)
}

func nameNormalization(name string) (out string) {
	wordsReg := regexp.MustCompile("\\w+")
	words := wordsReg.FindAllString(name, -1)
	for _, s := range words {
		out += s
	}
	return

}
