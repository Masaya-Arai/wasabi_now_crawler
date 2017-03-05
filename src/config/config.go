package config

import "time"

const AGOUTI_WORKER_LIMIT int = 5
const HTTP_WORKER_LIMIT int = 10

const AGOUTI_TIMEOUT_SECONDS time.Duration = 60
const HTTP_TIMEOUT_SECONDS time.Duration = 10

const MINIMUM_INTERVAL_SECONDS time.Duration = 1

const AGOUTI_PAGE_WIDTH int = 769
const AGOUTI_PAGE_HEIGHT int = 1

const USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.36"