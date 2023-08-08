package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func convertGoFileToProto(filePath string) string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var protoDefinitions []string
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			protoMessage := generateProtoMessage(typeSpec.Name.Name, structType)
			protoDefinitions = append(protoDefinitions, protoMessage)
		}
	}

	return strings.Join(protoDefinitions, "\n\n")
}

func generateProtoMessage(structName string, structType *ast.StructType) string {
	var fields []string

	for i, field := range structType.Fields.List {
		fieldType := field.Type
		fieldName := getProtoFieldName(field)

		switch actualType := fieldType.(type) {
		case *ast.Ident: // Basic types and other structs
			protoType, exists := goTypeToProtoType[actualType.Name]
			if !exists {
				// If the type is not mapped, it defaults to a Go type (this may not be a good practice, but it's handled this way for now for the sake of code simplicity)
				protoType = actualType.Name
			}
			fields = append(fields, fmt.Sprintf("  %s %s = %d;", protoType, fieldName, i+1))
		case *ast.ArrayType: // Arrays and slices
			elementType := actualType.Elt
			if ident, ok := elementType.(*ast.Ident); ok {
				protoType, exists := goTypeToProtoType[ident.Name]
				if !exists {
					protoType = ident.Name
				}
				fields = append(fields, fmt.Sprintf("  repeated %s %s = %d;", protoType, fieldName, i+1))
			}
		case *ast.MapType: // Map type
			keyType := actualType.Key.(*ast.Ident).Name
			protoKeyType, exists := goTypeToProtoType[keyType]
			if !exists {
				protoKeyType = keyType
			}

			valueType := actualType.Value.(*ast.Ident).Name
			protoValueType, exists := goTypeToProtoType[valueType]
			if !exists {
				protoValueType = valueType
			}
			fields = append(fields, fmt.Sprintf("  map<%s, %s> %s = %d;", protoKeyType, protoValueType, fieldName, i+1))
		}
	}

	return fmt.Sprintf("message %s {\n%s\n}", structName, strings.Join(fields, "\n"))
}

func getProtoFieldName(field *ast.Field) string {
	if field != nil && len(field.Names) > 0 {
		name := field.Names[0]
		tag := field.Tag.Value
		jsonTag := getJsonTag(tag)
		if jsonTag != "" {
			return jsonTag
		}
		return name.Name
	}
	return ""
}

func getJsonTag(tag string) string {
	tag = strings.Trim(tag, "`")
	tags := strings.Split(tag, " ")
	for _, t := range tags {
		if strings.HasPrefix(t, "json:") {
			jsonTag := strings.Trim(t[len("json:"):], "\"")
			if parts := strings.Split(jsonTag, ","); len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}

var goTypeToProtoType = map[string]string{
	"int8":    "int32",
	"int":     "int32",
	"int32":   "int32",
	"int64":   "int64",
	"float32": "float",
	"float64": "double",
	"string":  "string",
	"bool":    "bool",
	"byte":    "bytes",
	"any":     "bytes",
	// ...
}
