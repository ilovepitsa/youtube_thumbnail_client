package grpc

import (
	"context"
	"fmt"
	"io"
	"sync"

	pb "github.com/ilovepitsa/youtube_tumbnail/pkg/handlers/grpc/thumbnail"
	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

type ThumbnailClient struct {
	client pb.ThumbnailsClient
}

func NewThumbnailClient(conn grpc.ClientConnInterface) *ThumbnailClient {
	return &ThumbnailClient{
		client: pb.NewThumbnailsClient(conn),
	}
}

func (tc *ThumbnailClient) GetThumbnail(videoId string) ([]byte, error) {

	resp, err := tc.client.GetThumbnail(context.TODO(), &pb.ThumbnailRequest{Id: videoId})
	if err != nil {
		return nil, err
	}
	// log.Info("Answer: ", resp)
	// var res []byte
	// copy(res, resp.Data)
	return resp.Data, err
}

func (tc *ThumbnailClient) sendId(stream pb.Thumbnails_GetThumbnailsAsyncClient, videoIds []string, errC chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, id := range videoIds {
		if err := stream.Send(&pb.ThumbnailRequest{Id: id}); err != nil {
			errC <- fmt.Errorf("error while sending video %w", err)
			return
		}
	}
	err := stream.CloseSend()
	log.Error("send end", err)
}

func (tc *ThumbnailClient) recvData(stream pb.Thumbnails_GetThumbnailsAsyncClient, output chan []byte, errC chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	i := 0
	for {
		log.Error("get new Recv ", i)
		in, err := stream.Recv()
		if err == io.EOF {
			close(output)
			log.Error("recv end")
			return
		}

		if err != nil {
			errC <- fmt.Errorf("error while recv data %w", err)
			return
		}

		output <- in.Data
		i++
	}

}

func (tc *ThumbnailClient) GetThumbnailsAsync(waitc chan struct{}, videosId []string) (chan []byte, chan error) {
	outData := make(chan []byte)
	outError := make(chan error)
	go func() {
		defer close(outError)

		stream, err := tc.client.GetThumbnailsAsync(context.Background())
		if err != nil {
			outError <- err
			return
		}
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go tc.sendId(stream, videosId, outError, wg)
		go tc.recvData(stream, outData, outError, wg)

		wg.Wait()
		waitc <- struct{}{}
		close(waitc)
		log.Error("this end")
	}()
	return outData, outError
}
