package thumbnail

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	thumbnailGrpc "youtube-thumbnail-client/internal/grpc"

	log "github.com/sirupsen/logrus"
)

type ThumbnailService struct {
	client     *thumbnailGrpc.ThumbnailClient
	outputPath string
}

func NewService(client *thumbnailGrpc.ThumbnailClient, outputPath string) *ThumbnailService {

	return &ThumbnailService{
		client:     client,
		outputPath: outputPath,
	}
}

func getVideoId(videoUrlStr string) (string, error) {
	var videoId string
	videoUrl, err := url.Parse(videoUrlStr)
	if err != nil {
		return "", fmt.Errorf("cant parse video id from %s   err: %w", videoUrlStr, err)
	}
	params := videoUrl.Query()
	videoId = params.Get("v")
	return videoId, nil
}

func (ts *ThumbnailService) writeDataToFile(videoId string, data []byte) error {
	file, err := os.Create(path.Join(ts.outputPath, videoId+".jpg"))
	if err != nil {
		return fmt.Errorf("cant write data from %v to file: %w", videoId, err)
	}
	defer file.Close()
	log.Info(videoId)
	var byteBuffer bytes.Buffer
	byteBuffer.Write(data)
	_, err = io.Copy(file, &byteBuffer)
	if err != nil {
		return fmt.Errorf("cant write data from %v to file: %w", videoId, err)
	}
	return nil
}

func (ts *ThumbnailService) GetThumbnails(videoList []string) {
	for index, video := range videoList {

		videoId, err := getVideoId(video)

		log.Debug(fmt.Sprintf("video with link %s have id %s", video, videoId))
		if err != nil {
			log.Error(err)
			continue
		}
		// if videoId == "" {
		// 	continue
		// }

		result, err := ts.client.GetThumbnail(videoId)
		if err != nil {
			log.Error(err)
			continue
		}
		err = ts.writeDataToFile(strconv.Itoa(index), result)
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

func convertLinkToId(videoList []string) []string {
	var videoIds []string

	for _, videoUrl := range videoList {
		videoUrl, err := url.Parse(videoUrl)
		if err != nil {
			continue
		}
		params := videoUrl.Query()
		videoIds = append(videoIds, params.Get("v"))
	}

	return videoIds
}

func (ts *ThumbnailService) GetThumbnailsAsync(videoList []string) {
	waitc := make(chan struct{})
	outChan, errChan := ts.client.GetThumbnailsAsync(waitc, convertLinkToId(videoList))
	index := 0
RECV:
	for {
		select {
		case data := <-outChan:
			err := ts.writeDataToFile(strconv.Itoa(index), data)
			if err != nil {
				log.Error(err)
				continue
			}
		case err := <-errChan:
			log.Error(err)
			break RECV
		case <-waitc:
			break RECV
		}
		index++
	}

}
