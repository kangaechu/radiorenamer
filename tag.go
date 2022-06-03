package radiorenamer

import (
	"log"
	"strings"
	"time"

	"github.com/yyoshiki41/go-radiko"
)

type Tag struct {
	Title   string
	Title2  string
	Artist  string
	PubDate time.Time
	Comment string
}

func CreateTagFromPg(pg *radiko.Prog, station string) Tag {
	start, err := time.Parse("20060102150405", pg.Ft)
	if err != nil {
		log.Println("cannot parse date", pg.Ft)
		start = time.Unix(0, 0)
	}
	return Tag{pg.Title, "", station, start, pg.Info}
}

func (t Tag) toFfMpegOption() []string {
	metadata := []string{
		"-metadata", "title=" + strings.TrimSpace(t.Title+" "+t.PubDate.Format("2006年01月02日")+" "+t.Title2),
		"-metadata", "genre=Radio",
		"-metadata", "artist=" + t.Artist,
		"-metadata", "comment=" + t.Comment,
	}
	return metadata
}

func (t Tag) toFileName() string {
	return strings.TrimSpace(t.Title + " " + t.PubDate.Format("2006年01月02日") + " " + t.Title2)
}
