package check_code

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

// 缓存结构体
type cachedImage struct {
	img image.Image
	dy  int
}

// 全局缓存变量
var (
	imageCache     []cachedImage
	cacheInitOnce  sync.Once
	cacheInitError error
	cacheMutex     sync.RWMutex
	cacheReady     = make(chan struct{}) // 用于通知缓存已准备好
	isCacheReady   = false               // 缓存就绪标志
)

func FindBestMatch(data []byte) int {
	// 检查缓存是否就绪，如果没有就绪则等待
	if !isCacheReady {
		select {
		case <-cacheReady:
		case <-time.After(10 * time.Second):
			log.Println("等待预缓存超时")
			return 0
		}
	}

	img, err := decodeImage(data)
	if err != nil {
		fmt.Println("err:", err)
		return 0
	}
	return getMatchedX(img)
}

// 图像解码统一处理
func decodeImage(data []byte) (image.Image, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	data = nil
	return img, nil
}

type result struct {
	x    int
	diff int
}

// FindBestMatchWithImages 直接接收图片对象
func FindBestMatchWithImages(largeImg image.Image, maskPixels []pixelInfo, dx int) (int, int, error) {
	// 并行计算参数
	bounds := largeImg.Bounds()
	maxX := bounds.Dx() - dx
	var wg sync.WaitGroup

	// 启动多个 worker（按 CPU 核心数）
	numWorkers := runtime.NumCPU()
	xStep := maxX / numWorkers
	resultCh := make(chan result, numWorkers+1)

	// 如果已经是NRGBA类型，直接返回
	largeNRGBA, ok := largeImg.(*image.NRGBA)
	if !ok {
		// 将大图转换为 NRGBA 提升访问速度
		largeNRGBA = image.NewNRGBA(bounds)
		draw.Draw(largeNRGBA, bounds, largeImg, image.Point{}, draw.Src)
	}

	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		startX := workerID * xStep
		endX := startX + xStep
		if workerID == numWorkers-1 {
			endX = maxX
		}

		go func(start, end int) {
			defer wg.Done()

			currentDiff := 0
			bestXR := 0
			minDiffR := math.MaxInt64

			for x := start; x <= end; x++ {
				currentDiff = 0
				for _, p := range maskPixels {
					// 直接访问像素数组
					offset := largeNRGBA.PixOffset(x+p.X, p.Y)

					// 计算差异
					dr := int(largeNRGBA.Pix[offset]) - int(p.Color.R)
					if dr < 0 {
						dr = -dr
					}
					dg := int(largeNRGBA.Pix[offset+1]) - int(p.Color.G)
					if dg < 0 {
						dg = -dg
					}
					db := int(largeNRGBA.Pix[offset+2]) - int(p.Color.B)
					if db < 0 {
						db = -db
					}

					currentDiff += dr + dg + db
				}

				if currentDiff < minDiffR {
					minDiffR = currentDiff
					bestXR = x
				}
			}
			resultCh <- result{x: bestXR, diff: minDiffR}
		}(startX, endX)
	}

	// 等待并收集结果
	go func() {
		wg.Wait()
		close(resultCh)
	}()
	minDiff := math.MaxInt64
	bestX := 0

	for res := range resultCh {
		if res.diff < minDiff {
			minDiff = res.diff
			bestX = res.x
		}
	}

	return bestX, minDiff, nil
}

//go:embed merged/*.png
var embeddedFS embed.FS

// initImageCache 并发初始化图片缓存
func initImageCache() {
	startTime := time.Now()

	entries, err := embeddedFS.ReadDir("merged")
	if err != nil {
		cacheInitError = fmt.Errorf("读取嵌入目录失败: %v", err)
		isCacheReady = true
		close(cacheReady)
		return
	}

	var wg sync.WaitGroup
	tempCache := make([]cachedImage, 0, len(entries))
	var tempMutex sync.Mutex

	// 限制并发数，避免过多占用资源
	maxConcurrency := runtime.NumCPU() * 2
	semaphore := make(chan struct{}, maxConcurrency)

	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".png" {
			continue
		}

		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			filePath := path.Join("merged", filename)
			data, readErr := embeddedFS.ReadFile(filePath)
			if readErr != nil {
				log.Printf("读取失败 %s: %v", filePath, readErr)
				return
			}

			reader := bytes.NewReader(data)
			img, decErr := png.Decode(reader)
			if decErr != nil {
				log.Printf("解码失败 %s: %v", filePath, decErr)
				return
			}

			tempMutex.Lock()
			tempCache = append(tempCache, cachedImage{
				img: img,
				dy:  img.Bounds().Dy(),
			})
			tempMutex.Unlock()
		}(entry.Name())
	}

	wg.Wait()
	close(semaphore)

	cacheMutex.Lock()
	imageCache = tempCache
	cacheMutex.Unlock()

	if len(imageCache) == 0 {
		cacheInitError = fmt.Errorf("没有成功加载任何图片")
		log.Println("警告: 没有成功加载任何图片")
	} else if len(imageCache) != 10 {
		log.Printf("警告: 共加载 %d 张图片, 耗时 %v", len(imageCache), time.Since(startTime))
	}

	isCacheReady = true
	close(cacheReady)
}

// getCachedImages 获取缓存的图片
func getCachedImages() ([]cachedImage, error) {
	if !isCacheReady {
		return nil, fmt.Errorf("图片缓存尚未就绪")
	}

	if cacheInitError != nil {
		return nil, cacheInitError
	}

	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return imageCache, nil
}

func PreloadCache() {
	cacheInitOnce.Do(func() {
		go initImageCache()
	})
}

func getMatchedX(img image.Image) int {
	// 获取输入图片的非透明像素坐标和颜色
	maskPixels, dx, dy := lowerWhiteGetNonTransParentPixels(img)

	// 获取缓存的图片
	cachedImages, err := getCachedImages()
	if err != nil {
		log.Printf("获取缓存图片失败: %v", err)
		return 0
	}

	resultCh := make(chan result, len(cachedImages)+1)
	var wg sync.WaitGroup

	for _, cached := range cachedImages {
		// 验证高度一致
		if cached.dy != dy {
			continue
		}

		wg.Add(1)
		go func(cachedImg cachedImage) {
			defer wg.Done()

			// 使用缓存的图片直接进行匹配
			currentX, currentDiff, matchErr := FindBestMatchWithImages(cachedImg.img, maskPixels, dx)
			if matchErr != nil {
				fmt.Println("匹配错误:", matchErr)
				return
			}
			resultCh <- result{x: currentX, diff: currentDiff}
		}(cached)
	}

	// 等待并收集结果
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	minDiff := math.MaxInt64
	bestX := 0

	for res := range resultCh {
		if res.diff < minDiff {
			minDiff = res.diff
			bestX = res.x
		}
	}

	return bestX + 70
}

// 像素信息结构体
type pixelInfo struct {
	X, Y  int
	Color color.NRGBA
}

func lowerWhiteGetNonTransParentPixels(img image.Image) ([]pixelInfo, int, int) {
	// 创建新图片容器（使用NRGBA以保留透明度）
	bounds := img.Bounds()
	pixels := make([]pixelInfo, 0, 2680)

	const passNum uint8 = 25

	rgbaImg, ok := img.(*image.NRGBA)
	if !ok {
		rgbaImg = image.NewNRGBA(bounds)
		draw.Draw(rgbaImg, bounds, img, image.Point{}, draw.Src)
	}

	count := 0
	// 遍历每个像素进行调整
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			count++
			// 选择性跳过一些像素
			if count%3 == 0 {
				continue
			}
			// 获取原始颜色和透明度
			// 尝试直接访问 *image.RGBA
			idx := rgbaImg.PixOffset(x, y)
			r8, g8, b8, a8 := rgbaImg.Pix[idx], rgbaImg.Pix[idx+1], rgbaImg.Pix[idx+2], rgbaImg.Pix[idx+3]

			newR := r8 - passNum
			newG := g8 - passNum
			newB := b8 - passNum

			// 判断是否为纯白色且不透明
			if r8 == 255 && g8 == 255 && b8 == 255 && a8 == 255 {
				continue
			} else if a8 > 0 {
				pixels = append(pixels, pixelInfo{
					X:     x - bounds.Min.X, // 转换为相对坐标
					Y:     y - bounds.Min.Y,
					Color: color.NRGBA{R: newR, G: newG, B: newB},
				})
			}

		}
	}

	return pixels, bounds.Dx(), bounds.Dy()
}

// 包初始化时自动开始预加载
func init() {
	PreloadCache()
}
