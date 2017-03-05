package main

import (
	"common"
	"config"
	"context"
	"feed"
	"flag"
	"fmt"
	"log"
	"mysql"
	"runtime"
	"time"
	"util"
	"net/http"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	_ "github.com/go-sql-driver/mysql"
)

type feedItemCopy struct {
	feedId      int64
	url         string
	title       string
	description string
	content     string
	image       *gofeed.Image
	published   *time.Time
	updated     *time.Time
}

type feedItemChannels struct {
	resultChannel chan string
	httpChannel   chan *common.HttpChannels
	errorChannel  chan error
}

func (chs *feedItemChannels) closeAll() {
	close(chs.resultChannel)
	close(chs.errorChannel)
}

var redirectOption = flag.Bool("r", false, "get redirect url")

func getOriginalUrl(chs feedItemChannels, url string) {
	// 平均秒間リクエスト量の紳士協定
	time.Sleep(config.MINIMUM_INTERVAL_SECONDS * time.Second)

	var eur = http.ErrUseLastResponse

	transp := &http.Transport{}
	cl := &http.Client{
		Transport: transp,
		Timeout:   config.HTTP_TIMEOUT_SECONDS * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return eur
		},
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}

	req.Header.Set("User-Agent", config.USER_AGENT)

	hch := common.HttpChannels{
		TransportChannel: make(chan *http.Transport, 1),
		RequestChannel:   make(chan *http.Request, 1),
	}
	hch.TransportChannel <- transp
	hch.RequestChannel <- req
	chs.httpChannel <- &hch

	resp, err := cl.Do(req)
	if err != nil && err != eur {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}
	defer resp.Body.Close()

	if len(resp.Header["Location"]) > 0 {
		chs.resultChannel <- resp.Header["Location"][0]
	} else {
		chs.resultChannel <- ""
	}

	chs.errorChannel <- nil
}

func handleRedirectCrawler(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.HTTP_TIMEOUT_SECONDS*time.Second)
	defer cancel()

	chs := feedItemChannels{
		resultChannel: make(chan string, 1),
		httpChannel:   make(chan *common.HttpChannels, 1),
		errorChannel:  make(chan error, 1),
	}
	defer func() {
		hch := <-chs.httpChannel
		hch.CloseAll()
	}()
	defer chs.closeAll()

	go getOriginalUrl(chs, url)

	select {
	case err := <-chs.errorChannel:
		if err != nil {
			hch := <-chs.httpChannel
			transp := <-hch.TransportChannel
			req := <-hch.RequestChannel
			transp.CancelRequest(req)
			return "", err
		}

		return <-chs.resultChannel, nil
	case <-ctx.Done():
		hch := <-chs.httpChannel
		transp := <-hch.TransportChannel
		req := <-hch.RequestChannel
		transp.CancelRequest(req)

		return "", ctx.Err()
	}
}

func handleFeedItem(db mysql.Database, jc *common.JobController, tp int, fi feedItemCopy) {
	defer func() {
		<-jc.Channel
		jc.WaitGroup.Done()
	}()

	if tp == 1 {
		regUrl, err := handleRedirectCrawler(fi.url)
		if err != nil {
			fmt.Println(errors.Wrap(err, ""))
			return
		}
		if regUrl != "" {
			fi.url = regUrl
		}
	}

	if fi.url != "" {
		fi.url = util.RemoveParameters(fi.url)
	}

	if fi.title != "" {
		t := util.Html{Body: fi.title}
		t.Compress().RemoveNonTextElements().RemoveAllTags()
		fi.title = t.Body
	}

	if fi.description != "" {
		html := util.Html{Body: fi.description}
		html.Compress().RemoveNonTextElements().RemoveAllTags()
		fi.description = html.Body
	}

	if fi.published == nil && fi.updated == nil {
		now := time.Now()
		fi.published, fi.updated = &now, &now
	}
	if fi.published == nil {
		fi.published = fi.updated
	} else if fi.updated == nil {
		fi.updated = fi.published
	}

	_, err := db.Execute(
		`INSERT INTO pages
		 (feed_id, url, title, description, has_thumbnail,
		  published_at, updated_at, registered_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		  pages.url = VALUES(url)`,
		fi.feedId, fi.url,
		util.Substring(fi.title, 0, 256),
		util.Substring(fi.description, 0, 1024),
		0, fi.published.Unix(), fi.updated.Unix(), time.Now().Unix(),
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}
	//lastInsertId, _ := res.LastInsertId()

	// これもか
	/*if hasThumb == 1 {
		_, err = db.Execute(
			`INSERT INTO page_thumbnails (page_id, url, registered_at) VALUES (?, ?, ?)`,
			lastInsertId, img, time.Now().Unix(),
		)
		if err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}
	}*/

	// 過去取得した記事を自ら削っていくアホ
	/*if fi.content != "" {
		_, err = db.Execute(
			`INSERT INTO page_contents (page_id, content, registered_at) VALUES (?, ?, ?)`,
			lastInsertId, fi.content, time.Now().Unix(),
		)
		if err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}
	}*/
}

// TODO : feed14のpubdataがとれてない、なんとかして
func main() {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	flag.Parse()

	tp := 0
	if *redirectOption {
		tp = 1
	}

	db := mysql.Database{}
	if err := db.Connect(); err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}
	defer db.Close()

	jc := common.JobController{
		Channel: make(chan int, config.HTTP_WORKER_LIMIT),
	}

	feeds, err := db.FetchAll(
		`SELECT
		  f.id,
		  f.url
		 FROM feeds AS f
		 WHERE f.crawler_status = ?
		   AND f.do_redirect    = ?`,
		1, tp,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}

	for _, f := range feeds {
		feedId := f[0].(int64)
		feedUrl := f[1].(string)

		pages, err := feed.FetchFeedItem(feedUrl)
		if err != nil {
			fmt.Println(errors.Wrap(err, ""))
			continue
		}

		for _, p := range pages {
			if p.Url == "" || p.Title == "" {
				continue
			}

			fi := feedItemCopy{
				feedId:  feedId, url: p.Url, title: p.Title, description: p.Description,
				content: p.Content, image: p.Image, published: p.Published, updated: p.Updated,
			}

			jc.Channel <- 1
			jc.WaitGroup.Add(1)
			go handleFeedItem(db, &jc, tp, fi)
		}

		jc.WaitGroup.Wait()
	}
}
