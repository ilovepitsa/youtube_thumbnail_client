package main

import (
	"flag"
	"fmt"
	"time"

	"youtube-thumbnail-client/config"
	thumbnailGrpc "youtube-thumbnail-client/internal/grpc"
	"youtube-thumbnail-client/internal/thumbnail"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultConfigPath = "config/config.yaml"

// flags
var (
	async    = flag.Bool("async", false, "using for async download images.")
	videos   = flag.String("v", "", "videos links. Cant be used with flag 'file'. Must be last.")
	out      = flag.String("out", "./", "output dir for images.")
	file     = flag.String("file", "", "file with videos links. Cant be used with flag 'v'.")
	showTime = flag.Bool("t", false, "show duration of work program.  Can show that cache is working.")
)

func main() {

	flag.Parse()
	cfg, err := config.NewConfig(defaultConfigPath)
	if err != nil {
		fmt.Println("Cant read config")
	}
	var videoList []string
	SetLogrus(cfg.Logger)

	if *videos != "" && *file == "" {
		videoList = append(videoList, *videos)
		videoList = append(
			videoList,
			flag.Args()...,
		)
	} else if *file != "" && *videos == "" {
		vl, err := readVideoLinkFromFile(*file)
		if err != nil {
			log.Error(err)
			return
		}
		videoList = append(videoList, vl...)
	} else {
		fmt.Println("No input")
		return
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), opts...)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	cl := thumbnailGrpc.NewThumbnailClient(conn)

	service := thumbnail.NewService(cl, *out)
	start := time.Now()
	if *async {
		service.GetThumbnailsAsync(videoList)
	} else {
		service.GetThumbnails(videoList)
	}

	duration := time.Since(start)
	if *showTime {
		fmt.Println(duration.Milliseconds())
	}

}
