package image

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/chai2010/webp"
	"github.com/dsoprea/go-exif"
	"github.com/nfnt/resize"

	"github.com/tea4go/gh/utils"
	"golang.org/x/image/bmp"
)

type TExif struct {
	CameraName  string // 相机厂商
	DeviceModel string // 设备型号
	PhotoTime   string // 拍摄时间
	PixelX      string // 图片尺寸X
	PixelY      string // 图片尺寸Y
}

func get_tag_bytes(ex *TExif) []byte {
	bs, _ := json.MarshalIndent(ex, "", "  ")
	var data []byte
	data = make([]byte, 1024-len(bs))

	return append(bs, data...)
}

type IfdEntry struct {
	IfdPath     string                `json:"ifd_path"`
	FqIfdPath   string                `json:"fq_ifd_path"`
	IfdIndex    int                   `json:"ifd_index"`
	TagId       uint16                `json:"tag_id"`
	TagName     string                `json:"tag_name"`
	TagTypeId   exif.TagTypePrimitive `json:"tag_type_id"`
	TagTypeName string                `json:"tag_type_name"`
	UnitCount   uint32                `json:"unit_count"`
	Value       interface{}           `json:"value"`
	ValueString string                `json:"value_string"`
}

func get_exif_value(entries []*IfdEntry, name string) string {
	for _, entry := range entries {
		if entry.TagName == name {
			return entry.ValueString
		}
	}
	return ""
}

func get_exif(ex *TExif, data []byte) error {
	rawExif, err := exif.SearchAndExtractExif(data)
	if err != nil {
		return err
	}

	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	entries := make([]*IfdEntry, 0)
	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (re_err error) {
		defer func() {
			if state := recover(); state != nil {
				re_err = state.(error)
				log.Printf("获取标签失败[%s-%d]，%s\n", fqIfdPath, tagId, re_err.Error())
			}
		}()
		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		if err != nil {
			log.Printf("获取标签路径失败[%s-%d]，%s\n", fqIfdPath, tagId, err.Error())
			return
		}

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			if strings.Contains(err.Error(), "tag not found") {
				//log.Printf("未知的标签[%s-%d]\n", ifdPath, tagId)
			} else {
				log.Printf("获取标签失败[%s-%d]，%s\n", ifdPath, tagId, err.Error())
			}
			return
		}
		valueString := ""
		var value interface{}
		if tagType.Type() == exif.TypeUndefined {
			value, err = valueContext.Undefined()
			if err != nil {
				if err == exif.ErrUnhandledUnknownTypedTag {
					//log.Printf("获取标签值失败[%s-%d]，未知的类型\n", ifdPath, tagId)
					value = nil
				} else {
					log.Printf("获取标签值失败[%s-%d]，%s\n", ifdPath, tagId, err.Error())
					return
				}
			}
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString, err = valueContext.FormatFirst()
			if err != nil {
				log.Printf("获取标签值失败[%s-%d]，%s\n", ifdPath, tagId, err.Error())
				return
			}

			value = valueString
		}

		entry := IfdEntry{
			IfdPath:     ifdPath,
			FqIfdPath:   fqIfdPath,
			IfdIndex:    ifdIndex,
			TagId:       tagId,
			TagName:     it.Name,
			TagTypeId:   tagType.Type(),
			TagTypeName: tagType.Name(),
			UnitCount:   valueContext.UnitCount(),
			Value:       value,
			ValueString: valueString,
		}
		entries = append(entries, &entry)

		re_err = nil
		return
	}

	exif.Visit(exif.IfdStandard, im, ti, rawExif, visitor)

	// Make: "Xiaomi"
	// Model: "Che1-CL1"
	// DateTime: "2015:05:09 16:21:08"
	// ImageWidth: 4208
	// ImageLength: 3120
	ex.CameraName = get_exif_value(entries, "Make")
	ex.DeviceModel = get_exif_value(entries, "Model")
	ex.PhotoTime = get_exif_value(entries, "DateTime")
	ex.PixelX = get_exif_value(entries, "ImageWidth")
	ex.PixelY = get_exif_value(entries, "ImageLength")

	return nil
}

func ImageConv(filePath, outPath, outMPath string, quality float32, desWidth int, desHeight int) error {
	if !utils.FileIsExist(filePath) {
		return fmt.Errorf("文件不存在！(%s)", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("打开文件报错(%s)，%s", filePath, err.Error())
	}

	var img image.Image
	var srcWidth int
	var srcHeight int

	reader := bytes.NewReader(data)
	reader2 := bytes.NewReader(data)
	var imgtype = http.DetectContentType(data[:512])
	if strings.Contains(imgtype, "jpeg") {
		img, _ = jpeg.Decode(reader)
		img2, _ := jpeg.DecodeConfig(reader2)
		srcWidth = img2.Width
		srcHeight = img2.Height
	} else if strings.Contains(imgtype, "png") {
		img, _ = png.Decode(reader)
		img2, _ := png.DecodeConfig(reader2)
		srcWidth = img2.Width
		srcHeight = img2.Height
	} else if strings.Contains(imgtype, "bmp") {
		img, _ = bmp.Decode(reader)
		img2, _ := bmp.DecodeConfig(reader2)
		srcWidth = img2.Width
		srcHeight = img2.Height
	} else {
		return fmt.Errorf("图像文件格式(%s)不支持！(%s)", imgtype, filePath)
	}
	if err != nil {
		return fmt.Errorf("解析图像文件报错(%s)，%s", filePath, err.Error())
	}

	var ex TExif
	if strings.Contains(imgtype, "jpeg") {
		err = get_exif(&ex, data)
		if err != nil && !errors.Is(err, exif.ErrNoExif) {
			return fmt.Errorf("解析图像Exif报错(%s)，%s", filePath, err.Error())
		}
	}
	//获取图片文件1024尾部字符
	apbuf := get_tag_bytes(&ex)

	var bufm bytes.Buffer
	var m image.Image
	if desWidth < srcWidth {
		ratio := math.Min(float64(desWidth)/float64(srcWidth), float64(desHeight)/float64(srcHeight))
		newWidth := uint(math.Ceil(float64(srcWidth) * ratio))
		newHeight := uint(math.Ceil(float64(srcHeight) * ratio))
		m = resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	} else {
		m = img
	}

	if err = webp.Encode(&bufm, m, &webp.Options{Lossless: false, Quality: 85}); err != nil {
		return fmt.Errorf("生成webp图像报错，%s", err.Error())
	}

	//生成webp文件路径
	dirM := utils.GetFileDir(outMPath)
	if !utils.FileIsExist(dirM) {
		if err = utils.Mkdir(dirM); err != nil {
			return fmt.Errorf("创建webp图像路径报错(%s)，%s", dirM, err.Error())
		}
	}

	//生成webp文件
	var flag int
	// os.O_WRONLY | os.O_CREATE | O_EXCL       【如果已经存在，则失败】
	// os.O_WRONLY | os.O_CREATE                【如果已经存在，会覆盖写，不会清空原来的文件，而是从头直接覆盖写】
	// os.O_WRONLY | os.O_CREATE | os.O_APPEND  【如果已经存在，则在尾部添加写】
	flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if outfilem, err := os.OpenFile(outMPath, flag, os.ModeAppend|os.ModePerm); err != nil {
		return fmt.Errorf("生成webp图像报错(%s)，%s", outMPath, err.Error())
	} else {
		defer outfilem.Close()
		bufm.Write(apbuf[:1024])
		bufm.WriteTo(outfilem)
	}

	var buf bytes.Buffer
	if err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: quality}); err != nil {
		return fmt.Errorf("生成webp图像报错，%s", err.Error())
	}

	//生成webp文件路径
	dir := utils.GetFileDir(outPath)
	if !utils.FileIsExist(dir) {
		if err = utils.Mkdir(dir); err != nil {
			return fmt.Errorf("创建webp图像路径报错(%s)，%s", dir, err.Error())
		}
	}

	//生成webp文件
	if outfile, err := os.OpenFile(outPath, flag, os.ModeAppend|os.ModePerm); err != nil {
		return fmt.Errorf("生成webp图像报错(%s)，%s", outPath, err.Error())
	} else {
		defer outfile.Close()
		buf.Write(apbuf[:1024])
		buf.WriteTo(outfile)
	}
	return nil
}

func Image2Webp(filePath, outPath string, quality float32) error {
	if !utils.FileIsExist(filePath) {
		return fmt.Errorf("文件不存在！(%s)", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("打开文件报错(%s)，%s", filePath, err.Error())
	}

	var img image.Image
	var imgtype = http.DetectContentType(data[:512])
	reader := bytes.NewReader(data)
	if strings.Contains(imgtype, "jpeg") {
		img, err = jpeg.Decode(reader)
	} else if strings.Contains(imgtype, "png") {
		img, err = png.Decode(reader)
	} else if strings.Contains(imgtype, "bmp") {
		img, err = bmp.Decode(reader)
	} else {
		return fmt.Errorf("图像文件格式(%s)不支持！(%s)", imgtype, filePath)
	}
	if err != nil {
		return fmt.Errorf("解析图像文件报错(%s)，%s", filePath, err.Error())
	}

	var buf bytes.Buffer
	if err = webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: quality}); err != nil {
		return fmt.Errorf("生成webp图像报错，%s", err.Error())
	}

	//生成webp文件路径
	dir := utils.GetFileDir(outPath)
	if !utils.FileIsExist(dir) {
		if err = utils.Mkdir(dir); err != nil {
			return fmt.Errorf("创建webp图像路径报错(%s)，%s", dir, err.Error())
		}
	}

	//生成webp文件
	var flag int
	// os.O_WRONLY | os.O_CREATE | O_EXCL       【如果已经存在，则失败】
	// os.O_WRONLY | os.O_CREATE                【如果已经存在，会覆盖写，不会清空原来的文件，而是从头直接覆盖写】
	// os.O_WRONLY | os.O_CREATE | os.O_APPEND  【如果已经存在，则在尾部添加写】
	flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if outfile, err := os.OpenFile(outPath, flag, os.ModeAppend|os.ModePerm); err != nil {
		return fmt.Errorf("生成webp图像报错(%s)，%s", outPath, err.Error())
	} else {
		defer outfile.Close()
		buf.WriteTo(outfile)
	}

	return nil
}

func Image2Thumbnail(filePath, outPath string, desWidth int, desHeight int) error {
	if !utils.FileIsExist(filePath) {
		return fmt.Errorf("文件不存在！(%s)", filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("打开文件报错(%s)，%s", filePath, err.Error())
	}

	var img image.Image
	var srcWidth int
	var srcHeight int

	reader := bytes.NewReader(data)
	reader2 := bytes.NewReader(data)
	var imgtype = http.DetectContentType(data[:512])
	if strings.Contains(imgtype, "jpeg") {
		img, _ = jpeg.Decode(reader)
		img2, _ := jpeg.DecodeConfig(reader2)
		srcWidth = img2.Width
		srcHeight = img2.Height
	} else if strings.Contains(imgtype, "png") {
		img, _ = png.Decode(reader)
		img2, _ := png.DecodeConfig(reader2)
		srcWidth = img2.Width
		srcHeight = img2.Height
	} else if strings.Contains(imgtype, "bmp") {
		img, _ = bmp.Decode(reader)
		img2, _ := bmp.DecodeConfig(reader2)
		srcWidth = img2.Width
		srcHeight = img2.Height
	} else {
		return fmt.Errorf("图像文件格式(%s)不支持！(%s)", imgtype, filePath)
	}
	if err != nil {
		return fmt.Errorf("解析图像文件报错(%s)，%s", filePath, err.Error())
	}

	ratio := math.Min(float64(desWidth)/float64(srcWidth), float64(desHeight)/float64(srcHeight))
	newWidth := uint(math.Ceil(float64(srcWidth) * ratio))
	newHeight := uint(math.Ceil(float64(srcHeight) * ratio))

	var buf bytes.Buffer
	m := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
	if err = webp.Encode(&buf, m, &webp.Options{Lossless: false, Quality: 85}); err != nil {
		return fmt.Errorf("生成webp图像报错，%s", err.Error())
	}

	//生成webp文件路径
	dir := utils.GetFileDir(outPath)
	if !utils.FileIsExist(dir) {
		if err = utils.Mkdir(dir); err != nil {
			return fmt.Errorf("创建webp图像路径报错(%s)，%s", dir, err.Error())
		}
	}

	//生成webp文件
	var flag int
	// os.O_WRONLY | os.O_CREATE | O_EXCL       【如果已经存在，则失败】
	// os.O_WRONLY | os.O_CREATE                【如果已经存在，会覆盖写，不会清空原来的文件，而是从头直接覆盖写】
	// os.O_WRONLY | os.O_CREATE | os.O_APPEND  【如果已经存在，则在尾部添加写】
	flag = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if outfile, err := os.OpenFile(outPath, flag, os.ModeAppend|os.ModePerm); err != nil {
		return fmt.Errorf("生成webp图像报错(%s)，%s", outPath, err.Error())
	} else {
		defer outfile.Close()
		buf.WriteTo(outfile)
	}

	return nil
}
