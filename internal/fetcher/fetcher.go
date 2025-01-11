package fetcher

import (
	"sync"

	"tmd/internal/db"
	"tmd/minio"
	"tmd/pkg/filehandler"

	"github.com/gotd/td/telegram"
)

type Fetcher struct {
	client        *telegram.Client
	downloader    *filehandler.Downloader
	database      *db.DB
	storage       *minio.Storage
	dialogsLimit  int
	messagesLimit int

	meChan chan MeJob
	wg     sync.WaitGroup
}

func NewFetcher(client *telegram.Client,
	downloader *filehandler.Downloader,
	database *db.DB,
	storage *minio.Storage,
	dialogsLimit, messagesLimit int,
) *Fetcher {
	f := &Fetcher{
		client:        client,
		downloader:    downloader,
		database:      database,
		storage:       storage,
		dialogsLimit:  dialogsLimit,
		messagesLimit: messagesLimit,
		meChan:        make(chan MeJob, 100),
	}

	workerCount := 5
	for i := 0; i < workerCount; i++ {
		f.wg.Add(1)
		go f.workerMeJob()
	}

	return f
}

func (f *Fetcher) CloseWorkers() {
	close(f.meChan)
	f.wg.Wait()
}
