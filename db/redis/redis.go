package redis

import (
	"github.com/seeker-insurance/kit/config"
	"github.com/seeker-insurance/kit/log"
	r "github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

var Client *r.Client

func init() {
	cobra.OnInitialize(connectDB)
}

func connectDB() {
	url := config.RequiredString("redis_url")
	opts, err := r.ParseURL(url)
	log.Check(err)
	Client = r.NewClient(opts)
}
