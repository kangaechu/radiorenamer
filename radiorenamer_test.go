package radiorenamer

import (
	"reflect"
	"testing"
	"time"
)

func Test_parse(t *testing.T) {
	location, _ := time.LoadLocation(tz)

	type args struct {
		filename string
	}
	tests := []struct {
		name           string
		args           args
		wantRecordedAt time.Time
		wantStation    string
		wantArea       string
	}{
		{
			name:           "no areaID",
			args:           args{filename: "20220529140000-TBS.aac"},
			wantRecordedAt: time.Date(2022, 5, 29, 14, 0, 0, 0, location),
			wantStation:    "TBS",
			wantArea:       "",
		},
		{
			name:           "with areaID",
			args:           args{filename: "20220529140000-CBC-JP23.aac"},
			wantRecordedAt: time.Date(2022, 5, 29, 14, 0, 0, 0, location),
			wantStation:    "CBC",
			wantArea:       "JP23",
		},
		{
			name:           "station includes hyphen",
			args:           args{filename: "20220529140000-ALPHA-STATION-JP26.aac"},
			wantRecordedAt: time.Date(2022, 5, 29, 14, 0, 0, 0, location),
			wantStation:    "ALPHA-STATION",
			wantArea:       "JP26",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRecordedAt, gotStation, gotArea := parse(tt.args.filename)
			if !reflect.DeepEqual(gotRecordedAt, tt.wantRecordedAt) {
				t.Errorf("parse() gotRecordedAt = %v, want %v", gotRecordedAt, tt.wantRecordedAt)
			}
			if gotStation != tt.wantStation {
				t.Errorf("parse() gotStation = %v, want %v", gotStation, tt.wantStation)
			}
			if gotArea != tt.wantArea {
				t.Errorf("parse() gotArea = %v, want %v", gotArea, tt.wantArea)
			}
		})
	}
}
