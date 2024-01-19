package main

import (
	"bytes"
	"fmt"
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
var createCallTemplate = template.Must(template.New("createCall").Funcs(funcs).Parse(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
	<Say>Welcome to the reverse calling system. You will be connected to the number {{formatPhone .ConnectTo}}.</Say>
	<Dial callerId="{{.DialFirst}}">{{.ConnectTo}}</Dial>
</Response>`))

type createCallParams struct {
	DialFirst string
	ConnectTo string
}

func setupCall(client *twilio.RestClient, dialFirst string, connectTo string) (*twilioApi.ApiV2010Call, error) {
	var twiml bytes.Buffer
	err := createCallTemplate.ExecuteTemplate(&twiml, "createCall", createCallParams{DialFirst: dialFirst, ConnectTo: connectTo})
	if err != nil {
		return nil, err
	}
	params := &twilioApi.CreateCallParams{}
	params.SetTo(dialFirst)
	params.SetFrom("+14803264574")
	params.SetTwiml(twiml.String())
	params.SetMachineDetection("Enable")

	resp, err := client.Api.CreateCall(params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func main() {
	client := twilio.NewRestClient()
	if len(os.Args) != 3 {
		fmt.Println("Usage: reversecalling <dial-first-number> <connect-to-number>")
		os.Exit(1)
	}
	dialFirst := os.Args[1]
	connectTo := os.Args[2]
	call, err := setupCall(client, dialFirst, connectTo)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Call SID: ", *call.Sid)
}
