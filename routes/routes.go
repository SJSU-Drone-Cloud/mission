package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/SJSU-Drone-Cloud/mission/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	User string
	Pass string
}

func getUUID() string {
	uid := strings.Replace(uuid.New().String(), "-", "", -1)
	fmt.Println("New UUID:", uid)
	return uid
}

func setupCors(w *http.ResponseWriter, req *http.Request) {
	fmt.Println(req.Header.Get("Origin"))
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CRSF-Token, Authorization")
}

func NewRouter() *mux.Router {
	r := mux.NewRouter()
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	user := os.Getenv("MONGOUSER")
	pass := os.Getenv("MONGOPASS")

	s := &MongoDB{User: user, Pass: pass}
	//fmt.Println("trackingip:", s.TrackingIP, "registryip:", s.RegistryIP)
	r.HandleFunc("/mission/create", s.missionCreateHandler).Methods("POST")
	r.HandleFunc("/mission/update/{missionID}", s.updateMissionHandler).Methods("PUT")
	r.HandleFunc("/mission/drone/{droneID}", s.getDroneMissionsHandler).Methods("GET")
	r.HandleFunc("/mission/{missionID}", s.getMissionHandler).Methods("GET")
	return r
}

//theoretically, this should be able to access a cache of memory somewhere
func (m *MongoDB) missionCreateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("mission create handler called")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err tph bodyread:", err)
		return
	}

	cm := &models.CreateMission{}

	err = json.Unmarshal(body, cm)
	if err != nil {
		fmt.Println("in create mission unmarshalling:", err)
		w.WriteHeader(500)
		w.Write([]byte("create mission unmarshal error"))
		return
	}

	mission := models.Mission{
		ID:          primitive.NewObjectID(),
		MissionID:   getUUID(),
		DroneID:     cm.DroneID,
		DateCreated: time.Now().UTC(),
		LastUpdated: time.Now().UTC(),
		InProgress:  true,
		Waypoints:   cm.Waypoints,
		Parameters: models.Parameters{
			Altimeter: "null",
			Gyro:      "null",
			Barometer: "null",
			Lat:       0.00,
			Lng:       0.00,
			NumSats:   0,
			Voltage:   "null",
		},
	}

	uri := "mongodb+srv://" + m.User + ":" + m.Pass + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	// stick this into the mongo db backend
	collection := client.Database("DronePlatform").Collection("mission")
	_, err = collection.InsertOne(ctx, mission)
	if err != nil {
		fmt.Println("error in insert")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	/*
		########################
		this portion is important because now we must implement sending the json to the drone itself and
		get its response back, however there are a few problems here. I do not have an IP so I would have to
		request it from the registry, and also even if we did have an ip, we would not be able to reach out to
		atharvs wifi so we hvae to pretend it is like that have drone send back data on a polling policy

		so for now, we presume that this portion is implemented
	*/
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte("Mission Successfully, missionID: " + mission.MissionID))
}

func (m *MongoDB) updateMissionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("mission update handler called")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("err tph bodyread:", err)
		return
	}

	um := &models.UpdateMission{}

	err = json.Unmarshal(body, um)
	if err != nil {
		fmt.Println("in update mission unmarshalling:", err)
		w.WriteHeader(500)
		w.Write([]byte("create mission unmarshal error"))
		return
	}

	uri := "mongodb+srv://" + m.User + ":" + m.Pass + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	// stick this into the mongo db backend
	collection := client.Database("DronePlatform").Collection("mission")
	filter := bson.D{{"missionID", um.MissionID}}
	update := bson.D{{"$set", bson.D{{"lastUpdated", time.Now().UTC()}, {"parameters", um.Parameters}}}} //will probably run into problems here
	fmt.Println("pushing to db")
	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("error in insert")
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(um)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte("success"))

}

func (d *MongoDB) getMissionHandler(w http.ResponseWriter, r *http.Request) {
	/*
		/mission/{missionID}
	*/
	uri := "mongodb+srv://" + d.User + ":" + d.Pass + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	params := mux.Vars(r)
	mID := params["missionID"]
	mission := &models.Mission{}
	//unmarshal body into a struct
	collection := client.Database("DronePlatform").Collection("mission")
	err = collection.FindOne(
		ctx,
		bson.D{{"missionID", mID}}).Decode(mission)
	if err != nil {
		fmt.Println("error findone")
		fmt.Println(err)
		return
	}

	jsn, err := json.Marshal(mission)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsn)
}

func (d *MongoDB) getDroneMissionsHandler(w http.ResponseWriter, r *http.Request) {
	/*
		/mission/drone/{droneID}
	*/
	uri := "mongodb+srv://" + d.User + ":" + d.Pass + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("uri:", uri)
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	params := mux.Vars(r)
	dID := params["droneID"]
	missions := []models.Mission{}
	//unmarshal body into a struct

	collection := client.Database("DronePlatform").Collection("mission")
	cur, err := collection.Find(
		ctx,
		bson.D{{"droneID", dID}})
	if err != nil {
		fmt.Println(cur)
		fmt.Println("error with cur")
		fmt.Println(err)
		return
	}

	for cur.Next(ctx) {
		d := models.Mission{}
		err = cur.Decode(&d)
		if err != nil {
			fmt.Println(err)
			return
		}
		missions = append(missions, d)
	}
	cur.Close(ctx)
	if len(missions) == 0 {
		w.WriteHeader(500)
		w.Write([]byte("No data found."))
		return
	}
	jsn, err := json.Marshal(missions)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsn)
}

// func (s *Services) redirectRequestBody(host string, endpoint string)

// func (db *DroneDB) droneHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("in index handler")
// 	setupCors(&w, r)
// 	if r.Method == "OPTIONS" {
// 		return
// 	}
// 	clientOptions := options.Client().
// 		ApplyURI("mongodb+srv://thunderpurtz:" + db.password + "@cluster0.14i4y.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	client, err := mongo.Connect(ctx, clientOptions)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer func() {
// 		if err = client.Disconnect(ctx); err != nil {
// 			panic(err)
// 		}
// 	}()
// 	collection := client.Database("DronePlatform").Collection("trackingData")
// 	cur, err := collection.Find(ctx, bson.D{})
// 	if err != nil {
// 		fmt.Println(cur)
// 		fmt.Println("error with cur")
// 		fmt.Println(err)
// 		return
// 	}
// 	drones := []models.Drone{}

// 	for cur.Next(ctx) {
// 		d := models.Drone{}
// 		err = cur.Decode(&d)
// 		fmt.Println("d lat:", d.Coordinates.Lat, ":d lng:", d.Coordinates.Lng)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		drones = append(drones, d)
// 	}
// 	cur.Close(ctx)
// 	if len(drones) == 0 {
// 		w.WriteHeader(500)
// 		w.Write([]byte("No data found."))
// 		return
// 	}
// 	jsn, err := json.Marshal(drones)
// 	fmt.Println("jsn:", jsn)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(200)
// 	w.Write(jsn)
// }

// func getUUID() string {
// 	uid := strings.Replace(uuid.New().String(), "-", "", -1)
// 	fmt.Println("New UUID:", uid)
// 	return uid
// }

// func userGetHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("userget" + strconv.Itoa(globalcount))
// 	globalcount += 1
// 	session, _ := sessions.Store.Get(r, "session")
// 	untypedUserId := session.Values["user_id"]
// 	currentUserId, ok := untypedUserId.(int64)
// 	fmt.Println(currentUserId)
// 	if !ok {
// 		utils.InternalServerError(w)
// 		return
// 	}
// 	vars := mux.Vars(r) //hashmap of variable names and content passed for that variable
// 	username := vars["username"]
// 	fmt.Println("username", username)

// 	currentPageUserString := strings.TrimLeft(r.URL.Path, "/")
// 	currentPageUser, err := models.GetUserByUsername(currentPageUserString)
// 	if err != nil {
// 		utils.InternalServerError(w)
// 		return
// 	}
// 	currentPageUserID, err := currentPageUser.GetId()
// 	if err != nil {
// 		utils.InternalServerError(w)
// 		return
// 	}
// 	updates, err := models.GetUpdates(currentPageUserID)
// 	if err != nil {
// 		utils.InternalServerError(w)
// 		return
// 	}

// 	utils.ExecuteTemplate(w, "index.html", struct {
// 		Title       string
// 		Updates     []*models.Update
// 		DisplayForm bool
// 	}{
// 		Title:       username,
// 		Updates:     updates,
// 		DisplayForm: currentPageUserID == currentUserId,
// 	})

// }

// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	updates, err := models.GetAllUpdates()
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 		return
// 	}
// 	utils.ExecuteTemplate(w, "index.html", struct {
// 		Title       string
// 		Updates     []*models.Update
// 		DisplayForm bool
// 	}{
// 		Title:       "All updates",
// 		Updates:     updates,
// 		DisplayForm: true,
// 	})
// 	fmt.Println("get")
// }

// func postHandlerHelper(w http.ResponseWriter, r *http.Request) error {
// 	session, _ := sessions.Store.Get(r, "session")
// 	untypedUserID := session.Values["user_id"]
// 	userID, ok := untypedUserID.(int64)
// 	if !ok {
// 		return utils.InternalServer
// 	}
// 	currentPageUserString := strings.TrimLeft(r.URL.Path, "/")
// 	currentPageUser, err := models.GetUserByUsername(currentPageUserString)
// 	if err != nil {
// 		return utils.InternalServer
// 	}
// 	currentPageUserID, err := currentPageUser.GetId()
// 	if err != nil {
// 		return utils.InternalServer
// 	}
// 	if currentPageUserID != userID {
// 		return utils.BadPostError
// 	}
// 	r.ParseForm()
// 	body := r.PostForm.Get("adddrone")
// 	fmt.Println(body)
// 	err = models.PostUpdates(userID, body)
// 	if err != nil {
// 		return utils.InternalServer
// 	}
// 	return nil
// }

// func postHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("post handler called")
// 	err := postHandlerHelper(w, r)
// 	if err == utils.InternalServer {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 	}
// 	http.Redirect(w, r, "/", 302)
// }

// func UserPostHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println("user post handler called")
// 	fmt.Println(r.URL.Path)
// 	err := postHandlerHelper(w, r)
// 	if err == utils.BadPostError {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte("Cannot write to another user's page"))
// 	}
// 	http.Redirect(w, r, r.URL.Path, 302)
// }

// func loginGetHandler(w http.ResponseWriter, r *http.Request) {
// 	utils.ExecuteTemplate(w, "login.html", nil)
// }

// func loginPostHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	username := r.PostForm.Get("username")
// 	password := r.PostForm.Get("password")

// 	user, err := models.AuthenticateUser(username, password)
// 	if err != nil {
// 		switch err {
// 		case models.InvalidLogin:
// 			utils.ExecuteTemplate(w, "login.html", "User or Pass Incorrect")
// 		default:
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte("Internal server error"))
// 		}
// 		return
// 	}
// 	userId, err := user.GetId()
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 		return
// 	}
// 	sessions.GetSession(w, r, "session", userId)
// 	http.Redirect(w, r, "/", 302)
// }

// func logoutGetHandler(w http.ResponseWriter, r *http.Request) {
// 	sessions.EndSession(w, r)
// 	http.Redirect(w, r, "/login", 302)
// }

// func registerGetHandler(w http.ResponseWriter, r *http.Request) {
// 	utils.ExecuteTemplate(w, "register.html", nil)
// }

// func registerPostHandler(w http.ResponseWriter, r *http.Request) {
// 	r.ParseForm()
// 	username := r.PostForm.Get("username")
// 	password := r.PostForm.Get("password")
// 	err := models.RegisterUser(username, password)
// 	if err == models.UserNameTaken {
// 		utils.ExecuteTemplate(w, "register.html", "username taken")
// 		return
// 	}
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write([]byte("Internal server error"))
// 		return
// 	}
// 	http.Redirect(w, r, "/login", 302)
// }
