package uploadFile

import (
	"cgm_manager/utils"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	//uploadFile包使用方法

	savePath, err := utils.Mkdir("upload")
	if err != nil {
		fmt.Println(err.Error())
	}
	// 单文件上传，初始化时要求传入文件保存路径及上传表单的name
	fileUploader := NewFileUploader(Config{
		SavePath: savePath,
		FormName: "file",
	})

	//单文件上传，初始化时要求传入文件保存路径及上传表单的name
	fileUploader = NewMultiFileUploader(Config{
		SavePath: savePath,
		FormName: "file", //多文件上传时，客户端文件表单名应使用数组方式例如:file[]
	})
	//调用核心上传方法
	c := &gin.Context{}
	err = fileUploader.Upload(c) //依赖gin的context对象
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(fileUploader.Info)

	//执行Upload方法后如果成功，可以从File.Info结构体中获去相关，如果上传失败其错误信息也可以从File.Info.UploadErr中获取

}
