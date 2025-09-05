# Image - 图像处理

## 概述

Image包提供了全面的图像处理功能，包括格式转换、尺寸调整、EXIF信息提取等。支持JPEG、PNG、BMP等常见图像格式的读取，并能够将图像转换为WebP格式以获得更好的压缩率和性能。

## 主要功能

### 1. 支持的图像格式
- **输入格式**: JPEG、PNG、BMP
- **输出格式**: WebP
- **自动格式检测**: 基于文件内容自动识别格式

### 2. 核心结构

#### TExif - EXIF信息结构
```go
type TExif struct {
    CameraName  string // 相机厂商
    DeviceModel string // 设备型号
    PhotoTime   string // 拍摄时间
    PixelX      string // 图片尺寸X
    PixelY      string // 图片尺寸Y
}
```

#### IfdEntry - EXIF条目结构
```go
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
```

### 3. 主要功能函数

#### 图像转换与缩放
```go
func ImageConv(filePath, outPath, outMPath string, quality float32, desWidth int, desHeight int) error
```

#### 简单图像转WebP
```go
func Image2Webp(filePath, outPath string, quality float32) error
```

#### 生成缩略图
```go
func Image2Thumbnail(filePath, outPath string, desWidth int, desHeight int) error
```

#### EXIF信息提取
```go
func get_exif(ex *TExif, data []byte) error
func get_exif_value(entries []*IfdEntry, name string) string
```

## 使用示例

### 基本图像转换
```go
package main

import (
    "fmt"
    "log"
    "path/to/image"
)

func main() {
    // 简单的图像转WebP格式
    err := image.Image2Webp("input.jpg", "output.webp", 85.0)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("图像转换成功")
}
```

### 带缩放的图像转换
```go
package main

import (
    "fmt"
    "log"
    "path/to/image"
)

func main() {
    // 转换图像并生成两个版本：原尺寸和缩略图
    err := image.ImageConv(
        "input.jpg",        // 输入文件路径
        "output.webp",      // 输出文件路径（高质量）
        "output_thumb.webp", // 输出缩略图路径
        90.0,               // 高质量版本质量
        1920,               // 缩略图最大宽度
        1080,               // 缩略图最大高度
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("图像转换和缩放成功")
}
```

### 生成缩略图
```go
package main

import (
    "fmt"
    "log"
    "path/to/image"
)

func main() {
    // 生成缩略图
    err := image.Image2Thumbnail(
        "large_image.jpg", // 输入文件
        "thumbnail.webp",  // 输出缩略图
        300,              // 最大宽度
        200,              // 最大高度
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("缩略图生成成功")
}
```

### 批量图像处理
```go
package main

import (
    "fmt"
    "log"
    "path/filepath"
    "path/to/image"
    "path/to/utils"
)

func main() {
    inputDir := "input/"
    outputDir := "output/"
    
    // 获取所有图像文件
    files, err := filepath.Glob(filepath.Join(inputDir, "*"))
    if err != nil {
        log.Fatal(err)
    }
    
    for _, file := range files {
        // 检查是否为图像文件
        if !isImageFile(file) {
            continue
        }
        
        // 生成输出文件名
        filename := filepath.Base(file)
        ext := filepath.Ext(filename)
        name := filename[:len(filename)-len(ext)]
        outputFile := filepath.Join(outputDir, name+".webp")
        thumbFile := filepath.Join(outputDir, "thumbs", name+"_thumb.webp")
        
        // 转换图像
        err := image.ImageConv(file, outputFile, thumbFile, 85.0, 800, 600)
        if err != nil {
            log.Printf("转换失败 %s: %v", file, err)
            continue
        }
        
        fmt.Printf("转换完成: %s -> %s\n", file, outputFile)
    }
}

func isImageFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".bmp"
}
```

### 图像信息提取
```go
package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "path/to/image"
)

func main() {
    // 读取图像文件
    data, err := ioutil.ReadFile("photo.jpg")
    if err != nil {
        log.Fatal(err)
    }
    
    // 提取EXIF信息
    var exif image.TExif
    err = get_exif(&exif, data)
    if err != nil {
        log.Printf("提取EXIF信息失败: %v", err)
        return
    }
    
    // 输出EXIF信息
    fmt.Printf("相机厂商: %s\n", exif.CameraName)
    fmt.Printf("设备型号: %s\n", exif.DeviceModel)
    fmt.Printf("拍摄时间: %s\n", exif.PhotoTime)
    fmt.Printf("图片尺寸: %s x %s\n", exif.PixelX, exif.PixelY)
}
```

## 功能特性

### 1. 智能缩放
- 保持图像宽高比
- 使用Lanczos3算法确保高质量缩放
- 自动计算最佳缩放比例

### 2. EXIF信息处理
- 自动提取相机制造商信息
- 获取设备型号和拍摄时间
- 提取图像尺寸信息
- 支持将EXIF信息嵌入到输出文件

### 3. 文件处理安全性
- 输入文件存在性检查
- 自动创建输出目录
- 防止文件覆盖的安全模式
- 完整的错误处理和回滚

### 4. 格式支持
- 自动检测输入图像格式
- 统一输出为WebP格式
- 支持有损和无损压缩
- 可调节压缩质量

## 依赖库

- `github.com/chai2010/webp` - WebP编码支持
- `github.com/dsoprea/go-exif` - EXIF信息处理
- `github.com/nfnt/resize` - 图像缩放算法
- `golang.org/x/image/bmp` - BMP格式支持

## 注意事项

1. **文件权限**: 确保对输入文件有读权限，对输出目录有写权限
2. **内存使用**: 处理大图像时注意内存消耗，建议分批处理
3. **质量设置**: WebP质量参数范围0-100，建议使用75-90获得最佳平衡
4. **文件覆盖**: 默认使用O_EXCL标志防止覆盖现有文件
5. **EXIF数据**: 只有JPEG格式支持EXIF信息提取
6. **错误处理**: 所有函数都返回详细错误信息，便于调试

## 最佳实践

1. **批量处理**: 使用goroutine并行处理多个图像以提高效率
2. **质量优化**: 根据使用场景选择合适的质量参数
3. **存储优化**: 生成多个尺寸版本以适应不同显示需求
4. **错误恢复**: 实现重试机制处理临时性错误
5. **资源管理**: 及时关闭文件句柄避免资源泄漏

## 性能优化建议

- 对于大量小图片，考虑使用图像池减少内存分配
- 预先创建输出目录避免重复检查
- 使用适当的缓冲区大小优化I/O性能
- 根据CPU核心数调整并发处理数量