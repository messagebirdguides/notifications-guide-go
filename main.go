package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/messagebird/go-rest-api"
	"github.com/messagebird/go-rest-api/sms"
)

var (
	// Client allows us to access the same messagebird.Client object throughout the application.
	Client *messagebird.Client
	// CurrentOrders is a placeholder database.
	CurrentOrders []Order
)

func main() {
	initDB()
	Client = messagebird.New("<enter-your-api-key>")

	// Routes
	http.HandleFunc("/", orderPage)
	http.HandleFunc("/notifyCustomer", orderNotify)

	// Define route for static resources
	staticDir := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticDir))

	// Serve
	port := ":8080"
	log.Println("Serving application on", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Println(err)
	}
}

// === Types

// Order is the structure of data we get from our orders database.
// Can be used when Unmarshalling data from json, or retrieving data from a remote database.
type Order struct {
	ID    string
	Name  string
	Phone string
	Items []string
	// Possible values for Status:
	// "pending", "delayed", "confirmed", "delivered"
	Status string
}

// === Routes and routing logic

// orderPage http.HandlerFunc
func orderPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}
		if len(r.Form) == 0 {
			log.Println("Empty form.")
		}
		s := strings.Split(r.FormValue("orderStatus"), "_")
		for i, v := range CurrentOrders {
			if v.ID == s[0] {
				CurrentOrders[i].Status = s[1]
			}
		}
	}
	RenderDefaultTemplate(w, "views/orders.gohtml", CurrentOrders)
}

func orderNotify(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	s := r.FormValue("sendMessageTo")

	for _, v := range CurrentOrders {
		if v.ID == s {
			msgToSend := isOrderConfirmed(v.Status, v.Name)
			msg, err := sms.Create(Client, "NomNom", []string{v.Phone}, msgToSend, nil)
			if err != nil {
				log.Println(err)
			} else {
				// For development only
				log.Println(msg)
			}
		}
	}
	RenderDefaultTemplate(w, "views/orders.gohtml", CurrentOrders)
}

// RenderDefaultTemplate ...
func RenderDefaultTemplate(w http.ResponseWriter, thisView string, data interface{}) {
	renderthis := []string{thisView, "views/layouts/default.gohtml"}
	t, err := template.ParseFiles(renderthis...)
	if err != nil {
		log.Fatal(err)
	}
	err = t.ExecuteTemplate(w, "default", data)
	if err != nil {
		log.Fatal(err)
	}
}

func isOrderConfirmed(orderStatus string, recipientName string) string {
	switch orderStatus {
	case "pending":
		return "Hello, " + recipientName + ", thanks for ordering at OmNomNom Foods! We're still working on your order. Please be patient with us!"
	case "confirmed":
		return "Hello, " + recipientName + ", thanks for ordering at OmNomNom Foods! We are now preparing your food with love and fresh ingredients and will keep you updated."
	case "delayed":
		return "Hello, " + recipientName + ", sometimes good things take time! Unfortunately your order is slightly delayed but will be delivered as soon as possible."
	case "delivered":
		return "Hello, " + recipientName + ", you can start setting the table! Our driver is on their way with your order! Bon appetit!"
	default:
		return "We can't find your order! Please call our customer support for assistance."
	}
}

// In a production environment, we would be getting data from
// an external source such as a REST API endpoint, or a remote database.
// Here, we're mocking up a database using a simple array of structs.
func initDB() {
	CurrentOrders = []Order{
		{"c2972b5b4eef349fb1e5cc3e3150a2b6", "Hannah Hungry", "+319876543210", []string{"1 x Hipster Burger", "Fries"}, "pending"},
		{"1b992e39dc55f0c79dbe613b3ad02f29", "Mike Madeater", "+319876543211", []string{"1 x Chef Special Mozzarella Pizza"}, "delayed"},
		{"81dc9bdb52d04dc20036dbd8313ed055", "Don Cheetos", "+319876543212", []string{"1 x Awesome Cheese Platter"}, "confirmed"},
		{"5cb59f74fd4cd18fd90ffe79b4cb1dc0", "Ace Adventures", "+319876543213", []string{"1 x Variegated Salami Combo Box"}, "delivered"},
	}
}
