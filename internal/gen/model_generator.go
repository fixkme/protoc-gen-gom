package gen

import (
	"fmt"
	"strings"

	"github.com/fixkme/protoc-gen-gom/internal/mlog"
	"github.com/fixkme/protoc-gen-gom/internal/pbext"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type fieldInfo struct {
	*protogen.Field
	mName                  string // fieldName
	mType                  string // map[int64]*datas.MXX
	pbType                 string // map[int64]*datas.PBXX
	mapKeyType, mapValType string // for map
	isMap                  bool   // map
	isMessage              bool   // message
}
type messageInfo struct {
	*protogen.Message
	pbName string //PBXX
	fields []*fieldInfo
}

type ModelGenerator struct {
	*protogen.Plugin
	RootModels    []*messageInfo
	FileModels    map[*protogen.File][]*messageInfo
	Models        map[string]*messageInfo
	GenerateFiles map[string]*protogen.GeneratedFile
}

func NewModelGenerator(gen *protogen.Plugin) *ModelGenerator {
	return &ModelGenerator{
		Plugin:        gen,
		RootModels:    make([]*messageInfo, 0),
		FileModels:    make(map[*protogen.File][]*messageInfo),
		Models:        make(map[string]*messageInfo),
		GenerateFiles: make(map[string]*protogen.GeneratedFile),
	}
}

func (m *ModelGenerator) GenerateFile() error {
	for _, f := range m.Plugin.Files {
		if f.Generate {
			m.checkFileHasModel(f)
		}
	}
	if len(m.RootModels) == 0 {
		// 没有定义model数据
		return nil
	}

	// 递归搜索所有model
	if err := m.searchSubModels(); err != nil {
		return err
	}
	// 改成Model相关名字
	for _, model := range m.Models {
		name := pbToMName(model.pbName) // 前缀M表示model
		//Info("rename model => %s: %s", model.pbName, name)
		model.GoIdent.GoName = name
	}
	for file, models := range m.FileModels {
		geneFileName := generateFileName(file)
		g := m.GenerateFiles[geneFileName]
		for _, model := range models {
			for _, field := range model.fields {
				goType, pointer := fieldGoType(g, file, field.Field)
				if pointer {
					goType = "*" + goType
				}
				field.mType = goType
				field.mName = bigToSmallCamel(field.GoName)
				if field.isMap {
					keyType, _ := fieldGoType(g, file, field.Message.Fields[0])
					valType, _ := fieldGoType(g, file, field.Message.Fields[1])
					field.mapKeyType = keyType
					field.mapValType = valType
				}
			}
		}
	}
	// 开始生成文件
	for file, models := range m.FileModels {
		m.GenerateFileModel(file, models)
	}

	return nil
}

func (m *ModelGenerator) GenerateFileModel(file *protogen.File, models []*messageInfo) {
	if len(models) == 0 {
		return
	}
	filename := generateFileName(file)
	//g := m.Plugin.NewGeneratedFile(filename, file.GoImportPath)
	g := m.GenerateFiles[filename]

	if !m.Plugin.InternalStripForEditionsDiff() {
		genGeneratedHeader(m.Plugin, g, file)
	}

	g.P("package ", file.GoPackageName)
	g.P()

	for _, message := range models {
		genMessage(g, file, message)
	}
}

// func (m *ModelGenerator) GenerateProtoExt() {
// 	// 生成collector接口
// 	g := m.Plugin.NewGeneratedFile("pbext/ICollector.go", pbextPkg)
// 	g.P("package pbext")
// 	g.P()
// 	g.P("type ICollector interface {")
// 	g.P("Collect(string, any, bool)")
// 	g.P("}")
// }

func (m *ModelGenerator) checkFileHasModel(file *protogen.File) bool {
	list := make([]*protogen.Message, 0)
	for _, message := range file.Messages {
		isModel := false
		if strings.Contains(message.Comments.Leading.String(), "@model") {
			isModel = true
		} else if proto.HasExtension(message.Desc.Options(), pbext.E_IsModel) {
			val := proto.GetExtension(message.Desc.Options(), pbext.E_IsModel)
			if val.(bool) {
				isModel = true
			}
		}
		if isModel {
			list = append(list, message)
		}
	}
	if len(list) > 0 {
		for _, message := range list {
			m.addModel(message, true)
		}
		return true
	}
	return false
}

func (m *ModelGenerator) addModel(message *protogen.Message, isRoot bool) bool {
	fullName := string(message.Desc.FullName())
	if _, ok := m.Models[fullName]; ok {
		return false
	}

	fileName := message.Location.SourceFile
	file, ok := m.Plugin.FilesByPath[fileName]
	if !ok {
		err := fmt.Errorf("protogen.File not found %v", fileName)
		panic(err)
	}

	geneFileName := generateFileName(file)
	geneFile, ok := m.GenerateFiles[geneFileName]
	if !ok {
		geneFile = m.Plugin.NewGeneratedFile(geneFileName, file.GoImportPath)
		m.GenerateFiles[geneFileName] = geneFile
	}

	model := &messageInfo{
		Message: message,
		pbName:  message.GoIdent.GoName,
		fields:  make([]*fieldInfo, len(message.Fields)),
	}
	for i, field := range message.Fields {
		goType, pointer := fieldGoType(geneFile, file, field)
		if pointer {
			goType = "*" + goType
		}
		model.fields[i] = &fieldInfo{
			Field:     field,
			pbType:    goType,
			isMap:     field.Desc.IsMap(),
			isMessage: field.Desc.Kind() == protoreflect.MessageKind && !field.Desc.IsMap(), // 纯message类型，后面判断是去掉map情况
		}
	}
	m.Models[fullName] = model
	m.FileModels[file] = append(m.FileModels[file], model)
	if isRoot {
		m.RootModels = append(m.RootModels, model)
	}
	mlog.Info("add model fullName: %v, file: %v", fullName, fileName)
	return true
}

func (m *ModelGenerator) isModelMessage(message *protogen.Message) bool {
	fullName := string(message.Desc.FullName())
	_, ex := m.Models[fullName]
	return ex
}

// 收集子model
func (m *ModelGenerator) searchSubModels() (err error) {
	for _, message := range m.RootModels {
		if err = m.checkMessageHasModel(message.Message); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelGenerator) checkMessageHasModel(message *protogen.Message) (err error) {
	for _, field := range message.Fields {
		if err = m.checkFieldHasModel(field); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelGenerator) checkFieldHasModel(field *protogen.Field) error {
	if field.Desc.IsWeak() {
		return nil
	}
	if field.Desc.IsList() {
		return fmt.Errorf("field %v: %v", field.Desc.FullName(), "list(repeated) cant used in model")
	}

	if field.Desc.Kind() == protoreflect.MessageKind {
		if field.Desc.IsMap() {
			valueField := field.Message.Fields[1]
			m.checkFieldHasModel(valueField)
		} else {
			mlog.Info("field is model: %v, %v", field.Desc.FullName(), field.Message.Location.SourceFile)
			m.addModel(field.Message, false)
			m.checkMessageHasModel(field.Message)
		}
	}

	return nil
}
