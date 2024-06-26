package widget

import (
	"context"
	"html/template"
	"time"

	"github.com/glanceapp/glance/internal/assets"
	"github.com/glanceapp/glance/internal/feed"
)

type Videos struct {
	widgetBase `yaml:",inline"`
	Videos     feed.Videos `yaml:"-"`
	Channels   []string    `yaml:"channels"`
	Limit      int         `yaml:"limit"`
}

func (widget *Videos) Initialize() error {
	widget.withTitle("Videos").withCacheDuration(time.Hour)

	if widget.Limit <= 0 {
		widget.Limit = 25
	}

	return nil
}

func (widget *Videos) Update(ctx context.Context) {
	videos, err := feed.FetchYoutubeChannelUploads(widget.Channels)

	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}

	if len(videos) > widget.Limit {
		videos = videos[:widget.Limit]
	}

	widget.Videos = videos
}

func (widget *Videos) Render() template.HTML {
	return widget.render(widget, assets.VideosTemplate)
}
