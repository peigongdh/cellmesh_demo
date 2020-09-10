package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/davyxu/cellmesh/discovery"
	"github.com/davyxu/protoplus/model"
	"github.com/davyxu/protoplus/msgidutil"
	"github.com/davyxu/protoplus/util"

	"github.com/davyxu/cellmesh_demo/table"
)

// 从Proto文件中获取路由信息
func GenRouteTable(dset *model.DescriptorSet) (ret *table.RouteTable) {

	ret = new(table.RouteTable)

	for _, d := range dset.Structs() {

		if d.TagValueString("RouteRule") != "" && d.TagValueString("Service") != "" {

			ret.Rule = append(ret.Rule, &table.RouteRule{
				MsgName: d.Name,
				SvcName: d.TagValueString("Service"),
				Mode:    d.TagValueString("RouteRule"),
				MsgID:   msgidutil.StructMsgID(d),
			})
		}
	}

	return
}

// 上传路由表到consul KV
func UploadRouteTable(tab *table.RouteTable) error {

	data, err := json.MarshalIndent(tab, "", "\t")

	if err != nil {
		return err
	}

	fmt.Printf("Write '%s'", *flagConfigPath)
	return discovery.Default.SetValue(*flagConfigPath, string(data))
}

var (
	flagConfigPath = flag.String("configpath", "config_demo/route_rule", "discovery kv config path")
)

var (
	flagPackage = flag.String("package", "", "package name in source files")
)

func main() {

	flag.Parse()

	discovery.Default = memsd.NewDiscovery(nil)

	dset := new(model.DescriptorSet)
	dset.PackageName = *flagPackage

	var routeTable *table.RouteTable

	err := util.ParseFileList(dset)

	if err != nil {
		goto OnError
	}

	routeTable = GenRouteTable(dset)

	err = UploadRouteTable(routeTable)

	if err != nil {
		goto OnError
	}

	return

OnError:
	fmt.Println(err)
	os.Exit(1)
}
