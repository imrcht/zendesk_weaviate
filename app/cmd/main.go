package main

import (
	"context"
	"fmt"
	"log"
	"zendesk_weaviate/app/config"
	"zendesk_weaviate/app/handler"
	app_models "zendesk_weaviate/app/models"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

var (
	confPathCustom = "../zendesk_weaviate/etc/"
)

func GetWeaviateClient() *weaviate.Client {
	cfg := weaviate.Config{
		Host:   "10.10.100.186:18080",
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
	app_models.DropClass("Zendesk_solved_tickets", ctx, client)

	// Retrieve schema to confirm creation
	schema, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		log.Fatalf("Error retrieving schema: %v", err)
	}

	log.Printf("Connection successful! Available classes: %+v\n", schema.Classes)

	for _, class := range schema.Classes {
		log.Printf("Class: %s\n", class.Class)

		if class.Class == "UserConversation" || class.Class == "Agentic_chatbot_conversation_history" || class.Class == "Ludo" || class.Class == "TargetFirst" || class.Class == "Aaa" || class.Class == "MultiTenancyCollection" {
			continue
		}

		// GraphQL query to count vectors in the class
		query := client.GraphQL().Aggregate().WithClassName(class.Class).WithFields(graphql.Field{Name: "meta { count }"})

		// Execute the query
		result, err := query.Do(context.Background())
		if err != nil {
			log.Fatalf("GraphQL query failed: %v", err)
		}

		// Extract and print the count
		count := result.Data["Aggregate"].(map[string]interface{})[class.Class].([]interface{})[0].(map[string]interface{})["meta"].(map[string]interface{})["count"]
		fmt.Printf("Number of vectors in class %s: %v\n", class.Class, count)
	}

	// Create ticket class and import tickets on that class
	// err = handler.CreateTicketClassAndImportTickets(client, ctx)
	// if err != nil {
	// 	log.Println("[Error] while running ticket class script:", err)
	// 	fmt.Printf("error when running ticket class script %v", err)
	// }

	//  Create ticket cluster class and import ticket clusters on that class
	err = handler.CreateTicketClusterClassAndImportTickets(client, ctx)
	if err != nil {
		log.Println("[Error] while running ticket class script:", err)
		fmt.Printf("error when running ticket class script %v", err)
	}

	// Create article class and import articles on that class
	// err := handler.CreateArticleClassAndImportArticles(client, zendeskAccessToken, zendeskSubdomain, ctx)
	// if err != nil {
	// 	log.Println("[Error] while running article class script:", err)
	// 	fmt.Printf("error when running article class script %v", err)
	// }

	// * Running a sample query to check if the data is inserted or not

	// query := "The customer, Felix Anaya, is experiencing an ongoing issue with the Voxbone SMS platform and they have reached out to Voxbone for further instructions and next steps. An escalation was also submitted for this issue."
	// query := "Regarding Voxbone SMS platform, the issue is ongoing and further steps are being discussed with Voxbone."

	// query := "How to resolve the issue of persistent errors with AI and summaries not making sense?"
	// collectionName := "Zendesk_articles"

	// fmt.Println("=========> HybridArticleQueryWeaviate  <=========")

	// if err := models.HybridArticleQueryWeaviate(client, query, collectionName, ctx, "en", []string{}, []string{}, []string{}); err != nil {
	// 	log.Println("[Error] querying Weaviate:", err)
	// 	fmt.Printf("error when querying Weaviate %v", err)
	// }

	// fmt.Println("=========> QueryArticlesByVector  <=========")

	// if err := models.QueryArticlesByVector(ctx, client, collectionName, query, 5); err != nil {
	// 	log.Println("[Error] querying Weaviate:", err)
	// 	fmt.Printf("error when querying Weaviate %v", err)
	// }
}
