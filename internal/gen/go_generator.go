package gen

import (
	"strconv"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

var (
	pbextPkg   = protogen.GoImportPath("xx/pbext")
	reflectPkg = protogen.GoImportPath("reflect")
	fmtPkg     = protogen.GoImportPath("fmt")
	stringsPkg = protogen.GoImportPath("strings")
)

func InitGoModuleName(name string) {
	pbextPkg = protogen.GoImportPath(name + "/pbext")
}

func generateFileName(file *protogen.File) string {
	return file.GeneratedFilenamePrefix + ".pc.go"
}

func genMessage(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo) {
	leadingComments := appendDeprecationSuffix(m.Comments.Leading,
		m.Desc.ParentFile(),
		m.Desc.Options().(*descriptorpb.MessageOptions).GetDeprecated())
	g.P(leadingComments,
		"type ", m.GoIdent, " struct {")
	genMessageFields(g, f, m)
	g.P("}")
	g.P()

	genMessageMethods(g, f, m)
}

func genMessageFields(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo) {
	for _, field := range m.fields {
		g.P(field.Comments.Leading, field.mName, " ", field.mType, field.Comments.Trailing)
	}
	g.P()
	g.P("// 自己本身的同步key,由父对象指定")
	g.P("selfSyncID string")
	g.P("// 本对象所有属性的同步key数组")
	g.P("fieldSyncIDs [", len(m.Fields), "]string")
	g.P("// 收集字典,每帧清空同步")
	g.P("collector ", pbextPkg.Ident("ICollector"))
	g.P("// 监测变化回调")
	g.P("changedCb func(string)")
	g.P()
}

func genMessageMethods(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo) {
	genMessageCommonMethods(g, f, m)
	for i, field := range m.fields {
		genMessageFieldMethods(g, f, m, field, i)
	}
}

func genMessageCommonMethods(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo) {
	// 构造函数
	g.P("// 构造函数")
	g.P("func New", m.GoIdent.GoName, "() *", m.GoIdent.GoName, " {")
	g.P("m := &", m.GoIdent.GoName, "{}")
	for _, field := range m.fields {
		if field.isMap {
			g.P("m.", field.mName, " = ", "make("+field.mType+")")
		} else if field.isMessage {
			pkg, mMsg := getFieldMessageGoIdent(g, field.Field)
			if pkg == "" {
				g.P("m.", field.mName, " = ", "New", mMsg, "()")
			} else {
				g.P("m.", field.mName, " = ", pkg, ".", "New", mMsg, "{}")
			}
		}
	}
	for i, field := range m.fields {
		g.P("m.fieldSyncIDs[", i, "]", " = ", strconv.Quote(JSONSnakeCase(field.mName)))
	}
	g.P("return m")
	g.P("}")
	g.P()

	// PB构造函数
	g.P("// PB构造函数")
	g.P("func New", m.pbName, "() *", m.pbName, " {")
	g.P("pb := &", m.pbName, "{}")
	for i, field := range m.fields {
		if field.isMap {
			g.P("pb.", m.Fields[i].GoName, " = ", "make("+field.pbType+")")
		} else if field.isMessage {
			pkg, mMsg := getFieldMessageGoIdent(g, field.Field)
			pbMsg := mToPBName(mMsg)
			if pkg == "" {
				g.P("pb.", m.Fields[i].GoName, " = ", "New", pbMsg, "()")
			} else {
				g.P("pb.", m.Fields[i].GoName, " = ", pkg, ".", "New", pbMsg, "{}")
			}
		}
	}
	g.P("return pb")
	g.P("}")
	g.P()

	// 设置collector函数
	g.P("// 设置collector函数")
	g.P("func (m *", m.GoIdent.GoName, ") SetCollector(syncID string, collector ", pbextPkg.Ident("ICollector"), ", cb func(string)) {")
	g.P("m.selfSyncID = syncID")
	g.P("m.collector = collector")
	g.P("m.changedCb = cb")
	g.P("if syncID != ", strconv.Quote(""), " {")
	g.P("syncID = syncID + ", strconv.Quote("."))
	g.P("}")
	for i, field := range m.fields {
		g.P("m.fieldSyncIDs[", i, "]", " = ", "syncID + ", strconv.Quote(JSONSnakeCase(field.mName)))
	}
	for i, field := range m.fields {
		if field.isMap && mapValueIsMessage(field.Field) {
			g.P("for key, value := range m.", field.mName, " {")
			g.P("syncKey := ", fmtPkg.Ident("Sprintf"), "(\"%s.%v\", m.fieldSyncIDs[", i, "], key)")
			g.P("value.SetCollector(syncKey, collector, cb)")
			g.P("}")
		} else if !field.isMap && field.isMessage {
			g.P("m.", field.mName, ".SetCollector(m.fieldSyncIDs[", i, "], collector, cb)")
		}
	}
	g.P("}")
	g.P()

	// 检查数值变化函数
	g.P("// 检查数值变化函数")
	g.P("func (m *", m.GoIdent.GoName, ") checkDirty(valueOld any, valueNew any, key string, ntfClient bool) bool {")
	g.P("if ", reflectPkg.Ident("DeepEqual"), "(valueOld, valueNew) {")
	g.P("return false")
	g.P("}")
	g.P("if m.collector != nil {")
	g.P("m.collector.Collect(key, valueNew, ntfClient)")
	g.P("}")
	g.P("if m.changedCb != nil {")
	g.P("m.changedCb(key)")
	g.P("}")
	g.P("return true")
	g.P("}")
	g.P()

	// ToPB函数
	g.P("// ToPB函数")
	g.P("func (m *", m.GoIdent.GoName, ") ToPB() *", m.pbName, " {")
	g.P("if m == nil {")
	g.P("return nil")
	g.P("}")
	g.P("pb := New", m.pbName, "()")
	for _, field := range m.fields {
		if field.isMap {
			g.P("for key, value := range m.", field.mName, " {")
			if mapValueIsMessage(field.Field) {
				g.P("pb.", field.GoName, "[key]", " = value.ToPB()")
			} else {
				g.P("pb.", field.GoName, "[key]", " = value")
			}
			g.P("}")
		} else if field.isMessage {
			g.P("pb.", field.GoName, " = m.", field.mName, ".ToPB()")
		} else {
			g.P("pb.", field.GoName, " = m.", field.mName)
		}
	}
	g.P("return pb")
	g.P("}")
	g.P()

	// InitFromPB函数
	g.P("// InitFromPB函数")
	g.P("func (m *", m.GoIdent.GoName, ") InitFromPB(pb *", m.pbName, ") {")
	g.P("if pb == nil {")
	g.P("return")
	g.P("}")
	for _, field := range m.fields {
		if field.isMap {
			if mapValueIsMessage(field.Field) {
				pkg, mMsg := getFieldMessageGoIdent(g, field.Field.Message.Fields[1])
				g.P("for key, value := range pb.", field.GoName, " {")
				if pkg == "" {
					g.P("v := ", "New", mMsg, "()")
				} else {
					g.P("v := ", pkg, ".New", mMsg, "()")
				}
				g.P("v.InitFromPB(value)")
				g.P("m.", field.mName, "[key] = v")
				g.P("}")
			} else {
				g.P("if pb.", field.GoName, " != nil {")
				g.P("m.", field.mName, " = ", "pb.", field.GoName)
				g.P("}")
			}
		} else if field.isMessage {
			g.P("m.", field.mName, ".InitFromPB(pb.", field.GoName, ")")
		} else {
			g.P("m.", field.mName, " = pb.", field.GoName)
		}
	}
	g.P("}")
	g.P()

	// String 函数
	g.P("// String 函数")
	g.P("func (m *", m.GoIdent.GoName, ") String() string {")
	g.P("var strBuilder ", stringsPkg.Ident("Builder"))
	//g.P("strBuilder.WriteString(", strconv.Quote(m.GoIdent.GoName), ")")
	g.P("strBuilder.WriteString(\"{\")")
	for i, field := range m.fields {
		g.P("strBuilder.WriteString(\"", field.mName, ":\")")
		if field.isMap {
			g.P("strBuilder.WriteString(\"{\")")
			if mapValueIsMessage(field.Field) {
				g.P("for key, value := range m.", field.mName, " {")
				g.P("strBuilder.WriteString(", fmtPkg.Ident("Sprintf"), "(\"%v:%s \", key, value.String())", ")")
				g.P("}")
			} else {
				g.P("for key, value := range m.", field.mName, " {")
				g.P("strBuilder.WriteString(", fmtPkg.Ident("Sprintf"), "(\"%v:%v \", key, value)", ")")
				g.P("}")
			}
			g.P("strBuilder.WriteString(\"}\")")
		} else if field.isMessage {
			g.P("strBuilder.WriteString(m.", field.mName, ".", "String()", ")")
		} else {
			g.P("strBuilder.WriteString(", fmtPkg.Ident("Sprintf"), "(\"%v\", m.", field.mName, ")", ")")
		}
		if i < len(m.fields)-1 {
			g.P("strBuilder.WriteString(\", \")")
		}
	}
	g.P("strBuilder.WriteString(\"}\")")
	g.P("return strBuilder.String()")
	g.P("}")
	g.P()
}

func genMessageFieldMethods(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo, idx int) {
	genMessageFieldGetter(g, f, m, field)
	genMessageFieldSetter(g, f, m, field, idx)
	genMessageFieldAddFunc(g, f, m, field, idx)
	if field.isMap {
		genMessageFieldMapRemoveFunc(g, f, m, field, idx)
		genMessageFieldMapClearFunc(g, f, m, field)
		genMessageFieldMapLenFunc(g, f, m, field)
		genMessageFieldMapRangeFunc(g, f, m, field)
	}
}

func genMessageFieldGetter(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo) {
	g.P(field.Comments.Leading)
	g.P(field.Comments.Trailing)
	if field.isMap {
		g.P("func (m *", m.GoIdent.GoName, ") Get", field.GoName, "(key ", field.mapKeyType, ") (", field.mapValType, ", bool) {")
		g.P("value, exists := m.", field.mName, "[key]")
		g.P("return value, exists")
		g.P("}")
	} else {
		g.P("func (m *", m.GoIdent.GoName, ") Get", field.GoName, "() ", field.mType, " {")
		g.P("return m.", field.mName)
		g.P("}")
	}
	g.P()
}

func genMessageFieldSetter(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo, idx int) {
	if field.isMap {
		g.P("func (m *", m.GoIdent.GoName, ") Set", field.GoName, "(key ", field.mapKeyType, ", value ", field.mapValType, ") {")
		g.P("localSyncKey := ", fmtPkg.Ident("Sprintf"), "(\"%s.%v\", m.fieldSyncIDs[", idx, "], key)")
		g.P("var oldValue any")
		g.P("if v, ok := m.", field.mName, "[key]; ok {")
		g.P("oldValue = v")
		g.P("} else {")
		g.P("oldValue = nil")
		g.P("}")
		if mapValueIsMessage(field.Field) {
			g.P("if m.checkDirty(oldValue, value, localSyncKey, true) {")
			g.P("value.SetCollector(localSyncKey, m.collector, m.changedCb)")
			g.P("}")
		} else {
			g.P("m.checkDirty(oldValue, value, localSyncKey, true)")
		}
		g.P("m.", field.mName, "[key] = value")
		g.P("}")
	} else {
		g.P("func (m *", m.GoIdent.GoName, ") Set", field.GoName, "(value ", field.mType, ") {")
		if field.isMessage {
			g.P("if m.checkDirty(m.", field.mName, ", value, m.fieldSyncIDs[", idx, "], true) {")
			g.P("value.SetCollector(m.fieldSyncIDs[", idx, "], m.collector, m.changedCb)")
			g.P("}")
		} else {
			g.P("m.checkDirty(m.", field.mName, ", value, m.fieldSyncIDs[", idx, "], true)")
		}
		g.P("m.", field.mName, " = value")
		g.P("}")
	}
	g.P()
}

func genMessageFieldAddFunc(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo, idx int) {
	if field.isMessage {
		return
	}
	if field.isMap {
		if mapValueIsMessage(field.Field) {
			return
		}
		// map value 为整型
		if strings.HasPrefix(field.mapValType, "int") {
			g.P("func (m *", m.GoIdent.GoName, ") Add", field.GoName, "(key ", field.mapKeyType, ", add ", field.mapValType, ") ", field.mapValType, " {")
			g.P("localSyncKey := ", fmtPkg.Ident("Sprintf"), "(\"%s.%v\", m.fieldSyncIDs[", idx, "], key)")
			g.P("var oldValue any")
			g.P("var newValue ", field.mapValType)
			g.P("if v, ok := m.", field.mName, "[key]; ok {")
			g.P("oldValue, newValue = v, add + v")
			g.P("} else {")
			g.P("oldValue, newValue = nil, add")
			g.P("}")
			g.P("m.checkDirty(oldValue, newValue, localSyncKey, true)")
			g.P("m.", field.mName, "[key] = newValue")
			g.P("return newValue")
			g.P("}")
			g.P()
		}
	} else if strings.HasPrefix(field.mType, "int") {
		g.P("func (m *", m.GoIdent.GoName, ") Add", field.GoName, "(add ", field.mType, ") ", field.mType, " {")
		g.P("oldValue := m.", field.mName)
		g.P("m.", field.mName, " += add")
		g.P("m.checkDirty(oldValue, m.", field.mName, ", m.fieldSyncIDs[", idx, "], true)")
		g.P("return m.", field.mName)
		g.P("}")
		g.P()
	}
}

func genMessageFieldMapRemoveFunc(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo, idx int) {
	if !field.isMap {
		return
	}
	g.P("func (m *", m.GoIdent.GoName, ") Remove", field.GoName, "(key ", field.mapKeyType, ") {")
	g.P("localSyncKey := ", fmtPkg.Ident("Sprintf"), "(\"%s.%v\", m.fieldSyncIDs[", idx, "], key)")
	g.P("m.checkDirty(m.", field.mName, "[key], \"__DELETE__\", localSyncKey, true)")
	g.P("delete(m.", field.mName, ", key)")
	g.P("}")
	g.P()
}

func genMessageFieldMapClearFunc(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo) {
	if !field.isMap {
		return
	}
	g.P("func (m *", m.GoIdent.GoName, ") Clear", field.GoName, "() {")
	g.P("for k := range m.", field.mName, " {")
	g.P("m.Remove", field.GoName, "(k)")
	g.P("}")
	g.P("}")
	g.P()
}

func genMessageFieldMapLenFunc(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo) {
	if !field.isMap {
		return
	}
	g.P("func (m *", m.GoIdent.GoName, ") Get", field.GoName, "Count() int64 {")
	g.P("return int64(len(m.", field.mName, "))")
	g.P("}")
	g.P()
}

func genMessageFieldMapRangeFunc(g *protogen.GeneratedFile, f *protogen.File, m *messageInfo, field *fieldInfo) {
	if !field.isMap {
		return
	}
	g.P(field.Comments.Leading, field.Comments.Trailing)
	g.P("func (m *", m.GoIdent.GoName, ") Range", field.GoName, "(f func(key ", field.mapKeyType, ", value ", field.mapValType, ") bool) {")
	g.P("for k, v := range m.", field.mName, " {")
	g.P("if !f(k, v) {")
	g.P("return")
	g.P("}")
	g.P("}")
	g.P("}")
	g.P()
}

func getFieldMessageGoIdent(g *protogen.GeneratedFile, field *protogen.Field) (pkgName, msgName string) {
	ident := g.QualifiedGoIdent(field.Message.GoIdent)
	if idx := strings.Index(ident, "."); idx > 0 {
		return ident[:idx], ident[idx+1:]
	} else {
		return "", ident
	}
}

func mapValueIsMessage(field *protogen.Field) bool {
	if !field.Desc.IsMap() {
		return false
	}
	mapValue := field.Message.Fields[1]
	return mapValue.Desc.Kind() == protoreflect.MessageKind && !mapValue.Desc.IsMap()
}
