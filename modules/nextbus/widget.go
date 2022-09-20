package nextbus

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/logger"
	"github.com/wtfutil/wtf/view"
)

// Widget is the container for your module's data
type Widget struct {
	view.TextWidget

	settings *Settings
}

// NewWidget creates and returns an instance of Widget
func NewWidget(tviewApp *tview.Application, redrawChan chan bool, pages *tview.Pages, settings *Settings) *Widget {
	widget := Widget{
		TextWidget: view.NewTextWidget(tviewApp, redrawChan, pages, settings.common),

		settings: settings,
	}
	return &widget
}

/* -------------------- Exported Functions -------------------- */

// Refresh updates the onscreen contents of the widget
func (widget *Widget) Refresh() {
	// The last call should always be to the display function
	widget.display()
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) content() string {
	return getNextBus("ttc", "14646")
}

type AutoGenerated struct {
	Copyright   string      `json:"copyright"`
	Predictions Predictions `json:"predictions"`
}

type Prediction struct {
	AffectedByLayover string `json:"affectedByLayover"`
	Seconds           string `json:"seconds"`
	TripTag           string `json:"tripTag"`
	Minutes           string `json:"minutes"`
	IsDeparture       string `json:"isDeparture"`
	Block             string `json:"block"`
	DirTag            string `json:"dirTag"`
	Branch            string `json:"branch"`
	EpochTime         string `json:"epochTime"`
	Vehicle           string `json:"vehicle"`
}

type Direction struct {
	PredictionRaw json.RawMessage `json:"prediction"`
	Title         string          `json:"title"`
}

type Predictions struct {
	RouteTag    string    `json:"routeTag"`
	StopTag     string    `json:"stopTag"`
	RouteTitle  string    `json:"routeTitle"`
	AgencyTitle string    `json:"agencyTitle"`
	StopTitle   string    `json:"stopTitle"`
	Direction   Direction `json:"direction"`
}

func getNextBus(route string, stopID string) string {
	url := fmt.Sprintf("https://webservices.umoiq.com/service/publicJSONFeed?command=predictions&a=%s&stopId=%s", route, stopID)
	resp, err := http.Get(url)
	if err != nil {
		// TODO: add log msg
		logger.Log(fmt.Sprintf("Failed to make request to TTC. ERROR: %s", err))
		return "ERROR REQ"
	}
	body, readErr := io.ReadAll(resp.Body)

	if (readErr) != nil {
		// TODO: add log msg
		return "ERROR"
	}

	resp.Body.Close()

	var parsedResponse AutoGenerated

	// partial unmarshal, we don't have r.Predictions.Direction.PredictionRaw <- YET
	unmarshalError := json.Unmarshal(body, &parsedResponse)
	if unmarshalError != nil {
		// TODO: add log msg
		log.Fatal(err)
	}

	parseType := ""
	// hacky, try object parse first
	item := Prediction{}
	if err := json.Unmarshal(parsedResponse.Predictions.Direction.PredictionRaw, &item); err == nil {
		parseType = "object"
	}

	// if object parse failed, it probably means we have an array
	items := []Prediction{}
	if err := json.Unmarshal(parsedResponse.Predictions.Direction.PredictionRaw, &items); err == nil {
		parseType = "array"
	}

	// build the final string
	finalStr := ""
	if parseType == "array" {
		for _, itm := range items {
			// TODO: move to functions
			seconds, _ := strconv.Atoi(itm.Seconds)
			minutes, _ := strconv.Atoi(itm.Minutes)
			seconds = seconds % 60
			finalStr += fmt.Sprintf("%s [%02d:%02d] Bus: %s\n", parsedResponse.Predictions.RouteTitle, minutes, seconds, itm.Vehicle)
		}
	} else {
		// TODO: move to functions
		seconds, _ := strconv.Atoi(item.Seconds)
		minutes, _ := strconv.Atoi(item.Minutes)
		seconds = seconds % 60
		finalStr += fmt.Sprintf("%s [%02d:%02d] Bus: %s\n", parsedResponse.Predictions.RouteTitle, minutes, seconds, item.Vehicle)
	}

	return finalStr
}
func (widget *Widget) display() {
	widget.Redraw(func() (string, string, bool) {
		return widget.CommonSettings().Title, widget.content(), false
	})
}
