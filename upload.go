/*
*
介绍: uploadFile包基于gin的表单上传方法上提供一层对使用者友好的封装，主要特性为：支持单文件，多文件上传，文件大小，文件类型限制
作者: mike
邮箱: 614168741@qq.com
协议：本包采用GPL开源协议，在软件和软件的所有副本中都必须包含版权声明和许可声明。
*
*/
package uploadFile

import (
	"errors"
	"fmt"
	"gFile/utils"
	"github.com/gin-gonic/gin"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 * 10           //默认最大上传限制10M,单位是byte
const ENABLE_EXT_NAME = ".jpg,.png,.jpeg,xlsx,xls" //默认支持上传的文件后缀
var UnsupportedfileExtErr = errors.New("文件类型不支持")
var UploadFaildErr = errors.New("上传失败")
var FileSizeExceedsLimitErr = errors.New("文件大小超过限制")
var FileSavePathCannotBeEmpty = errors.New("文件保存路径不能为空")
var UploadFormNameCannotBeEmpty = errors.New("表单名配置不能为空")

type File struct {
	multipartFile bool //文件上传方式标识
	Config
	Info
	MultiInfo
}

// 上传配置
type Config struct {
	SavePath      string   //文件保存路径
	FormName      string   //上传表单名
	MaxUploadSize int64    //文件最大容量限制
	EnableExtName []string //文件类型限制
}

// 多文件上传后返回的信息
type MultiInfo []Info

// Info 单文件上传后返回的信息
type Info struct {
	RealPath       string    //上传文件真实存储位置
	Size           int64     //文件大小
	FileName       string    //文件名称
	CreateTime     time.Time //文件创建时间
	LastModifyTime time.Time //最后一次修改时间
	UploadErr      error     //上传过程中存在的错误
}

// 初始化上传对象，要求传入表单名称及文件保存位置
func _init(config Config) File {
	file := File{}
	file.Config = config
	return file
}

// 构造一个单文件上传器
func NewFileUploader(config Config) File {
	file := _init(config)
	file.multipartFile = false
	return file
}

// 构造一个多文件上传器
func NewMultiFileUploader(config Config) File {
	file := _init(config)
	file.multipartFile = true
	return file
}

// Upload 上传核心方法
func (file *File) Upload(c *gin.Context) error {
	if file.SavePath == "" { //判断保存路径
		return FileSavePathCannotBeEmpty
	}
	if file.FormName == "" { //判断上传表单名
		return UploadFormNameCannotBeEmpty
	}
	var fileExt string
	if file.multipartFile { //多文件上传逻辑块
		form, err := c.MultipartForm()
		if err != nil {
			return UploadFaildErr
		}
		files := form.File[file.FormName]
		for _, headers := range files { //检查错误
			fileName := headers.Filename
			fileSize := headers.Size
			if !file.SizeIs(fileSize) { //检查大小是否合规
				file.UploadErr = FileSizeExceedsLimitErr
				file.UploadErr = fmt.Errorf("文件名:%v %w", fileName, file.UploadErr)
			}
			fileExt = path.Ext(fileName)
			if !file.extIs(fileExt) { //检查文件类型是否合规
				file.UploadErr = UnsupportedfileExtErr
				file.UploadErr = fmt.Errorf("文件名:%v %w", fileName, file.UploadErr)
			}
		}
		if errors.Is(file.UploadErr, FileSizeExceedsLimitErr) || errors.Is(file.UploadErr, UnsupportedfileExtErr) {
			return file.UploadErr //任意一文件报错，即整体报错
		}
		for _, _headers := range files { //如果每个文件都合规则保存，要求一致性即不允许单个文件成功或失败
			fileNameInt := time.Now().Unix()
			fileNameStr := fmt.Sprintf("%v%v", strconv.FormatInt(fileNameInt, 10), utils.RandString(8, []rune{}))
			fileExt = path.Ext(_headers.Filename)
			fileName := fileNameStr + fileExt //文件名由时间戳生成，避免用户上传相同文件名文件造成覆盖丢失
			realPath := filepath.Join(file.SavePath, "/", fileName)
			err = c.SaveUploadedFile(_headers, realPath)
			if err != nil {
				return err
			}
			fileStat, err := os.Stat(realPath)
			if err != nil {
				return err
			}
			info := Info{}
			//上传成功保存文件信息
			info.RealPath = realPath
			info.Size = fileStat.Size()
			info.CreateTime = time.Now()
			info.FileName = fileStat.Name()
			info.LastModifyTime = fileStat.ModTime()
			infoArr := []Info{info}
			file.MultiInfo = infoArr
		}
		return nil
	} else { //单文件上传逻辑块
		_, headers, err := c.Request.FormFile(file.FormName)
		if err != nil {
			return UploadFaildErr
		}
		fileSize := headers.Size
		fileNameInt := time.Now().Unix()
		fileNameStr := fmt.Sprintf("%v%v", strconv.FormatInt(fileNameInt, 10), utils.RandString(8, []rune{}))
		fileExt = path.Ext(headers.Filename)
		fileName := fileNameStr + fileExt //文件名由时间戳生成，避免用户上传相同文件名的文件造成覆盖丢失
		if !file.SizeIs(fileSize) {       //检查大小是否合规
			file.UploadErr = FileSizeExceedsLimitErr
			return FileSizeExceedsLimitErr
		}
		if !file.extIs(fileExt) { //检查文件类型是否合规
			file.UploadErr = UnsupportedfileExtErr
			return UnsupportedfileExtErr
		}
		realPath := filepath.Join(file.SavePath, "/", fileName)
		err = c.SaveUploadedFile(headers, realPath)
		if err != nil {
			return err
		}
		fileStat, err := os.Stat(realPath)
		if err != nil {
			return err
		}
		//上传成功保存文件信息
		file.Info.RealPath = realPath
		file.Info.Size = fileStat.Size()
		file.Info.CreateTime = time.Now()
		file.Info.FileName = fileStat.Name()
		file.Info.LastModifyTime = fileStat.ModTime()
	}
	return nil

}

// 上传扩展校验
func (file File) extIs(fileExt string) bool {
	extMatch := func(fileExt string, EnableExtName []string) (enable bool) {
		enable = false

		for _, ext := range EnableExtName {
			if ext == fileExt {
				enable = true
				break
			}
		}
		return
	}
	defaultEnableExtName := strings.FieldsFunc(ENABLE_EXT_NAME, func(r rune) bool {
		return r == ','
	})
	//先从用户配置的扩展解析，再从默认的扩展解析，两者满足其一即可
	if len(file.EnableExtName) != 0 {
		return extMatch(fileExt, file.EnableExtName)
	} else {
		return extMatch(fileExt, defaultEnableExtName)
	}
}

// 上传文件大小校验
func (file File) SizeIs(fileSize int64) bool {
	//先从用户配置解析，如果未配置则使用默认配置
	if file.MaxUploadSize != 0 {
		return fileSize < file.MaxUploadSize
	} else {
		return fileSize < MAX_UPLOAD_SIZE
	}
}
