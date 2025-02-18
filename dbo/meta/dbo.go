package meta

import (
	_ "embed"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/kcmvp/app"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

var (
	//go:embed er.tmpl
	erTmpl string
	//go:embed schema.tmpl
	schemaTmpl string
	//go:embed column.tmpl
	columnTmpl string

	dbReg  = regexp.MustCompile(`db:\s*"([^"]*)"`)
	preReg = regexp.MustCompile(`\(([^)]+)\)`)
)

type ColProperty string

const (
	baseEntity                = "github.com/kcmvp/dbo/base"
	colName       ColProperty = "col"
	colRef        ColProperty = "ref"
	colPK         ColProperty = "pk"
	colSeq        ColProperty = "seq"
	sqlTypePrefix             = "database/sql.Null"
)

type Column lo.Tuple3[string, string, string]

// Property get column properties
func (c Column) Property(property ColProperty) mo.Option[string] {
	pairs := strings.Split(c.C, ";")
	for _, pair := range pairs {
		kv := strings.FieldsFunc(pair, func(r rune) bool {
			return r == '='
		})
		if strings.TrimSpace(kv[0]) == string(property) {
			return lo.IfF(len(kv) == 2, func() mo.Option[string] {
				v := strings.TrimSpace(kv[1])
				if len(v) == 0 {
					return mo.None[string]()
				}
				return mo.Some(v)
			}).Else(mo.Some[string](string(property)))
		}
	}
	return mo.None[string]()
}

// Ref Def get column reference
func (c Column) Ref() mo.Option[string] {
	return c.Property(colRef)
}

// Attr go struct attribute for ER diagram
func (c Column) Attr() string {
	return c.A
}

// AttrType return go type of the column
func (c Column) AttrType() string {
	return c.B
}

// Name column name without definition
func (c Column) Name() string {
	return preReg.ReplaceAllString(c.Property(colName).MustGet(), "")
}

func (c Column) Properties() string {
	return c.C
}

// Nullable identify a column can be nullable or not
func (c Column) Nullable() bool {
	return strings.HasPrefix(c.AttrType(), sqlTypePrefix) || strings.HasPrefix(c.AttrType(), "*")
}

// Def generate column definition
func (c Column) Def(db string) string {
	cTyp := strings.ReplaceAll(c.AttrType(), sqlTypePrefix, "")
	if cTyp != c.AttrType() {
		cTyp = strings.ToLower(cTyp)
	}
	cTyp = strings.ReplaceAll(cTyp, "*", "")
	op := mo.TupleToOption(lo.Find(TypeMappings(), func(m TypeMapping) bool {
		return string(m.A) == cTyp
	}))
	if op.IsAbsent() {
		panic(fmt.Sprintf("can not find type mapping for %s", c.AttrType()))
	}
	typ := op.MustGet()
	def := fmt.Sprintf("%s %s", lo.If(db == "mysql", typ.B).ElseIf(db == "pg", typ.C).Else(typ.D), lo.If(!c.Nullable(), "not null").Else(""))
	// precision
	matches := preReg.FindStringSubmatch(c.Property(colName).MustGet())
	if len(matches) > 1 {
		precision := matches[1]
		precision = strings.Replace(precision, ",", ", ", -1)
		re := regexp.MustCompile(`\s+`)
		precision = re.ReplaceAllString(precision, " ")
		// Replace the value in the second string
		def = preReg.ReplaceAllStringFunc(def, func(match string) string {
			return fmt.Sprintf("(%s)", precision)
		})
	}
	if c.Property(colPK).IsPresent() && c.AttrType() == "int64" {
		def = fmt.Sprintf("%s %s", def, DB(db).MustGet().Auto)
	}
	return strings.TrimSpace(def)
}

// Key return "PK" or "FK" for ER diagram
func (c Column) Key() mo.Option[string] {
	return lo.If(c.Property(colPK).IsPresent(), mo.Some("PK")).
		ElseIf(c.Property(colRef).IsPresent(), mo.Some("FK")).Else(mo.None[string]())
}

type Table struct {
	entity  string
	name    string
	columns []Column
	pkg     *packages.Package
}

// Entity returns the entity name of the table
func (t Table) Entity() string {
	return t.entity
}

// PkgPath return the package path of the table
func (t Table) PkgPath() string {
	return t.pkg.PkgPath
}

func (t Table) PkgName() string {
	return t.pkg.Name
}

// Columns return all columns of the table
func (t Table) Columns() []Column {
	t2 := lo.Map(t.columns, func(c Column, index int) lo.Tuple2[int, Column] {
		return lo.Tuple2[int, Column]{A: lo.IfF(c.Property(colSeq).IsPresent(), func() int {
			if t := mo.TupleToResult(strconv.Atoi(c.Property(colSeq).MustGet())); t.IsOk() {
				return t.MustGet()
			} else {
				return index
			}
		}).Else(index), B: c}
	})
	slices.SortFunc(t2, func(a, b lo.Tuple2[int, Column]) int {
		return a.A - b.A
	})
	return lo.Map(t2, func(item lo.Tuple2[int, Column], _ int) Column {
		return item.B
	})
}

// Name table name, for schema generation
func (t Table) Name() string {
	return t.name
}

// MaxWidth max width of the column name, for schema generation
func (t Table) MaxWidth() int {
	return lo.Max(lo.Map(t.columns, func(item Column, _ int) int {
		return len(item.Name())
	})) + 1
}

// PK return primary key of the table
func (t Table) PK() string {
	pk := mo.TupleToOption(lo.Find(t.Columns(), func(c Column) bool {
		return c.Key().IsPresent() && c.Key().MustGet() == "PK"
	}))
	return pk.MustGet().Name()
}

// Column return the corresponding column of the attribute
func (t Table) Column(attrName string) mo.Option[Column] {
	return mo.TupleToOption(lo.Find(t.columns, func(item Column) bool {
		return item.A == attrName
	}))
}

type DBO struct {
	g graph.Graph[string, Table]
}

// Tables return all the tables of the project
func (dbo DBO) Tables() []Table {
	return lo.Map(lo.Keys(mo.TupleToResult(dbo.g.PredecessorMap()).MustGet()),
		func(item string, index int) Table {
			return mo.TupleToResult(dbo.g.Vertex(item)).MustGet()
		})
}

// Edges return all the relationships among the table
func (dbo DBO) Edges() []string {
	if edges, err := dbo.g.Edges(); err != nil {
		return []string{}
	} else {
		return lo.Map(edges, func(item graph.Edge[string], _ int) string {
			return item.Properties.Attributes["ref"]
		})
	}
}

// Table get table of the entity
func (dbo DBO) Table(entity string) Table {
	t, _ := dbo.g.Vertex(entity)
	return t
}

// ER generate ER diagram of the tables
func (dbo DBO) ER(path string) error {
	err := os.MkdirAll(path, 0755)
	if err == nil {
		file := mo.TupleToResult(os.Create(filepath.Join(path, "er.d2")))
		if file.IsError() {
			return file.Error()
		}
		defer file.MustGet().Close()
		tmpl := mo.TupleToResult(template.New("er").Parse(erTmpl))
		return lo.IfF(tmpl.IsError(), func() error {
			return tmpl.Error()
		}).ElseF(func() error {
			return tmpl.MustGet().Execute(file.MustGet(), dbo)
		})
	}
	return err
}

// Schema generate schema of the tables
func (dbo DBO) Schema(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	for _, platform := range Platforms() {
		file := mo.TupleToResult(os.Create(filepath.Join(path, fmt.Sprintf("schema-%s.sql", platform))))
		if file.IsError() {
			return file.Error()
		}
		fns := template.FuncMap{
			"db": func() string {
				return platform
			},
		}
		tmpl := mo.TupleToResult(template.New(platform).Funcs(fns).Parse(schemaTmpl))
		if tmpl.IsError() {
			return tmpl.Error()
		}
		if err := tmpl.MustGet().Execute(file.MustGet(), dbo); err != nil {
			return err
		}
		file.MustGet().Close()
	}
	return nil
}

func (dbo DBO) Columns(path string) error {
	fMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}
	for _, table := range dbo.Tables() {
		dir := filepath.Join(path, "columns", strings.ToLower(table.entity))
		os.MkdirAll(dir, 0755)
		file := mo.TupleToResult(os.Create(filepath.Join(dir, fmt.Sprintf("%s_columns.go", lo.SnakeCase(table.entity)))))
		if file.IsError() {
			return file.Error()
		}
		tmpl := mo.TupleToResult(template.New(table.entity).Funcs(fMap).Parse(columnTmpl))
		if tmpl.IsError() {
			return tmpl.Error()
		}
		if err := tmpl.MustGet().Execute(file.MustGet(), table); err != nil {
			return err
		}
		file.MustGet().Close()
	}
	return nil
}

// Build DBO object for the project
func Build() mo.Result[DBO] {
	dag := graph.New[string, Table](func(table Table) string {
		return table.entity
	}, graph.Directed())
	cfg := &packages.Config{Mode: packages.LoadSyntax, Dir: app.RootDir()}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return mo.Err[DBO](err)
	}
	//var root string
	var iEntity *types.Interface
	pkgs = lo.Filter(pkgs, func(pkg *packages.Package, index int) bool {
		basePkg, ok := pkg.Imports[baseEntity]
		if ok && iEntity == nil {
			typPkg := basePkg.Types
			scope := typPkg.Scope()
			object := scope.Lookup("IEntity")
			if object != nil {
				iEntity, _ = object.Type().Underlying().(*types.Interface)
			}
		}
		return ok
	})
	if len(pkgs) == 0 {
		return mo.Err[DBO](fmt.Errorf("no entities found"))
	}
	for _, pkg := range pkgs {
		for _, syntax := range pkg.Syntax {
			ast.Inspect(syntax, func(node ast.Node) bool {
				if funcDecl, ok := node.(*ast.FuncDecl); ok && funcDecl.Recv != nil {
					if obj := pkg.TypesInfo.Defs[funcDecl.Name]; obj != nil && obj.Name() == "Table" {
						method, _ := obj.(*types.Func)
						receiver := method.Type().(*types.Signature).Recv().Type()
						var named *types.Named
						switch t := receiver.(type) {
						case *types.Named:
							named = t
						case *types.Pointer:
							named = t.Elem().(*types.Named)
						}
						if named.Obj().Exported() && implements(named, iEntity) {
							if str, ok := named.Underlying().(*types.Struct); ok {
								columns := parseColumn(str, iEntity)
								if columns.IsError() {
									err = fmt.Errorf("type %s: %s", named.Obj().Name(), columns.Error().Error())
									return false
								}
								for _, stmt := range funcDecl.Body.List {
									if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
										if len(retStmt.Results) > 0 {
											dag.AddVertex(
												Table{entity: named.Obj().Name(),
													name:    exprToString(retStmt.Results[len(retStmt.Results)-1]),
													pkg:     pkg,
													columns: columns.MustGet()})
											break
										}
									}
								}
								//@todo pk must exists
							}
						}
					}
				}
				return true
			})
			if err != nil {
				return mo.Err[DBO](err)
			}
		}
	}
	return build(dag)
}

func build(g graph.Graph[string, Table]) mo.Result[DBO] {
	// build edge
	for _, entity := range lo.Keys(mo.TupleToResult(g.PredecessorMap()).MustGet()) {
		t := mo.TupleToResult(g.Vertex(entity)).MustGet()
		for _, c := range t.columns {
			if c.Ref().IsPresent() {
				referred := strings.Split(c.Ref().MustGet(), ".")
				// check reference format
				if len(referred) != 2 {
					return mo.Err[DBO](fmt.Errorf("invalid column reference: %s", entity))
				}
				// referenced table must be there
				rt := mo.TupleToResult(g.Vertex(referred[0]))
				if rt.IsError() {
					return mo.Err[DBO](fmt.Errorf("%s can not find reference type: %s", entity, referred[0]))
				}
				// check existence of the attribute
				rc := rt.MustGet().Column(referred[1])
				if rc.IsAbsent() {
					return mo.Err[DBO](fmt.Errorf("can not find attribute %s in %s", referred[1], referred[0]))
				}
				// check type of the attribute
				rTyp := rc.MustGet()
				if rTyp.B != c.B && strings.HasSuffix(c.B, rTyp.B) && strings.HasSuffix(rTyp.B, c.B) {
					return mo.Err[DBO](fmt.Errorf("type of %s.%s and %s is different", entity, c.A, c.Ref().MustGet()))
				}
				// check precision, should do it later
				//if c.Precision() != rc.MustGet().Precision() {
				//	return mo.Err[DBX](fmt.Errorf("precision of %s.%s and % is different", entity, c.A, c.Ref().MustGet()))
				//}
				g.AddEdge(entity, referred[0], graph.EdgeAttribute("ref", fmt.Sprintf("%s.%s -> %s.%s", t.name, c.A, rt.MustGet().name, rc.MustGet().A)))
			}
		}
	}
	return mo.Ok[DBO](DBO{g: g})
}

func basicType(typ types.Type) bool {
	if ts := typ.String(); strings.HasPrefix(ts, "database/sql.Null") || ts == "time.Time" {
		return true
	}
	switch t := typ.(type) {
	case *types.Basic:
		return true
	case *types.Pointer:
		return basicType(t.Elem())
	}
	return false
}

func parseColumn(str *types.Struct, inter *types.Interface) mo.Result[[]Column] {
	// 1: can not have no-builtin type, if it has, it must be embedded
	var columns []Column
	for i := range str.NumFields() {
		if f := str.Field(i); f.Exported() && f.IsField() {
			if implements(f.Type(), inter) {
				return mo.Err[[]Column](fmt.Errorf("%s is entity type", f.Name()))
			}
			if !basicType(f.Type()) {
				if !f.Embedded() {
					fmt.Printf("%s is a basic type %s\n", f.Name(), f.Type().String())
					return mo.Err[[]Column](fmt.Errorf("%s is not a basic type", f.Name()))
				} else {
					if cStr, ok := f.Type().Underlying().(*types.Struct); ok {
						if child := parseColumn(cStr, inter); child.IsOk() {
							columns = append(columns, child.MustGet()...)
						} else {
							return mo.Err[[]Column](child.Error())
						}
					}
				}
			} else {
				if matched := dbReg.FindStringSubmatch(str.Tag(i)); len(matched) > 0 {
					c := Column{A: f.Name(), B: f.Type().String(), C: matched[1]}
					if c.Property(colName).IsAbsent() {
						return mo.Err[[]Column](fmt.Errorf("no column definition for %s", f.Name()))
					}
					columns = append(columns, c)
				}
			}
		}
	}
	return lo.If(len(columns) > 0, mo.Ok[[]Column](columns)).Else(mo.Err[[]Column](fmt.Errorf("no columns found")))
}

// implements Function to check if a type implements an interface
func implements(t types.Type, inter *types.Interface) bool {
	return types.Implements(t, inter) || types.Implements(types.NewPointer(t), inter)
}

// Helper function to convert AST expression to string
func exprToString(expr ast.Expr) string {
	var sb strings.Builder
	err := format.Node(&sb, token.NewFileSet(), expr)
	if err != nil {
		log.Fatalf("Could not format expression: %v", err)
	}
	return sb.String()
}
