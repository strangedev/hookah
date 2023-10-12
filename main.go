package main

import (
	"fmt"
	"github.com/dave/jennifer/jen"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"strings"
)

func loadPackage(path string) *packages.Package {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		panic(fmt.Errorf("loading packages for inspection: %v", err))
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0]
}

func main() {
	sourcePackage := os.Args[1]
	sourceType := os.Args[2]

	pkg := loadPackage(sourcePackage)

	obj := pkg.Types.Scope().Lookup(sourceType)
	if obj == nil {
		panic(fmt.Errorf("%s not found in declared types of %s",
			sourceType, pkg))
	}
	if _, ok := obj.(*types.TypeName); !ok {
		panic(fmt.Errorf("%v is not a named type", obj))
	}

	switch obj.Type().Underlying().(type) {
	case *types.Interface:
		// ok
	case *types.Struct:
		// ok
	default:
		panic(fmt.Errorf("type %v is not a struct or interface", obj))
	}

	segments := strings.Split(pkg.String(), "/")
	goPackage := segments[len(segments)-1]

	f := jen.NewFile(goPackage)
	f.PackageComment("Code generated by hookah, DO NOT EDIT.")
	f.ImportAlias("github.com/strangedev/hookah/pkg", "hookah")
	f.Type().Id("Hooked" + sourceType).Struct(
		jen.Op("*").Qual("github.com/strangedev/hookah/pkg", "Hookah").Index(jen.Id(sourceType)),
	)

	methodSet := types.NewMethodSet(obj.Type())
	ptrMethodSet := types.NewMethodSet(types.NewPointer(obj.Type()))
	declaredMethods := make(map[string]*types.Selection)
	for i := 0; i < methodSet.Len(); i++ {
		method := methodSet.At(i)
		methodName := method.Obj().Name()
		declaredMethods[methodName] = method
	}
	for i := 0; i < ptrMethodSet.Len(); i++ {
		method := ptrMethodSet.At(i)
		methodName := method.Obj().Name()
		if _, ok := declaredMethods[methodName]; ok {
			// Automatic indirect method
			continue
		}
		declaredMethods[methodName] = method
	}

	for name, method := range declaredMethods {
		methodImpl := method.Obj().(*types.Func)
		signature := methodImpl.Type().Underlying().(*types.Signature)

		generatedFunc := f.Func().Id("(h " + "Hooked" + sourceType + ")").Id(name)

		params := make([]jen.Code, 0, signature.Params().Len())
		paramNames := make([]jen.Code, 0, signature.Params().Len()+1)
		paramNames = append(paramNames, jen.Lit(name))
		for i := 0; i < signature.Params().Len(); i++ {
			param := signature.Params().At(i)
			params = append(params, jen.Id(param.Name()).Id(param.Type().String()))
			paramNames = append(paramNames, jen.Id(param.Name()))
		}

		returns := make([]jen.Code, 0, signature.Results().Len())
		for i := 0; i < signature.Results().Len(); i++ {
			returnDeclaration := signature.Results().At(i)
			returns = append(returns, jen.Id(returnDeclaration.Name()).Id(returnDeclaration.Type().String()))
		}

		body := make([]jen.Code, 0)
		returnList := make([]jen.Code, 0, signature.Results().Len())

		if len(returns) > 0 {
			body = append(body, jen.Id("returnValues").Op(":=").Id("h").Dot("RunMethodWithReturnHooks").Call(paramNames...))
		} else {
			body = append(body, jen.Id("_").Op("=").Id("h").Dot("RunMethodWithReturnHooks").Call(paramNames...))
		}

		for i := 0; i < signature.Results().Len(); i++ {
			returnDeclaration := signature.Results().At(i)
			returnName := returnDeclaration.Name()
			if returnName == "" {
				returnName = "returnValue" + fmt.Sprintf("%d", i)
			}
			switch returnDeclaration.Type().Underlying().(type) {
			case *types.Interface:
				returnVar := jen.Var().Id(returnName).Id(returnDeclaration.Type().String())
				anyVarName := returnName + "Any"
				anyVar := jen.Id(anyVarName).Op(":=").Id("returnValues").Index(jen.Lit(i)).Dot("Interface").Call()
				typeCheck := jen.If(jen.Id(anyVarName)).Op("!=").Nil().Block(
					jen.Id(returnName).Op("=").Id(returnName + "Any").Assert(jen.Id(returnDeclaration.Type().String())),
				)
				body = append(body, returnVar, anyVar, typeCheck)
			default:
				returnVar := jen.Id(returnName).Op(":=").Id("returnValues").Index(jen.Lit(i)).Dot("Interface").Call().Assert(jen.Id(returnDeclaration.Type().String()))
				body = append(body, returnVar)
			}

			returnList = append(returnList, jen.Id(returnName))
		}

		if len(returns) > 0 {
			body = append(body, jen.Return().List(returnList...))
		}

		generatedFunc.
			Params(params...).
			List(returns...).
			Block(body...)
	}

	outputPath := "generated.go"
	if err := os.WriteFile(outputPath, []byte(f.GoString()), 0644); err != nil {
		panic(err)
	}
}
