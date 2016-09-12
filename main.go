/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"
)

var nc *nats.Conn
var natsErr error

func eventHandler(m *nats.Msg) {
	var f Event

	err := f.Process(m.Data)
	if err != nil {
		return
	}

	if err = f.Validate(); err != nil {
		f.Error(err)
		return
	}

	err = deleteFirewall(&f)
	if err != nil {
		f.Error(err)
		return
	}

	f.Complete()
}

func deleteFirewall(ev *Event) error {
	creds := credentials.NewStaticCredentials(ev.DatacenterAccessKey, ev.DatacenterAccessToken, "")
	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(ev.DatacenterRegion),
		Credentials: creds,
	})

	req := ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(ev.SecurityGroupAWSID),
	}

	_, err := svc.DeleteSecurityGroup(&req)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	fmt.Println("listening for firewall.delete.aws")
	nc.Subscribe("firewall.delete.aws", eventHandler)

	runtime.Goexit()
}
