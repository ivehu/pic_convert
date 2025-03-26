# 图片转换工具

基于Go实现的实时图片格式转换工具，支持将JPG/PNG自动转换为WebP/AVIF格式。

## ✨ 功能特性
- 实时监控多个目录的文件变化
- 支持批量转换已有图片文件
- 自动生成.webp和.avif双格式输出
- 转换耗时统计与错误日志
- 支持多平台编译（Linux/Windows/macOS）

## 🛠️ 安装使用
#### Makefile 支持以下命令：
 - `make build`：编译当前平台
 - `make build-multi`：生成Linux/Windows/macOS的二进制文件
 - `make clean`：清理构建文件
 - `make run`：直接运行程序

## 🛠️ 前置要求
- Go 1.21+
- Linux系统需安装：
#### 安装libwebp
  sudo yum install libwebp-tools

#### 安装libavif
  https://github.com/AOMediaCodec/libavif

## 🛠️ supervisor运行程序
```
[program:pic_convert]
directory=/root/pic_convert
command=/root/pic_convert/pic_convert
autostart=true
autorestart=true
redirect_stderr=true
stdout_logfile=/webser/logs/pic_convert/pic_convert.log
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=5
stdout_capture_maxbytes=1MB
stdout_events_enabled=false
stopsignal=QUIT
```

## 🛠️ 使nginx支持WebP/AVIF格式访问
```
http{
    map $http_accept $_avif {
        ~image/avif           .avif;
        ~image/webp           .webp;
        default               '';
    }
    map $http_accept $_webp {
        ~image/webp           .webp;
        default               '';
    }
}

server{
      location ~* \.(jpg|jpeg|png)$ {
            add_header Vary Accept;
            # 按照用户请求的accept的值，会依次请求，找不到就404
            try_files   $uri$_avif  $uri$_webp  $uri  =404;
        }
}
```
修改mime.types支持avif
```
image/avif avif
``` 
