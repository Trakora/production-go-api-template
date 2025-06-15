package health

import "time"

type Check struct {
	Status    string        `json:"Status"`
	Debug     bool          `env:"SERVER_DEBUG,default=false"`
	Timestamp time.Time     `json:"Timestamp"`
	Uptime    time.Duration `json:"Uptime"`
}
