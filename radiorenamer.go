package radiorenamer

import (
	"context"
	"fmt"
	"github.com/yyoshiki41/go-radiko"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
		os.Exit(1)
	}
	if !isAudioFile(filename) {
		log.Fatal("invalid file format")
		os.Exit(1)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	recordedAt, station := parse(filename)
	centeredTime := getCenteredTime(ctx, filename, recordedAt)
	log.Println("time", recordedAt)
	log.Println("time", centeredTime)
	pg, station := getProgramInfo(ctx, centeredTime, station)

	tag := CreateTagFromPg(pg, station)
	metadata := tag.toFfMpegOption()
	output := filepath.Join(filepath.Dir(filename), tag.toFileName()+".m4a")
	err := PutM4aTag(ctx, filename, output, metadata)
	if err != nil {
		log.Fatal("cannot convert file", err)
	}
}

// parse parses yyyymmddhhmmss-<STATIONname>.m4a and returns following parameters.
// 	recordedAt: recorded time
//	station: recorded station
func parse(filename string) (recordedAt time.Time, station string) {
	fileNameWithoutExt := getFileNameWithoutExt(filename)
	elements := strings.Split(fileNameWithoutExt, "-")
	if len(elements) != 2 {
		log.Fatal("invalid filename format:", fileNameWithoutExt)
		os.Exit(1)
	}
	start := elements[0]
	station = elements[1]
	var err error
	recordedAt, err = time.ParseInLocation(datetimeLayout, start, location)
	if err != nil {
		fmt.Println(
			"Invalid start time format '%s': %s", start, err)
		os.Exit(1)
	}
	return recordedAt, station
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
func getProgramInfo(ctx context.Context, targetTime time.Time, stationID string) (*radiko.Prog, string) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	client, err := getClient(ctx, currentAreaID)
	if err != nil {
		fmt.Printf(
			"Failed to construct a radiko Client: %s", err)
		os.Exit(1)
	}

	_, err = client.AuthorizeToken(ctx)
	if err != nil {
		fmt.Printf(
			"Failed to get auth_token: %s", err)
		os.Exit(1)
	}
	pg, err := client.GetWeeklyPrograms(ctx, stationID)
	if err != nil {
		ctxCancel()
		fmt.Printf(
			"Failed to get the program: %s", err)
	}
	startStr := targetTime.Format(datetimeLayout)
	intStart, err := strconv.Atoi(startStr)
	if err != nil {
		log.Println("cannot convert targetTime", targetTime)
	}
	for _, prog := range pg[0].Progs.Progs {
		intFrom, err := strconv.Atoi(prog.Ft)
		if err != nil {
			log.Println("cannot convert from", prog.Ft)

		}
		intTo, err := strconv.Atoi(prog.To)
		if err != nil {
			log.Println("cannot convert from", prog.To)
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
