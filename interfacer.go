package main
import (
	"go/token"
	"go/parser"
	"go/ast"
	"fmt"
	"go/types"
	"go/importer"

	"errors"
	"strings"
	"os"
	"path/filepath"
	"io/ioutil"
)

const (
	GO_PATH_ENV_KEY = "GOPATH"
	GO_ROOT_ENV_KEY = "GOROOT"
)

func findPackage(pkgPath string) (string, error) {
	GoPath := os.Getenv(GO_PATH_ENV_KEY)
	GoRoot := os.Getenv(GO_ROOT_ENV_KEY)
	currentPath, err := os.Getwd()
	if(err != nil){
		panic(err)
	}
	goCurrentPath := filepath.Join(currentPath, pkgPath)
	if _, err := os.Stat(goCurrentPath); err == nil {
		return goCurrentPath, nil
	}
	goPathSrcLib := filepath.Join(GoPath, "src", pkgPath)
	if _, err := os.Stat(goPathSrcLib); err == nil {
		return goPathSrcLib, nil
	}
	goRootSrcLib := filepath.Join(GoRoot, "src", pkgPath)
	if _, err := os.Stat(goRootSrcLib); err == nil {
		return goRootSrcLib, nil
	}
	return "", errors.New("Package not found")
}

func GetTypeName(src string, t ast.Expr) (string, error) {
	switch cc := t.(type) {
	case *ast.Ident:
		return cc.Name, nil
	default:
		return "", errors.New("Error type name not resolved")
	}
}

type SrcFile struct {
	AstFile *ast.File
	Source string
}

func main() {

	var pkgTargetPath string = "testdata/alpha"
	var typeNameToFind string = "B"

	const interfaceTemplate = `
	type %v interface {
		%v
	}`

	pkgPath, err := findPackage(pkgTargetPath)
	if (err != nil){
		panic(err)
	}

	fset := token.NewFileSet() // positions are relative to fset
	astFiles := make([]*SrcFile, 0)
	err = filepath.Walk(pkgPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			if ext := filepath.Ext(path); ext == ".go" {
				fileName := filepath.Base(path)
				file, err := os.Open(path)
				if(err != nil){
					panic(err)
				}
				bytes, err := ioutil.ReadAll(file)
				if(err != nil){
					panic(err)
				}
				f, err := parser.ParseFile(fset, fileName, string(bytes), 0)
				if(err != nil) {
					panic(err)
				}

				fData := &SrcFile{AstFile:f, Source: string(bytes)}
				astFiles = append(astFiles,  fData)
				file.Close()
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	var listOfFuncDecls []string = make([]string, 0)

	for _, f :=  range astFiles {
		conf := types.Config{Importer: importer.Default()}
		pkg, err := conf.Check("interface_gen", fset, []*ast.File{f.AstFile}, nil)
		if err != nil {
			panic(err)
		}

		for _, imp := range pkg.Imports() {
			fmt.Println("Package: ", imp.Name())
		}

		ast.Inspect(f.AstFile, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				fmt.Println("Struct <<<:", x.Name)
				sst := x.Type.(*ast.StructType)
				for _, fl := range sst.Fields.List {
					vv := fl.Type.(*ast.Ident)
					if vv.Obj != nil {
						if _, isStruct := vv.Obj.Decl.(*ast.TypeSpec); isStruct {
							fmt.Println("Struct Found: ", vv.Name)
						}
					}

				}
			case *ast.FuncDecl:
				var currentTypeName string

				switch cc := x.Recv.List[0].Type.(type) {
				case *ast.Ident:
					currentTypeName = cc.Name
				case *ast.StarExpr:
					s, _ := GetTypeName(f.Source, cc.X)
					currentTypeName = s
				case ast.Expr:
					s, _ := GetTypeName(f.Source, cc)
					currentTypeName = s
				}
				if currentTypeName == typeNameToFind {
					funcParamsStr := f.Source[x.Type.Params.Pos() - 1:x.Type.Params.End() - 1]
					funcResultsStr := f.Source[x.Type.Results.Pos() - 1:x.Type.Results.End() - 1]
					listOfFuncDecls = append(listOfFuncDecls, fmt.Sprintf("%v %v %v", x.Name.Name, funcParamsStr, funcResultsStr))
				}
			}
			return true
		})

	}

	ifaceFuncts := strings.Join(listOfFuncDecls, "\n")

	fmt.Println("Interface: ")
	fmt.Println(fmt.Sprintf(interfaceTemplate, typeNameToFind, ifaceFuncts))

}