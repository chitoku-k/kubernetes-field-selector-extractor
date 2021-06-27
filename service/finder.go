package service

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"

	"github.com/chitoku-k/kubernetes-field-selector-extractor/domain"
	"github.com/sirupsen/logrus"
)

var (
	noop = func(fs.FileInfo) bool { return true }
)

type finderService struct {
	dir string
}

type FinderService interface {
	Do() ([]domain.FieldSelector, error)
}

func NewFinderService(dir string) FinderService {
	return &finderService{
		dir: dir,
	}
}

func (f *finderService) Do() ([]domain.FieldSelector, error) {
	var result []domain.FieldSelector

	set := token.NewFileSet()
	pkgs, err := parser.ParseDir(set, f.dir, noop, parser.Mode(0))
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	for name, pkg := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}

		logrus.Debugf("Package: %q", pkg.Name)

		for filename, file := range pkg.Files {
			logrus.Debugf("| Filename: %q", filename)

			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				selectorExpr, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if selectorExpr.Sel.Name != "AddFieldLabelConversionFunc" {
					return false
				}
				if len(call.Args) != 2 {
					logrus.Warnf("| | Invalid number of arguments passed to AddFieldLabelConversionFunc()")
					return false
				}

				kindExpr, ok := call.Args[0].(*ast.CallExpr)
				if !ok {
					logrus.Warnf("| | Unexpected expression for gvk: %#v", call.Args[0])
					return false
				}
				if len(kindExpr.Args) != 1 {
					logrus.Warnf("| | Invalid number of arguments passed to WithKind()")
					return false
				}

				var kind string
				kindArg, ok := kindExpr.Args[0].(*ast.BasicLit)
				if !ok {
					logrus.Warnf("| | Unexpected expression for kind: %#v", kindExpr.Args[0])
					kind = fmt.Sprintf("%#v", kindExpr.Args[0])
				} else {
					kind, err = strconv.Unquote(kindArg.Value)
					if err != nil {
						logrus.Warnf("| | Failed to unquote kind: %v", kindArg.Value)
						return false
					}
				}
				logrus.Debugf("| | Kind: %v", kind)
				selector := domain.FieldSelector{
					File:   filename,
					Kind:   kind,
					Labels: []string{},
				}

				funcExpr, ok := call.Args[1].(*ast.FuncLit)
				if !ok {
					logrus.Warnf("| | Unexpected expression for conversionFunc: %#v", call.Args[1])
					return false
				}

				var foundSwitchExpr bool
				for _, stmt := range funcExpr.Body.List {
					switchExpr, ok := stmt.(*ast.SwitchStmt)
					if !ok {
						continue
					}
					foundSwitchExpr = true

					for _, expr := range switchExpr.Body.List {
						caseClause, ok := expr.(*ast.CaseClause)
						if !ok {
							continue
						}

						for _, caseExpr := range caseClause.List {
							labelExpr, ok := caseExpr.(*ast.BasicLit)
							if !ok {
								continue
							}

							label, err := strconv.Unquote(labelExpr.Value)
							if err != nil {
								logrus.Warnf("| | | Failed to unquote: %v", labelExpr.Value)
								continue
							}
							logrus.Debugf("| | | Found: %q", label)
							selector.Labels = append(selector.Labels, label)
						}
					}
				}

				if !foundSwitchExpr {
					logrus.Warnf("| | | Not found switch expr")
				}
				result = append(result, selector)

				return false
			})
		}
	}

	return result, err
}
