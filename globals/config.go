package globals

import (
	"os"
	"strconv"
)

type Config struct {
	Port      int
	MediaDir  string
	ChunkSize int64
	RedisAddr string
}
var Conf = Config{}

func LoadConfig() {
	// Port
	if port, err := strconv.Atoi(os.Getenv("PORT")); err == nil {
		Conf.Port = port
	}else {
		Conf.Port = 6969
	}
	// Media Dir
	if dir := os.Getenv("MEDIA_DIR"); dir != "" {
		Conf.MediaDir = dir
	} else {
		Conf.MediaDir = "./media"
	}
	// ChunkSize
	if size, err := strconv.Atoi(os.Getenv("CHUNK_SIZE")); err == nil {
		Conf.ChunkSize = int64(size)
	}else {
		Conf.ChunkSize = 10 << 20
	}
	// RedisAddr
	if addr := os.Getenv("REDIS_URI"); addr != "" {
		Conf.RedisAddr = addr
	} else {
		Conf.RedisAddr = "localhost:6379"
	}
}