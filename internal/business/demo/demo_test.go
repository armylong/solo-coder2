package demo

import (
	"context"
	"fmt"
	"testing"

	confLibrary "github.com/armylong/go-library/service/conf"
)

func TestDemoBusiness_SetMessage(t *testing.T) {
	res, err := DemoBusiness.SetMessage(context.Background(), "longlonglong2")
	fmt.Println(res, err)
}

func TestDemoBusiness_GetMessage(t *testing.T) {
	res, err := DemoBusiness.GetMessage(context.Background())
	fmt.Println(res, err)
}

func TestDemoBusiness_GetMessageList(t *testing.T) {
	gaoDeMapKey := confLibrary.GetString("gaode-map-ppz-web.key")
	fmt.Println(gaoDeMapKey)
}
