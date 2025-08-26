package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

var spaceApiData = &SpaceAPIv15{
	APICompatibility: []string{"14", "15"},
	Space:            "Metalab",
	Logo:             "https://metalab.at/wiki/images/9/93/Metalab.at.svg",
	URL:              "https://metalab.at",
	Location: &Location{
		Address:     "Verein Metalab, Rathausstraße 6, 1010 Wien, Austria",
		Lat:         48.2093723,
		Lon:         16.356099,
		Timezone:    "Europe/Vienna",
		CountryCode: "AT",
	},
	SpaceFed: &SpaceFed{
		SpaceNet:  false,
		SpaceSAML: false,
	},
	State: &State{
		Open: nil,
	},
	Contact: &Contact{
		Phone:    "+43 720 002323",
		Mastodon: "@metalab@chaos.social",
		SIP:      "6382",
	},
	Links: []Link{
		{
			Name: "Metalab Wiki",
			URL:  "https://metalab.at/wiki",
		},
	},
	Projects: []string{
		"https://github.com/metalab",
		"https://metalab.at/wiki/Projekte_Neu",
	},
}

type LabStatusAPIResponse struct {
	State           string `json:"state"`
	LastChangedUnix int64  `json:"last_changed"`
	LastUpdatedUnix int64  `json:"last_updated"`
}

func Pointer[T any](d T) *T {
	return &d
}

func handleSpaceApiV15(w http.ResponseWriter, r *http.Request) {
	labState, labStateLastChange, labStateError := fetchLabState()
	if labStateError != nil {
		http.Error(w, labStateError.Error(), http.StatusInternalServerError)
		return
	} else {
		spaceApiData.State.Open = labState
		if labStateLastChange != nil {
			spaceApiData.State.LastChange = *labStateLastChange
		}
	}
	p, _ := json.Marshal(spaceApiData)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(p)
}

func fetchLabState() (*bool, *int64, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://eingang.metalab.at/status.json", nil)

	//req, err := http.NewRequest("GET", "http://localhost:3333/lab", nil)
	if err != nil {
		fmt.Printf("error while building rest request to state api: %v\n", err)
		return nil, nil, err
	}

	//set required header
	req.Header.Set("Content-Type", "application/json")

	//actually send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error while sending request to state api: %v\n", err)
		return nil, nil, err
	}

	//close the request and read the body
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error while reading response body from state api: %v\n", err)
		return nil, nil, err
	}

	/*var r LabStatusAPIResponse
	json.Unmarshal(body, &r)
	if r.State == "on" {
		return Pointer(true), &r.LastChangedUnix, nil
	} else if r.State == "off" {
		return Pointer(false), &r.LastChangedUnix, nil
	} else {
		return nil, nil, fmt.Errorf("unknown state: %s", r.State)
	}*/

	type LabStatus struct {
		Status string `json:"status"`
	}

	var r LabStatus
	json.Unmarshal(body, &r)

	if r.Status == "open" {
		return Pointer(true), nil, nil
	} else if r.Status == "closed" {
		return Pointer(false), nil, nil
	} else {
		return nil, nil, fmt.Errorf("unknown state: %s", r.Status)
	}

}

func main() {
	http.HandleFunc("/v14", handleSpaceApiV15) //v14 is also compatible with v15
	http.HandleFunc("/v15", handleSpaceApiV15)

	fmt.Println("Server starting on port 3334...")
	if err := http.ListenAndServe(":3334", nil); err != nil {
		log.Fatal(err)
	}
}

// SpaceAPIv15 represents the main SpaceAPI v15 structure
type SpaceAPIv15 struct {
	APICompatibility []string   `json:"api_compatibility"`
	Space            string     `json:"space"`
	Logo             string     `json:"logo,omitempty"`
	URL              string     `json:"url,omitempty"`
	Location         *Location  `json:"location,omitempty"`
	SpaceFed         *SpaceFed  `json:"spacefed,omitempty"`
	Cam              []string   `json:"cam,omitempty"`
	State            *State     `json:"state,omitempty"`
	Events           []Event    `json:"events,omitempty"`
	Contact          *Contact   `json:"contact,omitempty"`
	Sensors          *Sensors   `json:"sensors,omitempty"`
	Feeds            *Feeds     `json:"feeds,omitempty"`
	Links            []Link     `json:"links,omitempty"`
	Cache            *Cache     `json:"cache,omitempty"`
	Projects         []string   `json:"projects,omitempty"`
	RadioShow        *RadioShow `json:"radio_show,omitempty"`
}

// Location represents the physical location of the space
type Location struct {
	Address     string  `json:"address,omitempty"`
	Lat         float64 `json:"lat,omitempty"`
	Lon         float64 `json:"lon,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
	CountryCode string  `json:"country_code,omitempty"`
	Hint        string  `json:"hint,omitempty"`
	Areas       []Area  `json:"areas,omitempty"`
}

// Area represents a physical area within the space
type Area struct {
	Name         string  `json:"name,omitempty"`
	Description  string  `json:"description,omitempty"`
	SquareMeters float64 `json:"square_meters"` // Required
}

// SpaceFed represents SpaceFED authentication information
type SpaceFed struct {
	SpaceNet  bool `json:"spacenet"`  // Required
	SpaceSAML bool `json:"spacesaml"` // Required
}

// State represents the current state of the space
type State struct {
	Open          *bool      `json:"open"`
	LastChange    int64      `json:"lastchange,omitempty"`
	TriggerPerson string     `json:"trigger_person,omitempty"`
	Message       string     `json:"message,omitempty"`
	Icon          *StateIcon `json:"icon,omitempty"`
}

// StateIcon represents the URLs for state icons
type StateIcon struct {
	Open   string `json:"open"`   // Required
	Closed string `json:"closed"` // Required
}

// Event represents an event in the space
type Event struct {
	Name      string `json:"name"`      // Required
	Type      string `json:"type"`      // Required
	Timestamp int64  `json:"timestamp"` // Required
	Extra     string `json:"extra,omitempty"`
}

// Contact contains various contact methods for the space
type Contact struct {
	Phone      string      `json:"phone,omitempty"`
	SIP        string      `json:"sip,omitempty"`
	Keymasters []Keymaster `json:"keymasters,omitempty"`
	IRC        string      `json:"irc,omitempty"`
	Twitter    string      `json:"twitter,omitempty"`
	Mastodon   string      `json:"mastodon,omitempty"`
	Facebook   string      `json:"facebook,omitempty"`
	Identica   string      `json:"identica,omitempty"`
	Foursquare string      `json:"foursquare,omitempty"`
	Email      string      `json:"email,omitempty"`
	ML         string      `json:"ml,omitempty"`
	XMPP       string      `json:"xmpp,omitempty"`
	IssueMail  string      `json:"issue_mail,omitempty"`
	Gopher     string      `json:"gopher,omitempty"`
	Matrix     string      `json:"matrix,omitempty"`
	Mumble     string      `json:"mumble,omitempty"`
}

// Keymaster represents a person who has access to the space
type Keymaster struct {
	Name     string `json:"name,omitempty"`
	IRCNick  string `json:"irc_nick,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Email    string `json:"email,omitempty"`
	Twitter  string `json:"twitter,omitempty"`
	XMPP     string `json:"xmpp,omitempty"`
	Mastodon string `json:"mastodon,omitempty"`
	Matrix   string `json:"matrix,omitempty"`
}

// Sensors represent various sensor data in the space
type Sensors struct {
	Temperature    []TempSensor      `json:"temperature,omitempty"`
	CarbonDioxide  []CO2Sensor       `json:"carbondioxide,omitempty"`
	DoorLocked     []DoorSensor      `json:"door_locked,omitempty"`
	Barometer      []BarometerSensor `json:"barometer,omitempty"`
	Radiation      *RadiationSensors `json:"radiation,omitempty"`
	Humidity       []HumiditySensor  `json:"humidity,omitempty"`
	BeverageSupply []BeverageSensor  `json:"beverage_supply,omitempty"`
}

// BaseSensor contains common sensor fields
type BaseSensor struct {
	Location    string `json:"location"` // Required
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	LastChange  int64  `json:"lastchange,omitempty"`
}

// TempSensor represents a temperature sensor
type TempSensor struct {
	BaseSensor
	Value float64 `json:"value"` // Required
	Unit  string  `json:"unit"`  // Required
}

// CO2Sensor represents a CO2 sensor
type CO2Sensor struct {
	BaseSensor
	Value float64 `json:"value"` // Required
	Unit  string  `json:"unit"`  // Required
}

// DoorSensor represents a door lock sensor
type DoorSensor struct {
	BaseSensor
	Value bool `json:"value"` // Required
}

// BarometerSensor represents a barometer sensor
type BarometerSensor struct {
	BaseSensor
	Value float64 `json:"value"` // Required
	Unit  string  `json:"unit"`  // Required
}

// RadiationSensors represents all radiation sensor types
type RadiationSensors struct {
	Alpha     []RadiationSensor `json:"alpha,omitempty"`
	Beta      []RadiationSensor `json:"beta,omitempty"`
	Gamma     []RadiationSensor `json:"gamma,omitempty"`
	BetaGamma []RadiationSensor `json:"beta_gamma,omitempty"`
}

// RadiationSensor represents a radiation sensor
type RadiationSensor struct {
	BaseSensor
	Value            float64 `json:"value"` // Required
	Unit             string  `json:"unit"`  // Required
	DeadTime         float64 `json:"dead_time,omitempty"`
	ConversionFactor float64 `json:"conversion_factor,omitempty"`
}

// HumiditySensor represents a humidity sensor
type HumiditySensor struct {
	BaseSensor
	Value float64 `json:"value"` // Required
	Unit  string  `json:"unit"`  // Required
}

// BeverageSensor represents a beverage supply sensor
type BeverageSensor struct {
	BaseSensor
	Value float64 `json:"value"` // Required
	Unit  string  `json:"unit"`  // Required
}

// Feeds represents various feeds available for the space
type Feeds struct {
	Blog     *Feed `json:"blog,omitempty"`
	Wiki     *Feed `json:"wiki,omitempty"`
	Calendar *Feed `json:"calendar,omitempty"`
	Flickr   *Feed `json:"flickr,omitempty"`
}

// Feed represents a generic feed URL with type
type Feed struct {
	Type string `json:"type,omitempty"` // Type of the feed (e.g., rss, ical, atom)
	URL  string `json:"url"`            // Required
}

// Link represents external links related to the space
type Link struct {
	Name        string `json:"name"` // Required
	Description string `json:"description,omitempty"`
	URL         string `json:"url"` // Required
}

// Cache represents caching information
type Cache struct {
	Schedule string `json:"schedule"` // Required - cron-like schedule string
}

// RadioShow represents information about the space's radio show
type RadioShow struct {
	Name        string   `json:"name"`                 // Required
	URL         string   `json:"url"`                  // Required
	Type        string   `json:"type"`                 // Required
	StartTime   string   `json:"start_time,omitempty"` // ISO 8601 formatted time
	EndTime     string   `json:"end_time,omitempty"`   // ISO 8601 formatted time
	StreamURL   string   `json:"stream_url,omitempty"`
	StreamType  string   `json:"stream_type,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}
