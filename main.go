package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Directories []string `yaml:"directories"`
	Conversion  struct {
		Webp struct {
			Quality int  `yaml:"quality"`
			Method  int  `yaml:"method"`
			Threads bool `yaml:"threads"`
		} `yaml:"webp"`
		Avif struct {
			MinQuality int `yaml:"min_quality"`
			MaxQuality int `yaml:"max_quality"`
			Speed      int `yaml:"speed"`
			Depth      int `yaml:"depth"`
			Threads    int `yaml:"threads"`
		} `yaml:"avif"`
	} `yaml:"conversion"`
}

var cfg *Config

func loadConfig() *Config {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	return &cfg
}

func main() {
	cfg = loadConfig()
	ctx := context.Background()

	for _, dir := range cfg.Directories {
		go processExistingFiles(dir)
		go watchDirectory(ctx, dir)
	}

	select {} // 保持主程序运行
}

func processExistingFiles(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".jpg" || ext == ".png" {
			// 检查目标文件是否已存在且更新
			webpPath := path + ".webp"
			avifPath := path + ".avif"

			// 获取源文件修改时间
			srcModTime := info.ModTime()

			// 检查WebP文件
			webpExists := fileExists(webpPath)
			webpNewer := webpExists && fileModTime(webpPath).After(srcModTime)

			// 检查AVIF文件
			avifExists := fileExists(avifPath)
			avifNewer := avifExists && fileModTime(avifPath).After(srcModTime)

			// 仅当两个目标文件都不存在或需要更新时才转换
			if !(webpNewer && avifNewer) {
				convertImage(path)
			}
		}
		return nil
	})
}

// 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// 获取文件修改时间
func fileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// 修改watchDirectory函数
func watchDirectory(ctx context.Context, dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 递归添加监控
	addWatchRecursively(watcher, dir)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create {
				// 如果是新建目录则递归监控
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					addWatchRecursively(watcher, event.Name)
				}
			}
			if event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Write == fsnotify.Write {
				// 延迟处理确保文件完全写入
				time.AfterFunc(1*time.Second, func() {
					convertImage(event.Name)
				})
			}
		case err := <-watcher.Errors:
			log.Println("监控错误:", err)
		case <-ctx.Done():
			return
		}
	}
}

// 新增递归监控函数
func addWatchRecursively(watcher *fsnotify.Watcher, dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if err := watcher.Add(path); err != nil {
				log.Printf("监控添加失败: %s - %v", path, err)
			}
		}
		return nil
	})
}

func convertImage(srcPath string) {
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext != ".jpg" && ext != ".png" {
		return
	}

	// 生成目标文件名
	webpPath := srcPath + ".webp"
	avifPath := srcPath + ".avif"

	// 转换WebP
	startWebp := time.Now()
	args := []string{
		"-q", fmt.Sprintf("%d", cfg.Conversion.Webp.Quality),
		"-m", fmt.Sprintf("%d", cfg.Conversion.Webp.Method),
	}
	if cfg.Conversion.Webp.Threads {
		args = append(args, "-mt")
	}
	args = append(args, srcPath, "-o", webpPath)
	cmdWebp := exec.Command("cwebp", args...)
	if errWebp := cmdWebp.Run(); errWebp != nil {
		log.Printf("WebP转换失败 [%s] 耗时 %v: %v", webpPath, time.Since(startWebp).Round(time.Millisecond), errWebp)
	} else {
		log.Printf("成功转换WebP [%s] 耗时 %v", webpPath, time.Since(startWebp).Round(time.Millisecond))
	}

	// 转换AVIF
	startAvif := time.Now()
	cmdAvif := exec.Command("avifenc",
		"--min", fmt.Sprintf("%d", cfg.Conversion.Avif.MinQuality),
		"--max", fmt.Sprintf("%d", cfg.Conversion.Avif.MaxQuality),
		"-s", fmt.Sprintf("%d", cfg.Conversion.Avif.Speed),
		"--depth", fmt.Sprintf("%d", cfg.Conversion.Avif.Depth),
		"-j", fmt.Sprintf("%d", cfg.Conversion.Avif.Threads),
		srcPath,
		avifPath,
	)
	if errAvif := cmdAvif.Run(); errAvif != nil {
		log.Printf("AVIF转换失败 [%s] 耗时 %v: %v", avifPath, time.Since(startAvif).Round(time.Millisecond), errAvif)
	} else {
		log.Printf("成功转换AVIF [%s] 耗时 %v", avifPath, time.Since(startAvif).Round(time.Millisecond))
	}
}
