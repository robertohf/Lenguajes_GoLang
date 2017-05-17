package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/image/bmp"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kr/pretty"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

type Route struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

type Image struct {
	Nombre string `json:"nombre"`
	Tamaño Tamaño `json:tamaño`
}

type Tamaño struct {
	Alto  int `json:alto`
	Ancho int `json:ancho`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ejercicio1", route).Methods("POST")
	router.HandleFunc("/ejercicio2", restaurantList).Methods("POST")
	router.HandleFunc("/ejercicio3", grayScaling).Methods("POST")
	router.HandleFunc("/ejercicio4", redux).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func route(w http.ResponseWriter, req *http.Request) {

	var route Route
	_ = json.NewDecoder(req.Body).Decode(&route)

	client, err := maps.NewClient(maps.WithAPIKey("AIzaSyBmelZAhVTODrw_gjtueTuHEs9Aka_z9nM"))
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}

	r := &maps.DirectionsRequest{
		Origin:      route.Origin,
		Destination: route.Destination,
	}

	routes, _, err := client.Directions(context.Background(), r)
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}

	buffer := new(bytes.Buffer)
	buffer.WriteString("{\"routes\":[")

	json.NewDecoder(req.Body).Decode(&routes)

	for x := 0; x < len(routes[0].Legs[0].Steps); x++ {
		buffer.WriteString("{\"lat\":")
		buffer.WriteString(strconv.FormatFloat(routes[0].Legs[0].Steps[x].StartLocation.Lat, 'f', 5, 64))
		buffer.WriteString(", ")
		buffer.WriteString("\"lon\":")
		buffer.WriteString(strconv.FormatFloat(routes[0].Legs[0].Steps[x].StartLocation.Lng, 'f', 5, 64))
		buffer.WriteString("} ")
	}

	buffer.WriteString("]}")
	fmt.Fprintf(w, buffer.String())
}

func restaurantList(w http.ResponseWriter, req *http.Request) {

	var place Route
	_ = json.NewDecoder(req.Body).Decode(&place)

	client, err := maps.NewClient(maps.WithAPIKey("AIzaSyBmelZAhVTODrw_gjtueTuHEs9Aka_z9nM"))
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}

	origin_detail := &maps.GeocodingRequest{
		Address: place.Origin,
	}

	origin_response, _ := client.Geocode(context.Background(), origin_detail)

	r := &maps.NearbySearchRequest{

		Location: &origin_response[0].Geometry.Location,
		Radius:   100,
		Type:     maps.PlaceTypeRestaurant,
	}

	places, _ := client.NearbySearch(context.Background(), r)
	json.NewDecoder(req.Body).Decode(&places)

	buffer := new(bytes.Buffer)
	buffer.WriteString("{\"restaurantes\":[")

	for x := 0; x < len(places.Results); x++ {
		buffer.WriteString("{\"nombre\":\"")
		buffer.WriteString(places.Results[x].Name)
		buffer.WriteString("\", ")
		buffer.WriteString("\"lat\":")
		buffer.WriteString(strconv.FormatFloat(places.Results[x].Geometry.Location.Lat, 'f', 5, 64))
		buffer.WriteString(", ")
		buffer.WriteString("\"lon\":")
		buffer.WriteString(strconv.FormatFloat(places.Results[x].Geometry.Location.Lng, 'f', 5, 64))
	}

	buffer.WriteString("]}")
	fmt.Fprintf(w, buffer.String())
}

func redux(w http.ResponseWriter, req *http.Request) {

	var img Image
	_ = json.NewDecoder(req.Body).Decode(&img)

	bitmap, err := openImage(img.Nombre)
	if err != nil {
		fmt.Println(err)
	}

	bounds := bitmap.Bounds()
	width, height := bounds.Max.X/img.Tamaño.Ancho, bounds.Max.Y/img.Tamaño.Alto

	imgSet := image.NewRGBA(image.Rect(0, 0, img.Tamaño.Ancho, img.Tamaño.Alto))

	for y := 0; y < img.Tamaño.Alto; y++ {
		for x := 0; x < img.Tamaño.Ancho; x++ {
			pixel := bitmap.At(x*width, y*height)
			imgSet.Set(x, y, pixel)
		}
	}

	outfile, err := os.Create("lena_Redux.bmp")
	if err != nil {
		fmt.Println(err)
	}

	defer outfile.Close()

	pretty.Println(imgSet)
	json.NewEncoder(w).Encode(imgSet)
	bmp.Encode(outfile, imgSet)
}

func grayScaling(w http.ResponseWriter, req *http.Request) {

	var img Image
	_ = json.NewDecoder(req.Body).Decode(&img)

	bitmap, err := openImage(img.Nombre)
	if err != nil {
		fmt.Println(err)
	}

	bounds := bitmap.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	imgSet := image.NewRGBA(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			oldPixel := bitmap.At(x, y)
			r, g, b, _ := oldPixel.RGBA()

			avg := (r + g + b) / 3
			pixel := color.Gray{uint8(avg / 256)}

			imgSet.Set(x, y, pixel)
		}
	}

	outfile, err := os.Create("lena_GrayScale.bmp")
	if err != nil {
		fmt.Println(err)
	}

	defer outfile.Close()

	pretty.Println(imgSet)
	pretty.Println(width, height)
	bmp.Encode(outfile, imgSet)
}

func openImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return bmp.Decode(f)
}
