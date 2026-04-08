package example

import (
	"fmt"
	"testing"

	"github.com/fixkme/protoc-gen-gom/example/pbout/go/model"
)

func TestModel(t *testing.T) {
	m := model.NewMPlayerModel()
	m.SetPlayerId(10)
	act := model.NewMModelActivity()
	m.SetModelActivity2("myact", act)
	fmt.Println(m.String())
}
