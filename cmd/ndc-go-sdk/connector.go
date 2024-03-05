package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
)

const (
	connectorOutputFile   = "connector.generated.go"
	schemaOutputFile      = "schema.generated.json"
	typeMethodsOutputFile = "types.generated.go"
	textBlockErrorCheck   = `
  if err != nil {
    return err
  }
`
	textBlockErrorCheck2 = `
    if err != nil {
      return nil, err
    }
`
)

type connectorTypeBuilder struct {
	packageName string
	imports     map[string]string
	builder     *strings.Builder
}

func (ctb connectorTypeBuilder) String() string {
	var bs strings.Builder
	bs.WriteString(genFileHeader(ctb.packageName))
	if len(ctb.imports) > 0 {
		bs.WriteString("import (\n")
		for pkg, alias := range ctb.imports {
			if alias != "" {
				alias = alias + " "
			}
			bs.WriteString(fmt.Sprintf("  %s\"%s\"\n", alias, pkg))
		}
		bs.WriteString(")\n")
	}
	_, _ = bs.WriteString(ctb.builder.String())
	return bs.String()
}

type connectorGenerator struct {
	basePath     string
	moduleName   string
	rawSchema    *RawConnectorSchema
	typeBuilders map[string]*connectorTypeBuilder
}

func NewConnectorGenerator(basePath string, moduleName string, rawSchema *RawConnectorSchema) *connectorGenerator {
	return &connectorGenerator{
		basePath:     basePath,
		moduleName:   moduleName,
		rawSchema:    rawSchema,
		typeBuilders: make(map[string]*connectorTypeBuilder),
	}
}

func parseAndGenerateConnector(basePath string, directories []string, moduleName string) error {
	if err := os.Chdir(basePath); err != nil {
		return err
	}

	sm, err := parseRawConnectorSchemaFromGoCode(moduleName, ".", directories)
	if err != nil {
		return err
	}

	connectorGen := NewConnectorGenerator(basePath, moduleName, sm)
	return connectorGen.generateConnector(".")
}

func (cg *connectorGenerator) generateConnector(srcPath string) error {
	// generate schema.generated.json
	schemaBytes, err := json.MarshalIndent(cg.rawSchema.Schema(), "", "  ")
	if err != nil {
		return err
	}

	schemaPath := path.Join(srcPath, schemaOutputFile)
	if err := os.WriteFile(schemaPath, schemaBytes, 0644); err != nil {
		return err
	}

	targetPath := path.Join(srcPath, connectorOutputFile)
	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	w := bufio.NewWriter(f)
	defer func() {
		_ = w.Flush()
	}()

	if err := cg.genConnectorCodeFromTemplate(w); err != nil {
		return err
	}

	if err := cg.genTypeMethods(); err != nil {
		return err
	}

	return nil
}

func (cg *connectorGenerator) genConnectorCodeFromTemplate(w io.Writer) error {
	importLines := []string{}
	for importPath := range cg.rawSchema.Imports {
		importLines = append(importLines, fmt.Sprintf(`"%s"`, importPath))
	}

	return connectorTemplate.Execute(w, map[string]any{
		"Imports":    strings.Join(importLines, "\n"),
		"Module":     cg.moduleName,
		"Queries":    genConnectorFunctions(cg.rawSchema),
		"Procedures": genConnectorProcedures(cg.rawSchema),
	})
}

func genConnectorFunctions(rawSchema *RawConnectorSchema) string {
	if len(rawSchema.Functions) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, fn := range rawSchema.Functions {
		var argumentParamStr string
		_, _ = sb.WriteString(fmt.Sprintf("  case \"%s\":", fn.Name))
		if fn.ResultType.IsScalar {
			_, _ = sb.WriteString(`
    if len(request.Query.Fields) > 0 {
      return nil, schema.BadRequestError("cannot evaluate selection fields for scalar", nil)
    }`)
		}
		if fn.ArgumentsType != "" {
			argumentStr := fmt.Sprintf(`
    rawArgs, err := utils.ResolveArgumentVariables(request.Arguments, variables)
    if err != nil {
      return nil, schema.BadRequestError("failed to resolve argument variables", map[string]any{
        "cause": err.Error(),
      })
    }
    
    var args %s.%s
    if err = args.FromValue(rawArgs); err != nil {
      return nil, schema.BadRequestError("failed to resolve arguments", map[string]any{
        "cause": err.Error(),
      })
    }`, fn.PackageName, fn.ArgumentsType)
			_, _ = sb.WriteString(argumentStr)
			argumentParamStr = ", &args"
		}
		if fn.ResultType.IsScalar {
			_, _ = sb.WriteString(fmt.Sprintf("\n    return %s.%s(ctx, state%s)\n", fn.PackageName, fn.OriginName, argumentParamStr))
			continue
		}
		_, _ = sb.WriteString(fmt.Sprintf("\n    rawResult, err := %s.%s(ctx, state%s)", fn.PackageName, fn.OriginName, argumentParamStr))
		genGeneralOperationResult(&sb, fn.ResultType)

		if fn.ResultType.IsArray {
			_, _ = sb.WriteString("\n    result, err := utils.EncodeObjectsWithColumnSelection(request.Query.Fields, rawResult)")
		} else {
			_, _ = sb.WriteString("\n    result, err := utils.EncodeObjectWithColumnSelection(request.Query.Fields, rawResult)")
		}
		_, _ = sb.WriteString(textBlockErrorCheck2)
		_, _ = sb.WriteString("    return result, nil\n")

	}

	return sb.String()
}

func genGeneralOperationResult(sb *strings.Builder, resultType *TypeInfo) {
	sb.WriteString(textBlockErrorCheck2)
	if resultType.IsNullable {
		_, _ = sb.WriteString(`
    if rawResult == nil {
      return nil, nil
    }
`)
	} else {
		_, _ = sb.WriteString(`
    if rawResult == nil {
      return nil, schema.BadRequestError("expected not null result", nil)
    }
`)
	}
}

func genConnectorProcedures(rawSchema *RawConnectorSchema) string {
	if len(rawSchema.Procedures) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, fn := range rawSchema.Procedures {
		var argumentParamStr string
		_, _ = sb.WriteString(fmt.Sprintf("  case \"%s\":", fn.Name))
		if fn.ResultType.IsScalar {
			_, _ = sb.WriteString(`
    if len(operation.Fields) > 0 {
      return nil, schema.BadRequestError("cannot evaluate selection fields for scalar", nil)
    }`)
		} else if fn.ResultType.IsArray {
			_, _ = sb.WriteString(`
    selection, err := operation.Fields.AsArray()
    if err != nil {
      return nil, schema.BadRequestError("the selection field type must be array", map[string]any{
        "cause": err.Error(),
      })
    }`)
		} else {
			_, _ = sb.WriteString(`
    selection, err := operation.Fields.AsObject()
    if err != nil {
      return nil, schema.BadRequestError("the selection field type must be object", map[string]any{
        "cause": err.Error(),
      })
    }`)
		}
		if fn.ArgumentsType != "" {
			argumentStr := fmt.Sprintf(`
    var args %s.%s
    if err := json.Unmarshal(operation.Arguments, &args); err != nil {
      return nil, schema.BadRequestError("failed to decode arguments", map[string]any{
        "cause": err.Error(),
      })
    }`, fn.PackageName, fn.ArgumentsType)
			_, _ = sb.WriteString(argumentStr)
			argumentParamStr = ", &args"
		}

		if fn.ResultType.IsScalar {
			_, _ = sb.WriteString(fmt.Sprintf(`
    var err error
    result, err = %s.%s(ctx, state%s)`, fn.PackageName, fn.OriginName, argumentParamStr))
		} else {
			_, _ = sb.WriteString(fmt.Sprintf("\n    rawResult, err := %s.%s(ctx, state%s)\n", fn.PackageName, fn.OriginName, argumentParamStr))
			genGeneralOperationResult(&sb, fn.ResultType)

			if fn.ResultType.IsArray {
				_, _ = sb.WriteString("\n    result, err = utils.EvalNestedColumnArrayIntoSlice(selection, rawResult)\n")
			} else {
				_, _ = sb.WriteString("\n    result, err = utils.EvalNestedColumnObject(selection, rawResult)\n")
			}
		}

		_, _ = sb.WriteString(textBlockErrorCheck2)
	}

	return sb.String()
}

func (cg *connectorGenerator) genTypeMethods() error {
	if err := cg.genFunctionArgumentConstructors(); err != nil {
		return err
	}
	if err := cg.genObjectMethods(); err != nil {
		return err
	}
	if err := cg.genCustomScalarMethods(); err != nil {
		return err
	}
	for folderPath, builder := range cg.typeBuilders {
		schemaPath := path.Join(cg.basePath, folderPath, typeMethodsOutputFile)
		if err := os.WriteFile(schemaPath, []byte(builder.String()), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (cg *connectorGenerator) genObjectMethods() error {
	if len(cg.rawSchema.Objects) == 0 {
		return nil
	}

	objectKeys := getSortedKeys(cg.rawSchema.Objects)

	for _, objectName := range objectKeys {
		object := cg.rawSchema.Objects[objectName]
		if object.IsAnonymous {
			continue
		}
		sb := cg.getTypeBuilder(object.PackageName, object.PackageName)
		_, _ = sb.builder.WriteString(fmt.Sprintf(`
// ToMap encodes the struct to a value map
func (j %s) ToMap() map[string]any {
`, objectName))
		lines := cg.genObjectToMap(object, "j", "result", false, false)
		sb.builder.WriteString(strings.Join(lines, "\n"))
		sb.builder.WriteString(`
	return result
}`)
	}

	return nil
}

func (cg *connectorGenerator) genObjectToMap(object *ObjectInfo, selector string, name string, nullable bool, isArray bool) []string {

	fieldKeys := getSortedKeys(object.Fields)
	var lines []string
	if isArray {
		lines = []string{fmt.Sprintf("  var %s []map[string]any", name)}
		if nullable {
			lines = append(lines, fmt.Sprintf("  if %s != nil {", selector))
			selector = fmt.Sprintf("*%s", selector)
		}
		lines = append(lines, fmt.Sprintf("  %s = make([]map[string]any, len(%s))", name, selector))
		lines = append(lines, fmt.Sprintf("  for i, _item := range %s {", selector))
		loopLines := cg.genObjectToMap(object, "_item", "item", false, false)
		lines = append(lines, loopLines...)
		lines = append(lines, fmt.Sprintf("    %s[i] = item", name))
		lines = append(lines, "  }")
		if nullable {
			lines = append(lines, "}")
		}
		return lines
	}

	if nullable {
		lines = []string{fmt.Sprintf(`  
	var %s map[string]any
	if %s != nil {
		%s = map[string]any{`, name, selector, name)}
	} else {
		lines = []string{fmt.Sprintf("  %s := map[string]any{", name)}
	}

	for _, fieldKey := range fieldKeys {
		field := object.Fields[fieldKey]
		if field.Type.IsScalar {
			lines = append(lines, fmt.Sprintf("    \"%s\": %s.%s,", field.Key, selector, field.Name))
			continue
		}
		innerObject, ok := cg.rawSchema.Objects[field.Type.Name]
		if !ok {
			lines = append(lines, fmt.Sprintf("    \"%s\": %s.%s,", field.Key, selector, field.Name))
			continue
		}
		if !innerObject.IsAnonymous {
			if field.Type.IsArray {
				if field.Type.IsNullable {
					lines = append(lines, fmt.Sprintf("    \"%s\": utils.EncodeNullableMaps(%s.%s),", field.Key, selector, field.Name))
				} else {
					lines = append(lines, fmt.Sprintf("    \"%s\": utils.EncodeMaps(%s.%s),", field.Key, selector, field.Name))
				}
			} else {
				lines = append(lines, fmt.Sprintf("    \"%s\": utils.EncodeMap(%s.%s),", field.Key, selector, field.Name))
			}
			continue
		}
		varName := fmt.Sprintf("%s_%s", name, fieldKey)
		childLines := cg.genObjectToMap(innerObject, fmt.Sprintf("%s.%s", selector, field.Name), varName, field.Type.IsNullable, field.Type.IsArray)
		lines = append(childLines, lines...)
		lines = append(lines, fmt.Sprintf("    \"%s\": %s,", field.Key, varName))
	}
	lines = append(lines, "  }")
	if nullable {
		lines = append(lines, "  }")
	}
	return lines
}

// generate Scalar implementation for custom scalar types
func (cg *connectorGenerator) genCustomScalarMethods() error {
	if len(cg.rawSchema.CustomScalars) == 0 {
		return nil
	}

	scalarKeys := getSortedKeys(cg.rawSchema.CustomScalars)

	for _, scalarKey := range scalarKeys {
		scalar := cg.rawSchema.CustomScalars[scalarKey]
		sb := cg.getTypeBuilder(scalar.PackageName, scalar.PackageName)
		_, _ = sb.builder.WriteString(fmt.Sprintf(`
// ScalarName get the schema name of the scalar
func (j %s) ScalarName() string {
  return "%s"
}
`, scalar.Name, scalar.SchemaName))
	}
	return nil
}

func (cg *connectorGenerator) genFunctionArgumentConstructors() error {
	if len(cg.rawSchema.Functions) == 0 {
		return nil
	}

	for _, fn := range cg.rawSchema.Functions {
		if len(fn.Arguments) == 0 {
			continue
		}
		sb := cg.getTypeBuilder(fn.PackageName, fn.PackageName)
		_, _ = sb.builder.WriteString(fmt.Sprintf(`
// FromValue decodes values from map
func (j *%s) FromValue(input map[string]any) error {
  var err error
`, fn.ArgumentsType))

		argumentKeys := getSortedKeys(fn.Arguments)
		for _, key := range argumentKeys {
			arg := fn.Arguments[key]
			cg.genGetTypeValueDecoder(sb, arg.Type, key, arg.FieldName)
		}
		sb.builder.WriteString(`  return nil
}`)
	}

	return nil
}

func (cg *connectorGenerator) getTypeBuilder(fileName string, packageName string) *connectorTypeBuilder {
	bs, ok := cg.typeBuilders[fileName]
	if !ok {
		bs = &connectorTypeBuilder{
			packageName: packageName,
			imports: map[string]string{
				"github.com/hasura/ndc-sdk-go/utils": "",
			},
			builder: &strings.Builder{},
		}
		cg.typeBuilders[fileName] = bs
	}
	return bs
}

func genFileHeader(packageName string) string {
	return fmt.Sprintf(`// Code generated by github.com/hasura/ndc-sdk-go/codegen, DO NOT EDIT.
package %s
`, packageName)
}

func (cg *connectorGenerator) genGetTypeValueDecoder(sb *connectorTypeBuilder, ty *TypeInfo, key string, fieldName string) {
	typeName := ty.TypeAST.String()
	if strings.Contains(typeName, "complex64") || strings.Contains(typeName, "complex128") {
		panic("unsupported complex64 and complex128")
	}

	switch typeName {
	case "bool":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetBool(input, "%s")`, fieldName, key))
	case "*bool":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableBool(input, "%s")`, fieldName, key))
	case "string":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetString(input, "%s")`, fieldName, key))
	case "*string":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableString(input, "%s")`, fieldName, key))
	case "int", "int8", "int16", "int32", "int64", "rune", "byte":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetInt[%s](input, "%s")`, fieldName, typeName, key))
	case "uint", "uint8", "uint16", "uint32", "uint64":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetUint[%s](input, "%s")`, fieldName, typeName, key))
	case "*int", "*int8", "*int16", "*int32", "*int64", "*rune", "*byte":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableInt[%s](input, "%s")`, fieldName, strings.TrimPrefix(typeName, "*"), key))
	case "*uint", "*uint8", "*uint16", "*uint32", "*uint64":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableUint[%s](input, "%s")`, fieldName, strings.TrimPrefix(typeName, "*"), key))
	case "float32", "float64":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetFloat[%s](input, "%s")`, fieldName, typeName, key))
	case "*float32", "*float64":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableFloat[%s](input, "%s")`, fieldName, strings.TrimPrefix(typeName, "*"), key))
	case "time.Time":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetDateTime(input, "%s")`, fieldName, key))
	case "*time.Time":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableDateTime(input, "%s")`, fieldName, key))
	case "time.Duration":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetDuration(input, "%s")`, fieldName, key))
	case "*time.Duration":
		_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s, err = utils.GetNullableDuration(input, "%s")`, fieldName, key))
	default:
		if ty.IsNullable {
			pkgName, tyName := extractPackageAndTypeName(typeName)
			if pkgName != "" {
				sb.imports[pkgName] = ""
			}
			_, _ = sb.builder.WriteString(fmt.Sprintf(`  j.%s = new(%s)
  err = utils.DecodeNullableObjectValue(j.%s, input, "%s")`, fieldName, tyName, fieldName, key))
		} else {
			_, _ = sb.builder.WriteString(fmt.Sprintf(`  err = utils.DecodeObjectValue(&j.%s, input, "%s")`, fieldName, key))
		}
	}
	_, _ = sb.builder.WriteString(textBlockErrorCheck)
}

func extractPackageAndTypeName(name string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(name, "*"), "/")
	typeName := parts[len(parts)-1]
	typeNameParts := strings.Split(typeName, ".")
	if len(typeNameParts) < 2 {
		return "", typeName
	}
	if len(parts) == 1 {
		return typeNameParts[0], typeName
	}

	return strings.Join(append(parts[:len(parts)-1], typeNameParts[0]), "/"), typeName
}

func getSortedKeys[V any](input map[string]V) []string {
	var results []string
	for key := range input {
		results = append(results, key)
	}
	sort.Strings(results)
	return results
}