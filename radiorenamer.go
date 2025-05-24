package radiorenamer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yyoshiki41/go-radiko"
)

var (
	currentAreaID string
	location      *time.Location
)

const (
	tz             = "Asia/Tokyo"
	datetimeLayout = "20060102150405"
)

func init() {
	var err error

	currentAreaID, err = radiko.AreaID()
	if err != nil {
		panic(err)
	}

	location, err = time.LoadLocation(tz)
	if err != nil {
		panic(err)
	}
}

func Run(filename string) {
	log.Println("filename:", filename)
	if !exists(filename) {
		log.Fatal("file doesn't exist")
	}
	if !isAudioFile(filename) {
		log.Fatal("invalid file format")
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	recordedAt, station, area := parse(filename)
	centeredTime := getCenteredTime(ctx, filename, recordedAt)
	log.Println("time", recordedAt)
	log.Println("time", centeredTime)
	// Radikoから番組情報を取得
	pg, station := getProgramInfo(ctx, centeredTime, station, area)
	tag := CreateTagFromPg(pg, station)

	metadata := tag.toFfMpegOption()
	output := filepath.Join(filepath.Dir(filename), tag.toFileName()+".m4a")
	err := PutM4aTag(ctx, filename, output, metadata)
	if err != nil {
		log.Fatal("cannot convert file", err)
	}
}

// parse parses yyyymmddhhmmss-<STATIONname>.m4a and returns following parameters.
//
//	recordedAt: recorded time
//	station: recorded station
func parse(filename string) (recordedAt time.Time, station string, area string) {
	fileNameWithoutExt := getFileNameWithoutExt(filename)
	elements := strings.Split(fileNameWithoutExt, "-")
	var start string
	if len(elements) == 2 {
		start = elements[0]
		station = elements[1]
		area = ""
	} else if len(elements) >= 3 {
		start = elements[0]
		station = strings.Join(elements[1:len(elements)-1], "-")
		area = elements[len(elements)-1]
	} else {
		log.Fatal("invalid filename format:", fileNameWithoutExt)
	}
	var err error
	recordedAt, err = time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		log.Fatalf(
			"Invalid start time format '%s': %s", start, err)
	}
	return recordedAt, station, area
}

func getCenteredTime(ctx context.Context, filename string, startTime time.Time) (centeredTime time.Time) {
	dur, err := Duration(ctx, filename)
	if err != nil {
		log.Fatal("could not get centered time", err)
	}
	return startTime.Add(time.Duration(dur/2) * time.Second)
}

func getFileNameWithoutExt(path string) string {
	// Fixed with a nice method given by mattn-san
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}

// getProgramInfo fetches radiko program information from radiko.jp
func getProgramInfo(ctx context.Context, targetTime time.Time, stationID string, area string) (*radiko.Prog, string) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	var recordedAreaID string
	if area == currentAreaID {
		recordedAreaID = currentAreaID
	} else {
		recordedAreaID = area
	}
	client, err := getClient(ctx, recordedAreaID)
	if err != nil {
		log.Fatalf(
			"Failed to construct a radiko Client: %s", err)
	}

	_, err = client.AuthorizeToken(ctx)
	if err != nil {
		log.Fatalf(
			"Failed to get auth_token: %s", err)
	}
	pg, err := client.GetWeeklyPrograms(ctx, stationID)
	if err != nil {
		ctxCancel()
		fmt.Printf(
			"Failed to get the program: %s", err)
	}
	startStr := targetTime.Format(datetimeLayout)
	intStart, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		log.Println("cannot convert targetTime", targetTime)
	}
	for _, prog := range pg[0].Progs.Progs {
		intFrom, err := strconv.ParseInt(prog.Ft, 10, 64)
		if err != nil {
			log.Println("cannot convert from", prog.Ft, err)

		}
		intTo, err := strconv.ParseInt(prog.To, 10, 64)
		if err != nil {
			log.Println("cannot convert to", prog.To, err)
		}
		if intStart >= intFrom && intStart < intTo {
			return &prog, pg[0].Name
		}
	}
	log.Fatal("Cannot found Program Information")
	return nil, stationID
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func isAudioFile(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".aac", ".mp4", ".m4a":
		return true
	default:
		return false
	}
}
