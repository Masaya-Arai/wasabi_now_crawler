package main

import (
	"common"
	"config"
	"context"
	"flag"
	"fmt"
	"log"
	"mysql"
	"runtime"
	"time"
	"util"
	"net/http"
	"github.com/pkg/errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
	_ "github.com/go-sql-driver/mysql"
)

type pageInfo struct {
	id                         int
	url                        string
	thumbnailAttribute         string
	thumbnailSelector          *common.OptionalString
	thumbnailCaptionSelector   *common.OptionalString
	thumbnailCopyrightSelector *common.OptionalString
	contentSelector            *common.OptionalString
	copyrightSelector          *common.OptionalString
}

type pageResult struct {
	thumbnailUrl       string
	thumbnailCaption   string
	thumbnailCopyright string
	content            string
	copyright          string
}

type pageChannels struct {
	resultChannel    chan *pageResult
	webDriverChannel chan *agouti.WebDriver
	httpChannel      chan *common.HttpChannels
	errorChannel     chan error
}

func (chs *pageChannels) closeAll() {
	close(chs.resultChannel)
	close(chs.webDriverChannel)
	close(chs.errorChannel)
}

var javascriptOption = flag.Bool("j", false, "get contents using JavaScript")

func crawl(chs pageChannels, pi pageInfo) {
	// 平均秒間リクエスト量の紳士協定
	time.Sleep(config.MINIMUM_INTERVAL_SECONDS * time.Second)

	res := &pageResult{}

	transp := &http.Transport{}
	cl := &http.Client{
		Transport: transp,
		Timeout:   config.HTTP_TIMEOUT_SECONDS * time.Second,
	}

	req, err := http.NewRequest("GET", pi.url, nil)
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
	if err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}

	if pi.thumbnailSelector.Status {
		tmpUrl, _ := doc.Find(pi.thumbnailSelector.String).First().Attr(pi.thumbnailAttribute)
		if tmpUrl != "" {
			res.thumbnailUrl = util.GetRegularUrl(pi.url, tmpUrl)
		}
	}

	if pi.thumbnailCaptionSelector.Status {
		res.thumbnailCaption = doc.Find(pi.thumbnailCaptionSelector.String).First().Text()
	}

	if pi.thumbnailCopyrightSelector.Status {
		res.thumbnailCopyright = doc.Find(pi.thumbnailCopyrightSelector.String).First().Text()
	}

	if pi.contentSelector.Status {
		doc.Find(pi.contentSelector.String).Each(func(_ int, sel *goquery.Selection) {
			txt, err := goquery.OuterHtml(sel)
			tmp := util.Html{Url: util.Url{Url: pi.url}, Body: txt}
			tmp.Compress().RemoveNonTextElements().Simplify().ReplaceUrl()
			if err == nil {
				res.content += tmp.Body
			}
		})
	}

	if pi.copyrightSelector.Status {
		res.copyright = doc.Find(pi.copyrightSelector.String).First().Text()
	}

	chs.resultChannel <- res
	chs.errorChannel <- nil
}

func crawlWithAgouti(chs pageChannels, pi pageInfo) {
	res := &pageResult{}

	bo := agouti.Browser("phantomjs")
	capa := agouti.NewCapabilities()
	capa["phantomjs.page.settings.userAgent"] = config.USER_AGENT
	co := agouti.Desired(capa)
	to := agouti.Timeout(int(config.AGOUTI_TIMEOUT_SECONDS))

	wd := agouti.PhantomJS(bo, co, to)
	if err := wd.Start(); err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}
	chs.webDriverChannel <- wd

	page, err := wd.NewPage()
	if err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}
	if err := page.Size(config.AGOUTI_PAGE_WIDTH, config.AGOUTI_PAGE_HEIGHT); err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}
	if err := page.Navigate(pi.url); err != nil {
		chs.errorChannel <- errors.Wrap(err, "")
		return
	}

	if pi.thumbnailSelector.Status {
		tmpUrl, _ := page.First(pi.thumbnailSelector.String).Attribute(pi.thumbnailAttribute)
		if tmpUrl != "" {
			res.thumbnailUrl = util.GetRegularUrl(pi.url, tmpUrl)
		}
	}

	if pi.thumbnailCaptionSelector.Status {
		res.thumbnailCaption, _ = page.First(pi.thumbnailCaptionSelector.String).Text()
	}

	if pi.thumbnailCopyrightSelector.Status {
		res.thumbnailCopyright, _ = page.First(pi.thumbnailCopyrightSelector.String).Text()
	}

	if pi.contentSelector.Status {
		ms := page.All(pi.contentSelector.String)
		for i := 0; ; i++ {
			txt, err := ms.At(i).Text()
			if err != nil {
				break
			}
			if txt != "" {
				res.content += "<p>" + txt + "</p>"
			}
		}
	}

	if pi.copyrightSelector.Status {
		res.copyright, _ = page.First(pi.copyrightSelector.String).Text()
	}

	chs.resultChannel <- res
	chs.errorChannel <- nil
}

func handleCrawler(tp int, pi pageInfo) (*pageResult, error) {
	tLim := config.HTTP_TIMEOUT_SECONDS
	if tp == 1 {
		tLim = config.AGOUTI_TIMEOUT_SECONDS
	}

	ctx, cancel := context.WithTimeout(context.Background(), tLim*time.Second)
	defer cancel()

	chs := pageChannels{
		resultChannel:    make(chan *pageResult, 1),
		webDriverChannel: make(chan *agouti.WebDriver, 1),
		httpChannel:      make(chan *common.HttpChannels, 1),
		errorChannel:     make(chan error, 1),
	}
	/*
	defer func() {
		if tp == 0 {
			hch := <-chs.httpChannel
			hch.CloseAll()
		}
	}()
	defer chs.closeAll()
	*/

	switch tp {
	case 0:
		go crawl(chs, pi)
	case 1:
		go crawlWithAgouti(chs, pi)
	}

	select {
	case err := <-chs.errorChannel:
		fmt.Println(fmt.Sprintf("Finished crawling: page_id = %v", pi.id))

		if tp == 1 {
			wd := <-chs.webDriverChannel
			wd.Stop()
		} else {
			hch := <-chs.httpChannel
			transp := <-hch.TransportChannel
			req := <-hch.RequestChannel
			transp.CancelRequest(req)
		}

		if err != nil {
			return nil, errors.Wrap(err, "")
		}

		return <-chs.resultChannel, nil
	case <-ctx.Done():
		fmt.Println(fmt.Sprintf("Timed out: page_id = %v", pi.id))

		if tp == 1 {
			wd := <-chs.webDriverChannel
			wd.Stop()
		} else {
			hch := <-chs.httpChannel
			transp := <-hch.TransportChannel
			req := <-hch.RequestChannel
			transp.CancelRequest(req)
		}

		return nil, errors.Wrap(ctx.Err(), "")
	}
}

func handlePageInfo(db mysql.Database, jc *common.JobController, tp int, pi pageInfo) {
	defer func() {
		<-jc.Channel
		jc.WaitGroup.Done()
	}()

	pr, err := handleCrawler(tp, pi)
	if err != nil {
		return
	}

	if pr.content != "" {
		html := util.Html{Body: pr.content}
		html.Compress().Simplify()

		_, err = db.Execute(
			`INSERT INTO page_contents (page_id, content, registered_at) VALUES (?, ?, ?)`,
			pi.id, html.Body, time.Now().Unix(),
		)
		if err != nil {
			// TODO : 特殊文字の対応（server/clientの文字コードで対応）
			fmt.Println(errors.Wrap(err, ""))
		}
	}

	imgCap := util.Html{Body: pr.thumbnailCaption}
	imgCap.Compress().RemoveAllTags()
	pr.thumbnailCaption = imgCap.Body

	var hasThumb int
	if pi.thumbnailSelector.Status && pr.thumbnailUrl != "" {
		_, err = db.Execute(
			`INSERT INTO page_thumbnails (page_id, url, caption, copyright, registered_at)
			 VALUES (?, ?, ?, ?, ?)`,
			pi.id, pr.thumbnailUrl, pr.thumbnailCaption, pr.thumbnailCopyright, time.Now().Unix(),
		)
		if err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}

		hasThumb = 1
	}

	_, err = db.Execute(
		`UPDATE pages
		 SET
		  copyright     = ?,
		  has_thumbnail = ?,
		  is_scraped    = ?
		 WHERE id = ?`,
		pr.copyright, hasThumb, 1, pi.id,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}
}

func main() {
	numCpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numCpu)

	flag.Parse()

	tp := 0
	if *javascriptOption {
		tp = 1
	}

	db := mysql.Database{}
	if err := db.Connect(); err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}
	defer db.Close()

	wNum := config.HTTP_WORKER_LIMIT
	if tp == 1 {
		wNum = config.AGOUTI_WORKER_LIMIT
	}

	jc := common.JobController{
		Channel: make(chan int, wNum),
	}

	rows, err := db.Query(
		`SELECT
		  p.id,
		  p.url,
		  p.has_thumbnail,
		  f.thumbnail_attribute,
		  f.thumbnail_selector,
		  f.thumbnail_caption_selector,
		  f.thumbnail_copyright_selector,
		  f.content_selector,
		  f.copyright_selector
		 FROM       pages           AS p
		 INNER JOIN feeds           AS f  ON f.id = p.feed_id
		 WHERE p.is_scraped     = ?
		   AND f.crawler_status = ?
		   AND f.scraper_status = ?
		   AND f.do_need_js     = ?`,
		0, 1, 1, tp,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}

	for rows.Next() {
		var (
			id, hasThumb                                              int
			url, imgAttr, imgSel, imgCapSel, imgCrSel, contSel, crSel string
		)
		if err := rows.Scan(&id, &url, &hasThumb, &imgAttr, &imgSel, &imgCapSel, &imgCrSel, &contSel, &crSel); err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}

		status := false
		if hasThumb == 0 {
			status = true
		}
		is := common.OptionalString{
			String: imgSel,
			Status: status,
		}

		pi := pageInfo{
			id:                         id,
			url:                        url,
			thumbnailAttribute:         imgAttr,
			thumbnailSelector:          &is,
			thumbnailCaptionSelector:   (&common.OptionalString{}).Make(imgCapSel),
			thumbnailCopyrightSelector: (&common.OptionalString{}).Make(imgCrSel),
			contentSelector:            (&common.OptionalString{}).Make(contSel),
			copyrightSelector:          (&common.OptionalString{}).Make(crSel),
		}

		jc.Channel <- 1
		jc.WaitGroup.Add(1)
		go handlePageInfo(db, &jc, tp, pi)
	}

	jc.WaitGroup.Wait()
}
