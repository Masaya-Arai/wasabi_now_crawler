package main

import (
	"log"
	"mysql"
	"time"
	"util"
	"github.com/pkg/errors"
	_ "github.com/go-sql-driver/mysql"
)

/*
 * ・重複する記事の除外（削除ではない）
 * ・問題ない記事をarticlesに昇格
 * ・AP通信の画像はauthTokenをrefreshしないと見れなくなるので、暫定的に２日経過したら画像レコードを消去（再取得しない）
 */
func main() {
	// MySQLへ接続
	db := mysql.Database{}
	if err := db.Connect(); err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}

	var now int64 = time.Now().Unix()
	defer func() {
		db.Execute(`update queues set queue = ? where id = ?`, now, 1)
		defer db.Close()
	}()


	// queue取得
	a := db.QueryRow(
		`select
		  queue
		 from queues
		 where id = ?`,
		1,
	)
	var q int
	a.Scan(&q)

	// 対象記事を取得
	rows, err := db.Query(
		`select
		  p.id,
		  f.site_id,
		  p.title
		 from       pages         as p
		 inner join page_contents as pc on p.id = pc.page_id
		 inner join feeds         as f  on f.id = p.feed_id
		 where p.registered_at > ?
		   and char_length(pc.content) > ?`,
		q, 500,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, ""))
	}

	for rows.Next() {
		var (
			id, siteId int
			title string
		)
		if err := rows.Scan(&id, &siteId, &title); err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}

		// 前日分から重複がないか確認
		check := db.QueryRow(
			`select
			  count(a.id)
			 from       articles as a
			 inner join pages    as p on p.id = a.page_id
			 inner join feeds    as f on f.id = p.feed_id
			 where p.registered_at > ?
			   and p.title = ?`,
			q-86400*2, title,
		)
		var aid int
		if err = check.Scan(&aid); err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}
		if aid > 0 {
			continue
		}

		// 重複なければarticle登録
		res, err := db.Query(
			`insert into articles (page_id, hash_code, registered_at) values (?, ?, ?)`,
			id, util.SecureRandom(16), time.Now().Unix(),
		)
		if err != nil {
			log.Fatal(errors.Wrap(err, ""))
		}
		res.Close()
	}

	db.Execute(
		`update pages as p
		 inner join page_thumbnails as pt on p.id = pt.page_id
		 inner join feeds           as f  on f.id = p.feed_id
		 set p.has_thumbnail = ?
		 where pt.registered_at < ?
		   and f.site_id        = ?`,
		0, now-86400*1, 2,
	)

	db.Execute(
		`UPDATE pages AS p
                 INNER JOIN feeds AS f ON f.id = p.feed_id
                 INNER JOIN sites AS s ON s.id = f.site_id
                 SET p.has_thumbnail = ?
                 WHERE s.id = ?`,
		0, 18,
	)
}
