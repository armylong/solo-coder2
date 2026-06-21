package cmd

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/urfave/cli/v2"
)

var (
	obfNameMap   map[string]string
	obfUsedNames map[string]bool
	obfRng       *rand.Rand
	obfSrcPath   string
	obfDstPath   string
)

func GoObfuscateHandler(c *cli.Context) error {
	projectPath := c.String("path")
	if projectPath == "" {
		return fmt.Errorf("请输入项目路径")
	}

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %v", err)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("项目路径不存在: %s", absPath)
	}

	obfRng = rand.New(rand.NewSource(time.Now().UnixNano()))
	obfNameMap = make(map[string]string)
	obfUsedNames = make(map[string]bool)
	obfSrcPath = absPath
	obfDstPath = absPath + "_obfuscated"

	// 如果目标目录已存在则删除
	if _, err := os.Stat(obfDstPath); err == nil {
		fmt.Printf("删除已有输出目录: %s\n", obfDstPath)
		os.RemoveAll(obfDstPath)
	}

	// 复制项目到新目录
	fmt.Printf("复制项目到: %s\n", obfDstPath)
	if err := copyDir(obfSrcPath, obfDstPath); err != nil {
		return fmt.Errorf("复制项目失败: %v", err)
	}

	// 收集所有go文件(从副本目录)
	goFiles, err := collectGoFiles(obfDstPath)
	if err != nil {
		return fmt.Errorf("收集go文件失败: %v", err)
	}
	fmt.Printf("找到 %d 个go文件\n", len(goFiles))

	// 按package目录分组
	fset := token.NewFileSet()
	pkgFiles := make(map[string][]string)
	for _, f := range goFiles {
		dir := filepath.Dir(f)
		pkgFiles[dir] = append(pkgFiles[dir], f)
	}

	// 第一遍: 收集所有可重命名的标识符
	fmt.Println("收集标识符...")
	for dir, files := range pkgFiles {
		if err := collectIdentifiers(fset, dir, files); err != nil {
			return fmt.Errorf("收集标识符失败 %s: %v", dir, err)
		}
	}
	fmt.Printf("收集到 %d 个可重命名标识符\n", len(obfNameMap))

	// 第二遍: 执行混淆
	fmt.Println("执行混淆...")
	for dir, files := range pkgFiles {
		if err := obfuscatePackage(fset, dir, files); err != nil {
			return fmt.Errorf("混淆失败 %s: %v", dir, err)
		}
	}

	// go build验证
	fmt.Println("验证编译...")
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = obfDstPath
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("编译失败, 副本目录: %s, 原项目未修改", obfDstPath)
	}

	fmt.Printf("混淆完成! 输出目录: %s\n", obfDstPath)
	return nil
}

func collectGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func collectIdentifiers(fset *token.FileSet, dir string, files []string) error {
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		node, err := parser.ParseFile(fset, file, src, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("解析文件失败 %s: %v", file, err)
		}

		for _, decl := range node.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if shouldSkipFunc(d.Name.Name) {
					collectFuncParams(d, dir)
					continue
				}
				if isExported(d.Name.Name) {
					collectFuncParams(d, dir)
					continue
				}
				registerName(d.Name.Name, dir)
				collectFuncParams(d, dir)
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if !isExported(s.Name.Name) && !shouldSkipExported(s.Name.Name) {
							registerName(s.Name.Name, dir)
						}
						if structType, ok := s.Type.(*ast.StructType); ok {
							collectStructFields(structType, dir)
						}
					case *ast.ValueSpec:
						for _, ident := range s.Names {
							if !isExported(ident.Name) && !shouldSkipExported(ident.Name) {
								registerName(ident.Name, dir)
							}
						}
					}
				}
			}
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt:
				for _, expr := range x.Lhs {
					if ident, ok := expr.(*ast.Ident); ok && !isExported(ident.Name) {
						registerName(ident.Name, dir)
					}
				}
			case *ast.RangeStmt:
				if ident, ok := x.Key.(*ast.Ident); ok && !isExported(ident.Name) {
					registerName(ident.Name, dir)
				}
				if ident, ok := x.Value.(*ast.Ident); ok && !isExported(ident.Name) {
					registerName(ident.Name, dir)
				}
			case *ast.FuncLit:
				collectFuncLitParams(x, dir)
			}
			return true
		})
	}
	return nil
}

func collectStructFields(structType *ast.StructType, dir string) {
}

func collectFuncParams(fn *ast.FuncDecl, dir string) {
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				if !isExported(name.Name) {
					registerName(name.Name, dir)
				}
			}
		}
	}
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			for _, name := range result.Names {
				if !isExported(name.Name) {
					registerName(name.Name, dir)
				}
			}
		}
	}
}

func collectFuncLitParams(fn *ast.FuncLit, dir string) {
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			for _, name := range param.Names {
				if !isExported(name.Name) {
					registerName(name.Name, dir)
				}
			}
		}
	}
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			for _, name := range result.Names {
				if !isExported(name.Name) {
					registerName(name.Name, dir)
				}
			}
		}
	}
}

func shouldSkipFunc(name string) bool {
	if name == "init" || name == "main" {
		return true
	}
	if strings.HasPrefix(name, "Action") {
		return true
	}
	return false
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func shouldSkipExported(name string) bool {
	if strings.HasPrefix(name, "Action") {
		return true
	}
	return false
}

func hasJsonTag(field *ast.Field) bool {
	if field.Tag == nil {
		return false
	}
	return strings.Contains(field.Tag.Value, "json:")
}

func registerName(name, dir string) {
	if name == "_" || name == "" {
		return
	}
	key := dir + "::" + name
	if _, exists := obfNameMap[key]; exists {
		return
	}
	newName := generateName(name)
	obfNameMap[key] = newName
	obfUsedNames[newName] = true
}

// 生成与原名同长度的随机名称, 碰撞则重试
func generateName(oldName string) string {
	length := len(oldName)
	if length < 2 {
		length = 2
	}
	for attempt := 0; attempt < 100; attempt++ {
		result := make([]byte, length)
		for i := 0; i < length; i++ {
			result[i] = byte('a' + obfRng.Intn(26))
		}
		if len(oldName) > 0 && oldName[0] >= 'A' && oldName[0] <= 'Z' {
			result[0] = result[0] - 32
		}
		candidate := string(result)
		if !obfUsedNames[candidate] && !isBuiltin(candidate) && !isGoKeyword(candidate) {
			return candidate
		}
	}
	// 碰撞太多, 加长一位
	result := make([]byte, length+1)
	for i := 0; i < length+1; i++ {
		result[i] = byte('a' + obfRng.Intn(26))
	}
	if len(oldName) > 0 && oldName[0] >= 'A' && oldName[0] <= 'Z' {
		result[0] = result[0] - 32
	}
	return string(result)
}

func isBuiltin(name string) bool {
	builtins := map[string]bool{
		"true": true, "false": true, "nil": true,
		"bool": true, "int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "complex64": true, "complex128": true,
		"string": true, "byte": true, "rune": true, "error": true,
		"make": true, "new": true, "len": true, "cap": true, "append": true,
		"copy": true, "delete": true, "close": true, "panic": true, "recover": true,
		"print": true, "println": true, "complex": true, "real": true, "imag": true,
		"any": true, "comparable": true,
	}
	return builtins[name]
}

func isGoKeyword(name string) bool {
	keywords := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	return keywords[name]
}

func obfuscatePackage(fset *token.FileSet, dir string, files []string) error {
	for _, file := range files {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("混淆文件panic: %s, error: %v\n%s\n", file, r, debug.Stack())
				}
			}()
			if err := obfuscateFile(fset, dir, file); err != nil {
				fmt.Printf("混淆文件失败: %s, error: %v\n", file, err)
			}
		}()
	}
	return nil
}

func obfuscateFile(fset *token.FileSet, dir, filePath string) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	node, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("解析文件失败 %s: %v", filePath, err)
	}

	// 删除所有注释
	removeComments(node)

	// 收集import别名, 这些不能重命名
	importAliases := make(map[string]bool)
	for _, imp := range node.Imports {
		if imp.Name != nil && imp.Name.Name != "_" && imp.Name.Name != "." {
			importAliases[imp.Name.Name] = true
		} else {
			path := strings.Trim(imp.Path.Value, `"`)
			parts := strings.Split(path, "/")
			if len(parts) > 0 {
				importAliases[parts[len(parts)-1]] = true
			}
		}
	}

	// 收集import别名选择器的Sel, 这些不能重命名(外部包的字段/方法)
	externalSels := make(map[*ast.Ident]bool)
	ast.Inspect(node, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && importAliases[ident.Name] {
				externalSels[sel.Sel] = true
			}
		}
		return true
	})

	// 重命名标识符(跳过package名、import别名、外部包选择器)
	ast.Inspect(node, func(n ast.Node) bool {
		if x, ok := n.(*ast.Ident); ok {
			if x == node.Name {
				return true
			}
			if importAliases[x.Name] {
				return true
			}
			if externalSels[x] {
				return true
			}
			newName := lookupName(x.Name, dir)
			if newName != "" {
				x.Name = newName
			}
		}
		return true
	})

	// 字符串拆分和表达式变换(需要替换节点, 所以用单独遍历)
	transformLiterals(node)

	// 重排声明
	reorderDeclarations(node)

	// 注入废代码
	injectDeadCode(node)

	// 修复AST中nil的BlockStmt.List
	fixNilBlocks(node)

	// 写回文件
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return fmt.Errorf("格式化输出失败 %s: %v", filePath, err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("格式化代码失败 %s: %v", filePath, err)
	}

	return os.WriteFile(filePath, formatted, 0644)
}

// 删除所有注释
func removeComments(node *ast.File) {
	node.Comments = nil
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CommentGroup:
			x.List = nil
		case *ast.Field:
			x.Comment = nil
			x.Doc = nil
		case *ast.GenDecl:
			x.Doc = nil
			for _, spec := range x.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					vs.Doc = nil
					vs.Comment = nil
				}
				if ts, ok := spec.(*ast.TypeSpec); ok {
					ts.Doc = nil
					ts.Comment = nil
				}
				if is, ok := spec.(*ast.ImportSpec); ok {
					is.Doc = nil
					is.Comment = nil
				}
			}
		case *ast.FuncDecl:
			x.Doc = nil
		}
		return true
	})
}

// 字符串拆分和表达式变换
func transformLiterals(node *ast.File) {
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			transformBinaryExpr(x)
		case *ast.CallExpr:
			if x.Args == nil {
				return true
			}
			for i, arg := range x.Args {
				if arg == nil {
					continue
				}
				if ident, ok := arg.(*ast.Ident); ok && ident.Name == "true" {
					x.Args[i] = &ast.UnaryExpr{
						Op: token.NOT,
						X:  &ast.Ident{Name: "false"},
					}
				}
				if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if newExpr := splitString(lit); newExpr != nil {
						x.Args[i] = newExpr
					}
				}
			}
		case *ast.AssignStmt:
			if x.Rhs == nil {
				return true
			}
			for i, rhs := range x.Rhs {
				if rhs == nil {
					continue
				}
				if ident, ok := rhs.(*ast.Ident); ok && ident.Name == "true" {
					x.Rhs[i] = &ast.UnaryExpr{
						Op: token.NOT,
						X:  &ast.Ident{Name: "false"},
					}
				}
				if lit, ok := rhs.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if newExpr := splitString(lit); newExpr != nil {
						x.Rhs[i] = newExpr
					}
				}
			}
		case *ast.ReturnStmt:
			if x.Results == nil {
				return true
			}
			for i, result := range x.Results {
				if result == nil {
					continue
				}
				if ident, ok := result.(*ast.Ident); ok && ident.Name == "true" {
					x.Results[i] = &ast.UnaryExpr{
						Op: token.NOT,
						X:  &ast.Ident{Name: "false"},
					}
				}
				if lit, ok := result.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if newExpr := splitString(lit); newExpr != nil {
						x.Results[i] = newExpr
					}
				}
			}
		case *ast.ValueSpec:
			if x.Values == nil {
				return true
			}
			for i, val := range x.Values {
				if val == nil {
					continue
				}
				if ident, ok := val.(*ast.Ident); ok && ident.Name == "true" {
					x.Values[i] = &ast.UnaryExpr{
						Op: token.NOT,
						X:  &ast.Ident{Name: "false"},
					}
				}
				if lit, ok := val.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if newExpr := splitString(lit); newExpr != nil {
						x.Values[i] = newExpr
					}
				}
			}
		case *ast.IfStmt:
			if x.Cond == nil {
				return true
			}
			if ident, ok := x.Cond.(*ast.Ident); ok && ident.Name == "true" {
				x.Cond = &ast.UnaryExpr{
					Op: token.NOT,
					X:  &ast.Ident{Name: "false"},
				}
			}
		}
		return true
	})
}

func fixNilBlocks(node *ast.File) {
	emptyBlock := &ast.BlockStmt{List: []ast.Stmt{}}
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BlockStmt:
			if x.List == nil {
				x.List = []ast.Stmt{}
			}
		case *ast.IfStmt:
			if x.Body == nil {
				x.Body = emptyBlock
			}
			if x.Else != nil {
				if block, ok := x.Else.(*ast.BlockStmt); ok && block.List == nil {
					block.List = []ast.Stmt{}
				}
			}
		case *ast.ForStmt:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		case *ast.RangeStmt:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		case *ast.FuncDecl:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		case *ast.FuncLit:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		case *ast.SwitchStmt:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		case *ast.CaseClause:
			if x.Body == nil {
				x.Body = []ast.Stmt{}
			}
		case *ast.SelectStmt:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		case *ast.CommClause:
			if x.Body == nil {
				x.Body = []ast.Stmt{}
			}
		case *ast.TypeSwitchStmt:
			if x.Body == nil {
				x.Body = emptyBlock
			}
		}
		return true
	})
}

func splitString(lit *ast.BasicLit) *ast.BinaryExpr {
	value := lit.Value
	if len(value) < 6 {
		return nil
	}
	if strings.HasPrefix(value, "`") {
		return nil
	}
	inner := value[1 : len(value)-1]
	if len(inner) < 4 {
		return nil
	}
	// 按Go源码转义边界拆分, 避免截断转义序列
	boundaries := findSplitBoundaries(inner)
	if len(boundaries) < 2 {
		return nil
	}
	splitPos := boundaries[1+obfRng.Intn(len(boundaries)-1)]
	part1 := value[:1] + inner[:splitPos] + value[:1]
	part2 := value[:1] + inner[splitPos:] + value[:1]
	return &ast.BinaryExpr{
		X:  &ast.BasicLit{Kind: token.STRING, Value: part1},
		Op: token.ADD,
		Y:  &ast.BasicLit{Kind: token.STRING, Value: part2},
	}
}

func findSplitBoundaries(s string) []int {
	var boundaries []int
	i := 0
	for i < len(s) {
		boundaries = append(boundaries, i)
		if s[i] == '\\' {
			if i+1 >= len(s) {
				break
			}
			switch s[i+1] {
			case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', '\'', '"':
				i += 2
			case '0', '1', '2', '3', '4', '5', '6', '7':
				i += 2
				if i < len(s) && s[i] >= '0' && s[i] <= '7' {
					i++
					if i < len(s) && s[i] >= '0' && s[i] <= '7' {
						i++
					}
				}
			case 'x':
				i += 4
			case 'u':
				i += 6
			case 'U':
				i += 10
			default:
				i += 2
			}
		} else if s[i] >= 0x80 {
			// UTF-8多字节字符, 跳过完整字符
			_, size := utf8.DecodeRuneInString(s[i:])
			i += size
		} else {
			i++
		}
	}
	return boundaries
}

func transformBinaryExpr(expr *ast.BinaryExpr) {
	if expr.X == nil || expr.Y == nil {
		return
	}
	switch expr.Op {
	case token.ADD:
		if isNumericExpr(expr.X) && isNumericExpr(expr.Y) {
			expr.Op = token.SUB
			expr.Y = &ast.UnaryExpr{
				Op: token.SUB,
				X:  expr.Y,
			}
		}
	case token.EQL:
		expr.X, expr.Y = expr.Y, expr.X
	case token.LSS:
		expr.Op = token.GTR
		expr.X, expr.Y = expr.Y, expr.X
	case token.GTR:
		expr.Op = token.LSS
		expr.X, expr.Y = expr.Y, expr.X
	}
}

func isNumericExpr(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.BasicLit:
		return x.Kind == token.INT || x.Kind == token.FLOAT
	case *ast.Ident:
		return false
	case *ast.UnaryExpr:
		return isNumericExpr(x.X)
	case *ast.BinaryExpr:
		return isNumericExpr(x.X) && isNumericExpr(x.Y)
	default:
		return false
	}
}

func lookupName(name, dir string) string {
	if name == "_" || name == "" {
		return ""
	}
	if isBuiltin(name) || isGoKeyword(name) {
		return ""
	}
	key := dir + "::" + name
	if newName, ok := obfNameMap[key]; ok {
		return newName
	}
	return ""
}

func reorderDeclarations(node *ast.File) {
	var imports []*ast.GenDecl
	var others []ast.Decl
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			imports = append(imports, genDecl)
		} else {
			others = append(others, decl)
		}
	}
	obfRng.Shuffle(len(others), func(i, j int) {
		others[i], others[j] = others[j], others[i]
	})
	node.Decls = make([]ast.Decl, 0, len(imports)+len(others))
	for _, imp := range imports {
		node.Decls = append(node.Decls, imp)
	}
	node.Decls = append(node.Decls, others...)
}

// 注入废代码: 废方法 + 包级废变量
func injectDeadCode(node *ast.File) {
	deadFuncCount := 1 + obfRng.Intn(3)
	for i := 0; i < deadFuncCount; i++ {
		deadFunc := generateDeadFunc()
		node.Decls = append(node.Decls, deadFunc)
	}

	deadVarCount := 1 + obfRng.Intn(3)
	for i := 0; i < deadVarCount; i++ {
		deadVar := generateDeadVar()
		node.Decls = append(node.Decls, deadVar)
	}
}

// 生成废方法
func generateDeadFunc() *ast.FuncDecl {
	funcName := generateUniqueName(true)
	paramName := generateUniqueName(false)

	var body []ast.Stmt
	retStmt := &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `""`},
		},
	}
	switch obfRng.Intn(3) {
	case 0:
		body = []ast.Stmt{retStmt}
	case 1:
		body = []ast.Stmt{
			&ast.ForStmt{
				Init: &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(paramName)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "0"}},
				},
				Post: &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent(paramName)},
					Tok: token.ADD_ASSIGN,
					Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
				},
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent(paramName),
					Op: token.LSS,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "10"},
				},
			},
			retStmt,
		}
	case 2:
		body = []ast.Stmt{
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{
					X:  ast.NewIdent(paramName),
					Op: token.GTR,
					Y:  &ast.BasicLit{Kind: token.INT, Value: "0"},
				},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{ast.NewIdent(paramName)},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{&ast.BinaryExpr{
								X:  ast.NewIdent(paramName),
								Op: token.MUL,
								Y:  &ast.BasicLit{Kind: token.INT, Value: "2"},
							}},
						},
					},
				},
				Else: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{
							Lhs: []ast.Expr{ast.NewIdent(paramName)},
							Tok: token.ASSIGN,
							Rhs: []ast.Expr{&ast.UnaryExpr{
								Op: token.SUB,
								X:  ast.NewIdent(paramName),
							}},
						},
					},
				},
			},
			retStmt,
		}
	}

	return &ast.FuncDecl{
		Name: ast.NewIdent(funcName),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent(paramName)},
						Type:  ast.NewIdent("int"),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{Type: ast.NewIdent("string")},
				},
			},
		},
		Body: &ast.BlockStmt{List: body},
	}
}

// 生成包级废变量
func generateDeadVar() *ast.GenDecl {
	varName := generateUniqueName(true)
	return &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names: []*ast.Ident{ast.NewIdent(varName)},
				Type:  ast.NewIdent("int"),
				Values: []ast.Expr{
					&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", obfRng.Intn(1000))},
				},
			},
		},
	}
}

// 生成不重复的随机名称
func generateUniqueName(exported bool) string {
	for attempt := 0; attempt < 100; attempt++ {
		length := 4 + obfRng.Intn(5)
		result := make([]byte, length)
		for i := 0; i < length; i++ {
			result[i] = byte('a' + obfRng.Intn(26))
		}
		if exported {
			result[0] = result[0] - 32
		}
		candidate := string(result)
		if !obfUsedNames[candidate] {
			obfUsedNames[candidate] = true
			return candidate
		}
	}
	return fmt.Sprintf("Z%d", obfRng.Intn(99999))
}

func copyDir(src, dst string) error {
	return exec.Command("cp", "-r", src, dst).Run()
}
