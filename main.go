package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"

	template "text/template"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

func formatPhone(number string) string {
	splitted := strings.Split(number, "")
	return strings.Join(splitted, " ")
}

var funcs = map[string]any{"formatPhone": formatPhone}
var forwardTemplate = template.Must(template.New("forward").Funcs(funcs).Parse(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Say>Connecting you to number {{formatPhone .Callee}}. Please wait.</Say>
	<Dial callerId="{{.Caller}}">{{.Callee}}</Dial>
</Response>`))

type forwardParams struct {
	Caller string
	Callee string
}

func handleForward(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		panic(err)
	}
	fmt.Printf("forward: %v\n", req.Form)
	params := forwardParams{Caller: req.FormValue("To"), Callee: req.FormValue("Digits")}
	var buf bytes.Buffer
	err = forwardTemplate.ExecuteTemplate(&buf, "forward", params)
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
	resp.Write(buf.Bytes())
}

func handleStatusCallback(resp http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		panic(err)
	}
	fmt.Printf("status callback: %v\n", req.Form)
}

func httpServer(stop chan struct{}) {
	http.HandleFunc("/forward", handleForward)
	http.HandleFunc("/statusCallback", handleStatusCallback)
	fmt.Println("Starting http listener")
	if err := http.ListenAndServe(":8888", nil); err != nil {
		panic(err)
	}
	stop <- struct{}{}
}

func main() {
	serverStop := make(chan struct{})
	go httpServer(serverStop)
	client := twilio.NewRestClient()
	params := &twilioApi.CreateCallParams{}
	if len(os.Args) != 2 {
		fmt.Println("Usage: reversecalling <number>")
		os.Exit(1)
	}
	callee := os.Args[1]
	params.SetTo(callee)
	params.SetFrom("+14803264574")
	params.SetTwiml(
		`<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Say>Welcome to the reverse calling system. You will be connected to the number that you enter. Please enter the number to dial now.</Say>
	<Gather timeout="15" input="dtmf" action="http://agency.onald.net:8888/forward"/>
</Response>`)
	params.SetStatusCallback("http://agency.onald.net:8888/statusCallback")

	resp, err := client.Api.CreateCall(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Call Status: " + *resp.Status)
		fmt.Println("Call Sid: " + *resp.Sid)
		fmt.Println("Call Direction: " + *resp.Direction)
	}
	<-serverStop
}
