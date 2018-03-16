package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/sync/errgroup"
)

func handler() error {
	api := anaconda.NewTwitterApiWithCredentials(
		os.Getenv("TWITTER_OAUTH_TOKEN"),
		os.Getenv("TWITTER_OAUTH_TOKEN_SECRET"),
		os.Getenv("TWITTER_CONSUMER_KEY"),
		os.Getenv("TWITTER_CONSUMER_SECRET"))
	defer api.Close()

	userIDStrs := strings.Split(os.Getenv("TARGET_USER_IDS"), " ")
	userIDs := make([]int64, 0, len(userIDStrs))
	for _, part := range userIDStrs {
		id, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		userIDs = append(userIDs, id)
	}

	date := time.Now().AddDate(0, -1, 0)

	var eg errgroup.Group
	for _, userID := range userIDs {
		userID := userID

		eg.Go(func() error {
			v := url.Values{}
			v.Set("count", "200")
			v.Set("exclude_replies", "true")
			v.Set("include_rts", "false")
			v.Set("user_id", strconv.FormatInt(userID, 10))
			timeline, err := api.GetUserTimeline(v)
			if err != nil {
				return err
			}

			for _, status := range timeline {
				if status.Retweeted {
					continue
				}

				createdAt, err := status.CreatedAtTime()
				if err != nil {
					continue
				}
				if createdAt.Before(date) {
					continue
				}

				if _, err := api.Retweet(status.Id, false); err != nil {
					log.Println(err)
				}
			}
			return nil
		})
	}
	return eg.Wait()
}

func main() {
	lambda.Start(handler)
}
