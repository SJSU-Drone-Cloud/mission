package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Mission struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	MissionID   string             `json:"missionID" bson:"missionID"`
	DroneID     string             `json:"droneID" bson:"droneID"`
	DateCreated time.Time          `json:"dateCreated" bson:"dateCreated"`
	LastUpdated time.Time          `json:"lastUpdated" bson:"lastUpdated"`
	InProgress  bool               `json:"inProgress" bson:"inProgress"`
	Waypoints   []Coordinates      `json:"waypoints" bson:"waypoints"`
	Parameters  Parameters         `json:"parameters" bson:"parameters"`
}

type Parameters struct {
	Altimeter string  `json:"altimeter" bson:"altimeter"`
	Gyro      string  `json:"gyro" bson:"gyro"`
	Barometer string  `json:"barometer" bson:"barometer"`
	Lat       float64 `json:"lat" bson:"lat"`
	Lng       float64 `json:"lng" bson:"lng"`
	NumSats   int     `json:"numSats" bson:"numSats"`
	Voltage   string  `json:"voltage" bson:"voltage"`
}

func (p *Parameters) UnmarshalJSON(data []byte) error {
	var a map[string]string
	if err := json.Unmarshal(data, &a); err != nil {
		fmt.Println("custom parameters unmarshal error")
		return err
	}
	p.Altimeter = a["droneID"]
	p.Gyro = a["gyro"]
	p.Barometer = a["barometer"]
	lat, err := strconv.ParseFloat(a["lat"], 64)
	if err != nil {
		return err
	}
	lng, err := strconv.ParseFloat(a["lng"], 64)
	if err != nil {
		return err
	}
	p.Lat = lat
	p.Lng = lng

	ns, err := strconv.ParseInt(a["connected_sats"], 10, 64)
	if err != nil {
		return err
	}
	p.NumSats = int(ns)
	p.Voltage = a["voltage"]
	return nil
}

type CreateMission struct {
	DroneID   string        `json:"droneID" bson:"droneID"`
	Waypoints []Coordinates `json:"waypoints" bson:"waypoints"`
}

type UpdateMission struct {
	MissionID  string
	DroneID    string
	Parameters Parameters
}

type Coordinates struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

func (co *Coordinates) UnmarshalJSON(data []byte) error {
	var a map[string]string
	fmt.Println("coordinates")
	if err := json.Unmarshal(data, &a); err != nil {
		fmt.Println("Custom Coordinates unmarshal error")
		return err
	}
	lat, err := strconv.ParseFloat(a["lat"], 64)
	if err != nil {
		return err
	}
	lng, err := strconv.ParseFloat(a["lng"], 64)
	if err != nil {
		return err
	}
	co.Lat = lat
	co.Lng = lng
	return nil
}

// func (m *Mission) UnmarshalJSON(data []byte) error {
// 	var a map[string]string
// 	fmt.Println("registerDrone")
// 	if err := json.Unmarshal(data, &a); err != nil {
// 		fmt.Println("Custom mission unmarshal error")
// 		return err
// 	}
// 	m.Address = a["address"]
// 	lat, err := strconv.ParseFloat(a["lat"], 64)
// 	if err != nil {
// 		return err
// 	}
// 	lng, err := strconv.ParseFloat(a["lng"], 64)
// 	if err != nil {
// 		return err
// 	}
// 	rd.Lat = lat
// 	rd.Lng = lng
// 	return nil
// }
