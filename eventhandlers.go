package main

import (
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2" // make sure to use v2 cloudevents here
	keptn "github.com/keptn/go-utils/pkg/lib"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

/**
* Here are all the handler functions for the individual event
* See https://github.com/keptn/spec/blob/0.8.0-alpha/cloudevents.md for details on the payload
**/

// GenericLogKeptnCloudEventHandler is a generic handler for Keptn Cloud Events that logs the CloudEvent
func GenericLogKeptnCloudEventHandler(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data interface{}) error {
	log.Printf("Handling %s Event: %s", incomingEvent.Type(), incomingEvent.Context.GetID())
	log.Printf("CloudEvent %T: %v", data, data)

	return nil
}

// HandleMigrateTriggeredEvent handles migration for django projects within the same namespace
func HandleMigrateTriggeredEvent(myKeptn *keptnv2.Keptn, incomingEvent cloudevents.Event, data *MigrateTriggeredEventData) error {
	log.Printf("Handling migrate.triggered Event: %s", incomingEvent.Context.GetID())

	// Send out a migrate.started CloudEvent
	// The get-sli.started cloud-event is new since Keptn 0.8.0 and is required to be send when the task is started
	_, err := myKeptn.SendTaskStartedEvent(&keptnv2.EventData{}, ServiceName)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to send task started CloudEvent (%s), aborting...", err.Error())
		log.Println(errMsg)
		return err
	}

	namespace := fmt.Sprintf("%s-%s", data.Project, data.Stage)

	str, err := keptn.ExecuteCommand("kubectl", []string{
		"-n", namespace,
		"exec", "deployment/" + data.Service, "--",
		"python", "manage.py", "migrate", "--noinput"})

	log.Print(str)

	if err != nil {
		// report error
		log.Print(err)
		// send out a migrate.finished failed CloudEvent
		_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
			Status:  keptnv2.StatusErrored,
			Result:  keptnv2.ResultFailed,
			Message: err.Error(),
		}, ServiceName)

		return err
	}

	// Done

	// Finally: send out a migrate.finished CloudEvent
	_, err = myKeptn.SendTaskFinishedEvent(&keptnv2.EventData{
		Result:  keptnv2.ResultPass,
		Status:  keptnv2.StatusSucceeded,
		Message: str,
	}, ServiceName)

	if err != nil {
		errMsg := fmt.Sprintf("Failed to send task finished CloudEvent (%s), aborting...", err.Error())
		log.Println(errMsg)
		return err
	}

	return nil
}
