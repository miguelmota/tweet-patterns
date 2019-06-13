package client

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// Config ...
type Config struct {
	Username          string
	ConsumerKey       string
	ConsumerSecret    string
	AccessTokenKey    string
	AccessTokenSecret string
}

// Client ...
type Client struct {
	username string
	tc       *twitter.Client
}

// NewClient ...
func NewClient(config *Config) *Client {
	if config == nil {
		panic(errors.New("Config is required"))
	}

	cf := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.AccessTokenKey, config.AccessTokenSecret)
	httpClient := cf.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)
	return &Client{
		tc:       client,
		username: config.Username,
	}
}

// Save ...
func (c *Client) Save() (string, error) {
	var tweets []twitter.Tweet
	var maxID int64
	for i := 0; i < 10; i++ {
		ts, err := c.fetchTweets(maxID)
		if err != nil {
			return "", err
		}

		tweets = append(tweets, ts...)
		maxID = tweets[len(tweets)-1].ID
	}

	data := make(plotter.XYZs, len(tweets))

	for _, tweet := range tweets {
		t, err := time.Parse(time.UnixDate, tweet.CreatedAt)
		if err != nil {
			return "", err
		}

		l, err := time.LoadLocation("America/Los_Angeles")
		if err != nil {
			panic(err)
		}
		t = t.In(l)

		day := t.Weekday()
		hour := t.Hour()
		likes := tweet.FavoriteCount

		data = append(data, plotter.XYZ{
			X: float64(day),
			Y: float64(hour),
			Z: float64(likes),
		})
	}

	minZ, maxZ := math.Inf(1), math.Inf(-1)
	for _, xyz := range data {
		if xyz.Z > maxZ {
			maxZ = xyz.Z
		}
		if xyz.Z < minZ {
			minZ = xyz.Z
		}
	}

	p, err := plot.New()
	if err != nil {
		return "", err
	}

	p.Title.Text = fmt.Sprintf("%s latest %d tweet likes", c.username, len(tweets))
	p.X.Label.Text = "Weekday"
	p.Y.Label.Text = "Hour"
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{{0, "Sun"}, {1, "Mon"}, {2, "Tue"}, {3, "Wed"}, {4, "Thu"}, {5, "Fri"}, {6, "Sat"}, {7, "Sun"}})

	yticks := make([]plot.Tick, 24)
	for i := 0; i < 24; i++ {
		yticks[i] = plot.Tick{float64(i), fmt.Sprintf("%d", i+1)}
	}
	p.Y.Tick.Marker = plot.ConstantTicks(yticks)

	sc, err := plotter.NewScatter(data)
	if err != nil {
		return "", err
	}

	sc.GlyphStyleFunc = func(i int) draw.GlyphStyle {
		c := color.RGBA{R: 28, G: 161, B: 242, A: 1}
		var minRadius, maxRadius = vg.Points(1), vg.Points(20)
		rng := maxRadius - minRadius
		_, _, z := data.XYZ(i)
		d := (z - minZ) / (maxZ - minZ)
		r := vg.Length(d)*rng + minRadius
		return draw.GlyphStyle{Color: c, Radius: r, Shape: draw.CircleGlyph{}}
	}

	p.Add(sc)

	filename := fmt.Sprintf("%s.png", c.username)
	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		return "", err
	}

	return filename, nil
}

func (c *Client) fetchTweets(maxID int64) ([]twitter.Tweet, error) {
	excludeReplies := true
	includeRetweets := false
	trimUser := true

	tweets, _, err := c.tc.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName:      c.username,
		Count:           200,
		ExcludeReplies:  &excludeReplies,
		IncludeRetweets: &includeRetweets,
		TrimUser:        &trimUser,
		MaxID:           maxID,
	})
	if err != nil {
		return nil, err
	}

	return tweets, nil
}
