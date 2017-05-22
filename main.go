package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/image/bmp"
	"image"
	"image/color"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"googlemaps.github.io/maps"
)

type Route struct {
	Origen  string `json:"origen"`
	Destino string `json:"destino"`
}

type Image struct {
	Nombre string `json:"nombre"`
	Data   string `json:"data"`
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
		Origin:      route.Origen,
		Destination: route.Destino,
	}

	routes, _, err := client.Directions(context.Background(), r)
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}

	json_route := new(bytes.Buffer)
	json_route.WriteString("{\"routes\":[")

	json.NewDecoder(req.Body).Decode(&routes)

	for x := 0; x < len(routes[0].Legs[0].Steps); x++ {
		json_route.WriteString("{\"lat\":")
		json_route.WriteString(strconv.FormatFloat(routes[0].Legs[0].Steps[x].StartLocation.Lat, 'f', 5, 64))
		json_route.WriteString(", ")
		json_route.WriteString("\"lon\":")
		json_route.WriteString(strconv.FormatFloat(routes[0].Legs[0].Steps[x].StartLocation.Lng, 'f', 5, 64))
		json_route.WriteString("}, ")

	}

	c := trimLastChar(json_route.String(), "}, ")
	c = (c + "} ]}")

	fmt.Fprintf(w, c)
}

func restaurantList(w http.ResponseWriter, req *http.Request) {

	var place Route
	_ = json.NewDecoder(req.Body).Decode(&place)

	client, err := maps.NewClient(maps.WithAPIKey("AIzaSyBmelZAhVTODrw_gjtueTuHEs9Aka_z9nM"))
	if err != nil {
		log.Fatalf("Fatal Error: %s", err)
	}

	origin_detail := &maps.GeocodingRequest{
		Address: place.Origen,
	}

	origin_response, _ := client.Geocode(context.Background(), origin_detail)

	r := &maps.NearbySearchRequest{

		Location: &origin_response[0].Geometry.Location,
		Radius:   100,
		Type:     maps.PlaceTypeRestaurant,
	}

	places, _ := client.NearbySearch(context.Background(), r)
	json.NewDecoder(req.Body).Decode(&places)

	json_restaurants := new(bytes.Buffer)
	json_restaurants.WriteString("{\"restaurantes\":[")

	for x := 0; x < len(places.Results); x++ {
		json_restaurants.WriteString("{\"nombre\":\"")
		json_restaurants.WriteString(places.Results[x].Name)
		json_restaurants.WriteString("\", ")
		json_restaurants.WriteString("\"lat\":")
		json_restaurants.WriteString(strconv.FormatFloat(places.Results[x].Geometry.Location.Lat, 'f', 5, 64))
		json_restaurants.WriteString(", ")
		json_restaurants.WriteString("\"lon\":")
		json_restaurants.WriteString(strconv.FormatFloat(places.Results[x].Geometry.Location.Lng, 'f', 5, 64))

	}

	c := trimLastChar(json_restaurants.String(), "}, ")
	c = (c + "} ]}")

	fmt.Fprintf(w, c)
}

func grayScaling(w http.ResponseWriter, req *http.Request) {

	var img Image
	_ = json.NewDecoder(req.Body).Decode(&img)

	image_bmp := base64ToImage(img.Data, img.Nombre, 1)

	bitmap, _ := openImage(image_bmp)

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

	file_name := trimLastChar(img.Nombre, ".bmp")
	file_name = (file_name + "_GrayScale.bmp")

	outfile, _ := os.Create(file_name)
	defer outfile.Close()

	bmp.Encode(outfile, imgSet)

	new_img_data := b64.StdEncoding.EncodeToString(imgSet.Pix)

	json_image := ("{\"nombre\":\"" + file_name + "\",\"data\":\"" + new_img_data + "\"}")
	fmt.Fprintf(w, json_image)
}

func redux(w http.ResponseWriter, req *http.Request) {

	var img Image
	_ = json.NewDecoder(req.Body).Decode(&img)

	image_bmp := base64ToImage(img.Data, img.Nombre, 0)

	bitmap, _ := openImage(image_bmp)

	bounds := bitmap.Bounds()
	width, height := bounds.Max.X/img.Tamaño.Ancho, bounds.Max.Y/img.Tamaño.Alto

	imgSet := image.NewRGBA(image.Rect(0, 0, img.Tamaño.Ancho, img.Tamaño.Alto))

	for y := 0; y < img.Tamaño.Alto; y++ {
		for x := 0; x < img.Tamaño.Ancho; x++ {
			pixel := bitmap.At(x*width, y*height)
			imgSet.Set(x, y, pixel)
		}
	}

	file_name := trimLastChar(img.Nombre, ".bmp")
	file_name = (file_name + "_Redux.bmp")

	outfile, _ := os.Create(file_name)
	defer outfile.Close()

	bmp.Encode(outfile, imgSet)

	new_img_data := b64.StdEncoding.EncodeToString(imgSet.Pix)

	json_image := ("{\"nombre\":\"" + file_name + "\",\"data\":\"" + new_img_data + "\"}")
	fmt.Fprintf(w, json_image)
}

func base64ToImage(data, nombre string, c int) string {

	img_data := data

	data_current, _ := b64.StdEncoding.DecodeString(img_data)

	bmp_name := trimLastChar(nombre, ".bmp")

	if c != 0 {
		bmp_name = bmp_name + "_GrayScale.bmp"
	} else {
		bmp_name = bmp_name + "_Redux.bmp"
	}

	outfile, _ := os.Create(bmp_name)
	outfile.Write(data_current)
	defer outfile.Close()

	return bmp_name
}

func openImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return bmp.Decode(f)
}

func trimLastChar(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
