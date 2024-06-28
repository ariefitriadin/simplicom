package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	proto "github.com/ariefitriadin/simplicom/cmd/auth-service/proto"
)

/*
*
for checking purpose
*/
func main() {
	// Set up a connection to the server.
	conn, err := grpc.NewClient("localhost:3001", grpc.WithTransportCredentials(insecure.NewCredentials())) // Adjust the port accordingly
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := proto.NewAuthServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := client.CreateClient(ctx, &proto.CreateClientRequest{
		Domain:      "localhost",
		RedirectUri: "http://localhost:8080/callback",
	})
	if err != nil {
		log.Fatalf("could not create client: %v", err)
	}
	log.Printf("CreateClient: %s", r.Message)
}
