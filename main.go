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

type Coordinates struct {
	lat float64 `json:"lat"`
	lng float64 `json:"lng"`
}

type Image struct {
	Nombre string `json:"nombre"`
	Size   Size   `json:size`
}

type Size struct {
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
		buffer.WriteString("}, ")

		if x == (len(routes[0].Legs[0].Steps)) {
			buffer.WriteString("{\"lat\":")
			buffer.WriteString(strconv.FormatFloat(routes[0].Legs[0].Steps[x].EndLocation.Lat, 'f', 5, 64))
			buffer.WriteString(", ")
			buffer.WriteString("\"lon\":")
			buffer.WriteString(strconv.FormatFloat(routes[0].Legs[0].Steps[x].EndLocation.Lng, 'f', 5, 64))
			buffer.WriteString("} ")
		}
	}

	buffer.WriteString("]}")
	fmt.Fprintf(w, buffer.String())
}

func restaurantList(w http.ResponseWriter, req *http.Request) {

	var place Coordinates
	_ = json.NewDecoder(req.Body).Decode(&place)

	client, err := maps.NewClient(maps.WithAPIKey("AIzaSyBmelZAhVTODrw_gjtueTuHEs9Aka_z9nM"))
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}
	r := &maps.NearbySearchRequest{

		Location: &maps.LatLng{place.lat, place.lng},
		Radius:   100,
		Type:     maps.PlaceTypeRestaurant,
	}

	places, err := client.NearbySearch(context.Background(), r)
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}

	json.NewEncoder(w).Encode(places)
	pretty.Println(places)
}

func redux(w http.ResponseWriter, req *http.Request) {

	var img Image
	_ = json.NewDecoder(req.Body).Decode(&img)

	bitmap, err := openImage(img.Nombre)
	if err != nil {
		fmt.Println(err)
	}

	bounds := bitmap.Bounds()
	width, height := bounds.Max.X/img.Size.Ancho, bounds.Max.Y/img.Size.Alto

	imgSet := image.NewRGBA(image.Rect(0, 0, img.Size.Ancho, img.Size.Alto))

	pretty.Println(width, height)

	for y := 0; y < img.Size.Alto; y++ {
		for x := 0; x < img.Size.Ancho; x++ {
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
