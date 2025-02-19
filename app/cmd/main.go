package main

import (
	"context"
	"fmt"
	"log"
	"zendesk_weaviate/app/config"
	"zendesk_weaviate/app/models"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

var (
	confPathCustom = "../zendesk-weaviate/etc/"
)

func GetWeaviateClient() *weaviate.Client {
	cfg := weaviate.Config{
		Host:   "172.30.127.145:18080",
		Scheme: "http",
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		log.Println("Error creating client:", err)
		panic(err)
	}

	return client
}

func main() {
	configPath := confPathCustom
	config.LoadConfig(configPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zendeskAccessToken := config.GetConfig().Zendesk.AccesssToken
	zendeskSubdomain := config.GetConfig().Zendesk.Subdomain

	log.Println("zendesk access token:", zendeskAccessToken)
	log.Println("zendesk subdomain:", zendeskSubdomain)

	client := GetWeaviateClient()
	// models.GetSchemas(client)

	// Create ticket class and import tickets on that class
	// err := handler.CreateTicketClassAndImportTickets(client, ctx)
	// if err != nil {
	// 	log.Println("[Error] while running ticket class script:", err)
	// 	fmt.Printf("error when running ticket class script %v", err)
	// }

	// Create article class and import articles on that class
	// err := handler.CreateArticleClassAndImportArticles(client, zendeskAccessToken, zendeskSubdomain, ctx)
	// if err != nil {
	// 	log.Println("[Error] while running article class script:", err)
	// 	fmt.Printf("error when running article class script %v", err)
	// }

	// * Running a sample query to check if the data is inserted or not
	query := "I have issue on microsoft dynamics"
	collectionName := "Zendesk_article_test5"
	if err := models.HybridArticleQueryWeaviate(client, query, collectionName, ctx); err != nil {
		log.Println("[Error] querying Weaviate:", err)
		fmt.Printf("error when querying Weaviate %v", err)
	}

}
