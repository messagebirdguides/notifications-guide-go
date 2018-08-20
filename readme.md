# SMS Order Notifications
### ⏱ 15 min build time 

## Why build SMS order notifications? 

Have you ever ordered home delivery to find yourself wondering whether your order was received correctly and how long it'll take to arrive? Some experiences are seamless and others... not so much. 

For on-demand industries such as food delivery, ridesharing and logistics, excellent customer service during the ordering process is essential. One easy way to stand out from the crowd is providing proactive communication to keep your customers in the loop about the status of their orders. Irresepective of whether your customer is waiting for a package delivery or growing "hangry" (i.e. Hungry + Angry) awaiting their food delivery, sending timely SMS order notifications is a great strategy to create a seamless user experience.

The [MessageBird SMS Messaging API](https://developers.messagebird.com/docs/sms-messaging) provides an easy way to fully automate and integrate a notifications application into your order handling software. Busy employees can trigger the notifications application with the push of a single button - no more confused *hangry* customers and a best-in-class user experience, just like that!

## About our application

In this MessageBird Developer Guide, we'll show you how to build a runnable Order Notifications application in Go. The application is a prototype order management system deployed by our fictitious food delivery company, *Birdie NomNom Foods*.

Birdie NomNom Foods have set up the following workflow:

- New incoming orders are in a _pending_ state.
- Once the kitchen starts preparing an order, it moves to the _confirmed_ state. A message is sent to the customer to inform them about this.
- When the food is made and handed over to the delivery driver, staff marks the order _delivered_. A message is sent to the customer to let them know it will arrive momentarily.
- If preparation takes longer than expected, it can be moved to a _delayed_ state. A message is sent to the customer asking them to hang on just a little while longer. Thanks to this, Birdie NomNom Foods saves time spent answering *"Where's my order?"* calls.

**Pro-tip:** Follow this tutorial to build the whole application from scratch or, if you want to see it in action right away, you can download, clone or fork the sample application from the [MessageBird Developer Guides GitHub repository](https://github.com/messagebirdguides/notifications-guide-go).

## Getting started

We'll be building our single-page web application with:

* the latest version of [Go](https://golang.org), and
* the [MessageBird's REST API package for Go](https://github.com/messagebird/go-rest-api)

**Let's get started!**

### Structure of your application

We'll need the following components for our application to be viable:

- **A data source**: This should be a database or a REST endpoint containing information about our customers and their orders. For this guide, we'll be mocking up data already imported from an external data source.
- **A web interface to manage orders**: The web interface will display information on customer orders, and allow us to change the order status and send SMS messages to customers.
- **Route handler that contains message sending logic**: This handler would contain logic that:
    1. Checks our order against the orders database.
    2. Populates and makes a `NewMessage()` request with the appropriate information.

### Project Setup

Create a folder for your application. In this folder, create the 
following subfolders:

 - `views`
 - `views/layouts`

 Because we're dealing with data that we'd like to display as tables, we want to be able to serve CSS to make our tables readable. Create a "static" subfolder to contain our CSS and future static assets:

 - `static`

 We'll use the following packages from the Go standard library to build our routes and views:

- `net/http`: A HTTP package for building our routes and a simple http server.
- `html/template`: A HTML template library for building views.

### Create your API Key 🔑

To start making API calls, we need to generate an access key. MessageBird provides keys in _live_ and _test_ modes. For this tutorial you will need to use a live key. Otherwise, you will not be able to test the complete flow. Read more about the difference between test and live API keys [here](https://support.messagebird.com/hc/en-us/articles/360000670709-What-is-the-difference-between-a-live-key-and-a-test-key-).

Go to the [MessageBird Dashboard](https://dashboard.messagebird.com/en/user/index); if you have already created an API key it will be shown right there. Click on the eye icon to make the access key visible, then select and copy it to your clipboard. If you do not see any key on the dashboard or if you're unsure whether this key is in _live_ mode, go to the _Developers_ section and open the [API access (REST) tab](https://dashboard.messagebird.com/en/developers/access). Here, you can create new keys and manage your existing ones.

If you are having any issues creating your API key, please don't hesitate to contact support at support@messagebird.com.

**Pro-tip:** To keep our demonstration code simple, we will be saving our API key in `main.go`. However, hardcoding your credentials in the code is a risky practice that should never be used in production applications. A better method, also recommended by the [Twelve-Factor App Definition](https://12factor.net/), is to use environment variables. You can use open source packages such as [GoDotEnv](https://github.com/joho/godotenv) to read your API key from a `.env` file into your Go application. Your `.env` file should be written as follows:

`````env
MESSAGEBIRD_API_KEY=YOUR-API-KEY
`````

To use [GoDotEnv](https://github.com/joho/godotenv) in your application, install it by running:

````bash
go get -u github.com/joho/godotenv
````

Then, import it in your application:

````go
import (
  // Other imported packages
  "os"

  "github.com/joho/godotenv"
)

func main(){
  // GoDotEnv loads any ".env" file located in the same directory as main.go
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env file")
  }

  // Store the value for the key "MESSAGEBIRD_API_KEY" in the loaded '.env' file.
  apikey := os.Getenv("MESSAGEBIRD_API_KEY")

  // The rest of your application ...
}
````

## Initialize the MessageBird Client

Install the [MessageBird's REST API package for Go](https://github.com/messagebird/go-rest-api) by running:

````go
go get -u github.com/messagebird/go-rest-api
````

In your project folder, create a `main.go` file, and write the following code:

````go
package main

import (
  "github.com/messagebird/go-rest-api"
)

// Client ...
// We're initializing "Client" as a global variable so that we can access its methods in our handlers.
var Client *messagebird.Client

func main(){
  Client = messagebird.New(<enter-your-apikey>)
}
````

## Setting up your data source

To keep this guide simple, we'll be using a placeholder data source instead of an actual remote data source.

Just above `main()`, add the following code:

````go
// CurrentOrders contains a list of Orders.
// This is defined globally so that HTTP handlers can access them easily.
var CurrentOrders []Order
````

Then, add the following code after the body of `main()`:

````go
// Order structures our individual orders.
type Order struct{
    ID    string
    Name  string
    Phone string
    Items []string
    // Possible values for Status:
    // "pending", "delayed", "confirmed", "delivered"
    Status string
}

func initDB() {
    CurrentOrders = []Order{
        {"c2972b5b4eef349fb1e5cc3e3150a2b6", "Hannah Hungry", "+319876543210", []string{"1 x Hipster Burger", "Fries"}, "pending"},
        {"1b992e39dc55f0c79dbe613b3ad02f29", "Mike Madeater", "+319876543211", []string{"1 x Chef Special Mozzarella Pizza"}, "delayed"},
        {"81dc9bdb52d04dc20036dbd8313ed055", "Don Cheetos", "+319876543212", []string{"1 x Awesome Cheese Platter"}, "confirmed"},
        {"81dc9bdb52d04dc20036dbd8313ed055", "Ace Adventures", "+319876543213", []string{"1 x Variegated Salami Combo Box"}, "delivered"},
    }
}
````

Here, we've created a placeholder database in the form of a list of orders, each containing an "ID", "Name", "Phone", "Items", and "Status" field.

## Dealing with our routes

Next, we'll define our routes so that we can quickly test if the customer information we're getting from our data source can be displayed correctly.

### Define routes

We want three routes:

- First route displays and updates all current orders and their statuses. 
- Second route sends notifications to the respective customer for each order.
- Third route is for serving our CSS files, and any other static assets we may eventually need.

Modify `main.go` to look like the following:

````go
func main() {
    initDB()
    Client = messagebird.NewV2(<enter-your-api-key>)

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
````

### Writing our templates

We're going to write two templates:

- `views/layouts/default.gohtml`: This is our base template, and is useful when you need to expand your project quickly.
- `views/orders.gohtml`: This is the template that displays our orders and actions available for each order.

#### `default.gohtml`

For `default.gohtml`, write the following code:

````html
{{ define "default" }}
<!DOCTYPE html>
  <head>
    <meta charset="utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>MessageBird Verify Example</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/static/main.css" type="text/css"/>
  </head>
  <body>
    <main>
    <h1>MessageBird Notification Example</h1>
    {{ template "yield" . }}
    </main>
  </body>
</html>
{{ end }}
````

Notice that we've defined a stylesheet resource at `/static/main.css`. Let's create that stylesheet now. Make a `main.css` file in your `static` subfolder, and write in it the following CSS:

````css
/* This makes our tables a bit prettier */
table {
  border-collapse:collapse;
  width:90%;
  overflow-x:auto;
}
thead {
  text-align:left;
}
table td, table th {
  border-bottom: 1px solid black;
  padding:5px;
}
````

#### `orders.gohtml`

`orders.gohtml` is more complicated. Let's start out with a basic template:

````html
{{ define "yield" }}
{{ . }}
{{ end }}
````

We know the shape of our data: we're looking at a list of structs of type "Order". Each order has a "ID", "Name", "Phone", "Items", and "Status" field. We're using the "ID" field to make sure that we're addressing a unique order, but we don't need to display it.

To get to each individual field in each order, we can write our template in `orders.gohtml` as follows:

````html
{{ define "yield" }}
<h2>List of current orders</h2>
<table>
    <thead>
        <th>Name</th>
        <th>Phone</th>
        <th>Items</th>
        <th>Status</th>
        <th>Action</th>
    </thead>
    <tbody>
        {{ range .}}
        <tr>
            <td>{{ .Name }}</td>
            <td>{{ .Phone }}</td>
            <td>
            <ul>
            {{ range .Items }}
            <li>{{ . }}</li>
            {{ end }}
            </ul>
            <td>{{ .Status }}</td>
            <td>
            <button>Actions</button>
            </td>
        </tr>
        {{ end }}
    </tbody>
</table>
{{ end }}
````

`{{ range . }}` iterates through the list of structs in `CurrentOrders` that we'll pass to the template when calling `ExecuteTemplate()`, and produces a table with our order information filled in.

Finally, we need a way to make "POST" requests to our `/notifyCustomer` route from the `orders.gohtml` template. We need two `<form>` fields:

- One for updating the status of our orders
- Another for sending notifications to customers.

Modify `orders.gohtml` to look like the following:

````html
{{ define "yield" }}
<h2>List of current orders</h2>
<table>
  <thead>
    <th>Name</th>
    <th>Phone</th>
    <th>Items</th>
    <th>Status</th>
    <th>Action</th>
  </thead>
<tbody>
  {{ range .}}
  <tr>
    <td>{{ .Name }}</td>
    <td>{{ .Phone }}</td>
    <td>
      <ul>
      {{ range .Items }}
      <li>{{ . }}</li>
      {{ end }}
      </ul>
    <td>{{ .Status }}</td>
    <td>
      <form action="/" method="post">
      <label for="orderStatus">Select order status:</label>
      <select name="orderStatus">
      <option value="" selected="selected">--Order status--</option>
      <option value="{{.ID}}_pending">Pending</option>
      <option value="{{.ID}}_delayed">Delayed</option>
      <option value="{{.ID}}_confirmed">Confirmed</option>
      <option value="{{.ID}}_delivered">Delivered</option>
      </select>
      <button type="submit">Update</button>
      </form>
      <form action="/notifyCustomer" method="post">
      <button type="submit" name="sendMessageTo" value="{{.ID}}">Notify Customer</button>
      </form>
    </td>
  </tr>
  {{ end }}
</tbody>
</table>
{{ end }}
````

### Writing our handlers

Now, we'll write our `orderPage` and `orderNotify` handlers.

To keep our code [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself), we'll write a helper function that does a few templating tasks for us. At the bottom of `main.go`, add the following code:

````go
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
````

#### `orderPage` handler

Add the following code just above `main()` in `main.go`:

````go
func orderPage(w http.ResponseWriter, r *http.Request){
    RenderDefaultTemplate(w, "views/orders.gohtml", CurrentOrders)
}
````

This renders and executes our `orders.gohtml` and `default.gohtml` templates, and passes the `CurrentOrders` object to the resulting view.

To add the ability to update our orders, we have to check for a "POST" request submitted to the this (`"/"`) route. Rewrite `orderPage()` so that it looks like this:

````go
func orderPage(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        r.ParseForm()
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
````

Here, we parse the submitted form for values if we detect that a "POST" method is sent to the `orderPage` handler. This gives us the "ID" of the order being updated, and the resulting "Status" the order should have.

We match our form's order ID to its corresponding order "ID" in the `CurrentOrders` list, and update the "Status" field for that order "ID". Once that's done, the handler continues to call `RenderDefaultTemplate()`, rendering the updated information.

#### `orderNotify` handler

In our `orderNotify` handler, we need to:

1. Get the order "ID" that we want to send a notification for.
2. Get the order "Status".
3. Get the phone number of the customer for that order "ID".
4. Send a message to that phone number.

Under our `orderPage()` handler, add the following code:

````go
func orderNotify(w http.ResponseWriter, r *http.Request){
    err := r.ParseForm()
    if err != nil {
        log.Println(err)
    }
    s := r.FormValue("sendMessageTo")

    for _, v := range CurrentOrders {
        if v.ID == s {
            msgToSend := isOrderConfirmed(v.Status, v.Name)
            _, err := Client.NewMessage("NomNom", []string{v.Phone}, msgToSend, nil)
            if err != nil {
                log.Println(err)
            }
        }
    }
    RenderDefaultTemplate(w, "views/orders.gohtml", CurrentOrders)
}
````

So, here we parse our form for the "sendMessageTo" field that contains the order "ID" we need to send a notification for. Then we iterate through our `CurrentOrders` object, and find the order that matches the "ID" sent through our form. Once we've found the relevant order, we can use its fields to construct a message to send to our customer.

Notice that we're calling a `isOrderConfirmed()` helper function to construct our `msgToSend` parameter. Add the code for `isOrderConfirmed()` just under the body of `orderNotify()`:

````go
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
````

Our `isOrderConfirmed()` helper matches the status of the order to a list of predefined messages, and returns a message string (complete with the customer's name). We assign this to the `msgToSend` variable for use when triggering the SMS notification.

Finally, we trigger an SMS notification by sending a `NewMessage` request to the MessageBird servers with the following line in `orderNotify()`:

````go
_, err := Client.NewMessage("NomNom", []string{v.Phone}, msgToSend, nil)
````

## Testing the Application

We now have a fully working application, but we won't be able to test our application because it's still using dummy data taken from our `CurrentOrders` object. Plus, if you're using a test API key, our code in `main.go` doesn't give us visible feedback for each `NewMessage()` call.

To set up a development copy of the code to test if our implementation of `NewMessage()` works, we can modify a few things:

1. Change the "Phone" fields in your `CurrentOrders` object to a test phone number that can receive messages. This phone number should also be saved in your MessageBird account as a contact.
2. Modify our `NewMessage()` call in `orderNotify()` so that we log the message object returned:

````go
msg, err := Client.NewMessage("NomNom", []string{v.Phone}, msgToSend, nil)
    if err != nil {
        log.Println(err)
    } else {
        log.Println(msg)
    }
````

Now, a successful `NewMessage()` call would log a message object that looks like the following:

````
&{ae554eccfa9047d9aa7cbe261f65d80b https://rest.messagebird.com/messages/ae554eccfa9047d9aa7cbe261f65d80b mt sms NomNom Hello, Don Cheetos,thanks for ordering at OmNomNom Foods! We are now preparing your food with love and fresh ingredients and will keep you updated.  <nil> 10 map[] plain 1  <nil> +0000 +0000 {1 1 0 0 [{6596963426 sent 2018-08-19 17:58:20 +0000 +0000}]} []}
````

Now your can begin testing your application!

Run your application in the terminal:

````bash
go run main.go
````

1. Point your browser at http://localhost:8080/ to see the table of orders.
2. For any order displayed, select a new order status and click "Update" to update the order status. The page should display the updated order status.
3. For any order displayed, click on "Send Notification" to send an SMS notification to the customer. You should see a message object logged in the terminal each time you send a message.

## Nice work!

You now have a running SMS Notifications application!

You can now use the flow, code snippets and UI examples from this tutorial as an inspiration to build your own SMS Notifications system. Don't forget to download the code from the [MessageBird Developer Guides GitHub repository](https://github.com/messagebirdguides/notifications-guide-go).

## Next steps

Want to build something similar but not quite sure how to get started? Please feel free to let us know at support@messagebird.com, we'd love to help!
