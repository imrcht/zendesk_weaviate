package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"zendesk_weaviate/app/config"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

func GetSchemas(client *weaviate.Client) {

	schema, err := client.Schema().Getter().Do(context.Background())
	if err != nil {
		log.Fatalf("Failed to retrieve schema: %v", err)
	}

	// Print schema in a readable format
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}
	fmt.Println("Schema after creation:", string(schemaJSON))
}

func HybridTicketQueryWeaviate(client *weaviate.Client, query, collectionName string, ctx context.Context) error {
	startConfigClient := time.Now()
	graphqlQuery := client.GraphQL().HybridArgumentBuilder().
		WithQuery(query).
		WithAlpha(0.5).
		WithProperties([]string{"summary"})

	log.Println("Hybrid query:", graphqlQuery)

	delayConfigClient := time.Since(startConfigClient)
	startSearch := time.Now()
	result, err := client.GraphQL().Get().
		WithClassName(collectionName).
		WithFields(
			graphql.Field{Name: "ticket_id"},
			graphql.Field{Name: "tags"},
			graphql.Field{Name: "summary"},
			graphql.Field{
				Name: "_additional",
				Fields: []graphql.Field{
					{Name: "score"},
				},
			},
		).
		WithHybrid(graphqlQuery).
		WithLimit(5).
		Do(ctx)
	delaySearch := time.Since(startSearch)

	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	startParsing := time.Now()
	tickets, _ := result.Data["Get"].(map[string]interface{})[collectionName].([]interface{})
	fmt.Printf("query : %s\n", query)

	for _, ticket := range tickets {
		ticketData := ticket.(map[string]interface{})
		id := ticketData["ticket_id"].(float64)
		fmt.Printf("Ticket : ID: %v, Summary : %s, Tags : %s, Score %v\n",
			int(id),
			ticketData["summary"],
			ticketData["tags"],
			ticketData["_additional"].(map[string]interface{})["score"])
	}

	delayParsing := time.Since(startParsing)

	CountObjectsInClass(collectionName, client)
	fmt.Printf("Number of tickets found : %v \n", len(tickets))
	fmt.Printf("Config client time : %v\n", delayConfigClient)
	fmt.Printf("Search time : %v\n", delaySearch)
	fmt.Printf("Parsing time : %v\n", delayParsing)

	return nil
}

func HybridArticleQueryWeaviate(client *weaviate.Client, query, collectionName string, ctx context.Context) error {
	startConfigClient := time.Now()
	graphqlQuery := client.GraphQL().HybridArgumentBuilder().
		WithQuery(query).
		WithAlpha(0.5).
		WithProperties([]string{"content"})

	log.Println("Hybrid query:", graphqlQuery)

	delayConfigClient := time.Since(startConfigClient)
	startSearch := time.Now()
	result, err := client.GraphQL().Get().
		WithClassName(collectionName).
		WithFields(
			graphql.Field{Name: "article_id"},
			graphql.Field{Name: "title"},
			graphql.Field{Name: "source"},
			graphql.Field{
				Name: "_additional",
				Fields: []graphql.Field{
					{Name: "score"},
				},
			},
		).
		WithHybrid(graphqlQuery).
		WithLimit(5).
		Do(ctx)
	delaySearch := time.Since(startSearch)

	if err != nil {
		return fmt.Errorf("query execution failed: %w", err)
	}

	startParsing := time.Now()
	articles, _ := result.Data["Get"].(map[string]interface{})[collectionName].([]interface{})
	fmt.Printf("query : %s\n", query)

	for _, article := range articles {
		articleData := article.(map[string]interface{})
		id := articleData["article_id"].(string)
		fmt.Printf("Article : ID: %v, Title : %s, Source : %s, Score %v\n",
			id,
			articleData["title"],
			articleData["source"],
			articleData["_additional"].(map[string]interface{})["score"])
	}

	delayParsing := time.Since(startParsing)

	CountObjectsInClass(collectionName, client)
	fmt.Printf("Number of articles found : %v \n", len(articles))
	fmt.Printf("Config client time : %v\n", delayConfigClient)
	fmt.Printf("Search time : %v\n", delaySearch)
	fmt.Printf("Parsing time : %v\n", delayParsing)

	return nil
}

func CountObjectsInClass(className string, client *weaviate.Client) {
	// Initialize GraphQL client

	result, err := client.GraphQL().Aggregate().
		WithClassName(className).
		WithFields(graphql.Field{
			Name: "meta",
			Fields: []graphql.Field{
				{Name: "count"},
			},
		}).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	count := result.Data["Aggregate"].(map[string]interface{})[className].([]interface{})[0].(map[string]interface{})["meta"].(map[string]interface{})["count"].(float64)
	fmt.Printf("Number of objects in %s: %s\n", className, strings.ReplaceAll(strconv.FormatInt(int64(count), 10), ",", " "))

	meta, err := client.Misc().MetaGetter().Do(context.Background())
	if err != nil {
		log.Fatalf("Error retrieving metadata: %v", err)
	}

	fmt.Printf("Weaviate cluster version: %s\n", meta.Version)
}

func EnsureTicketClassExists(ctx context.Context, client *weaviate.Client, className string) error {
	// Verify if the class already exists
	existingClasses, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schemas : %w", err)
	}

	for _, class := range existingClasses.Classes {
		if class.Class == className {
			log.Printf("Found class %s", className)
			return nil
		}
	}

	log.Printf("The class '%s' does not exist. Creating...", className)

	indexInverted := true

	log.Println("Creating class with host and adding baseUrl : ", config.GetConfig().Embeddings.Host)
	// Define the new schema
	newClass := &models.Class{
		Class:       className,
		Description: "A collection of tickets",
		VectorConfig: map[string]models.VectorConfig{
			"summary_vector": {
				VectorIndexType: "hnsw",
				Vectorizer: map[string]interface{}{
					"text2vec-openai": map[string]interface{}{
						"sourceProperties": []string{"summary"},
						"model":            config.GetConfig().Embeddings.Model,
						"base_url":         config.GetConfig().Embeddings.Host,
						"baseUrl":          config.GetConfig().Embeddings.Host,
						"dimensions":       3584,
					},
				},
			},
		},
		Properties: []*models.Property{
			{
				Name:          "ticket_id",
				Description:   "ID of the ticket",
				DataType:      []string{"int"},
				IndexInverted: &indexInverted,
			},
			{
				Name:          "summary",
				Description:   "Summary of the ticket",
				DataType:      []string{"text"},
				IndexInverted: &indexInverted,
			},
			{
				Name:          "tags",
				Description:   "Tags of the ticket",
				DataType:      []string{"string[]"},
				IndexInverted: &indexInverted,
			},
			{
				Name:          "lang_code",
				Description:   "Language of the ticket comments",
				DataType:      []string{"string"},
				IndexInverted: &indexInverted,
			},
			{
				Name:        "type",
				Description: "Type of ticket",
				DataType:    []string{"string"},
			},
			{
				Name:        "prevchunk_id",
				Description: "Ticket previous chunk id",
				DataType:    []string{"string"},
			},
			{
				Name:        "postchunk_id",
				Description: "Ticket next chunk id",
				DataType:    []string{"string"},
			},
			{
				Name:        "chunk_id",
				Description: "Ticket current chunk id",
				DataType:    []string{"string"},
			},
		},
	}

	// Create the new schema
	err = client.Schema().ClassCreator().WithClass(newClass).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create class '%s' : %w", className, err)
	}

	// Retrieve schema to confirm creation
	schema, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		log.Fatalf("Error retrieving schema: %v", err)
	}

	log.Printf("Connection successful! Available classes: %+v\n", schema.Classes)

	log.Printf("The class '%s' was created successfully.", className)
	return nil
}

func EnsureArticleClassExists(ctx context.Context, client *weaviate.Client, className string) error {
	// Verify if the class already exists
	existingClasses, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schemas : %w", err)
	}

	for _, class := range existingClasses.Classes {
		if class.Class == className {
			log.Printf("Found class %s", className)
			return nil
		}
	}

	log.Printf("The class '%s' does not exist. Creating...", className)

	indexInverted := true

	// Define the new schema
	newClass := &models.Class{
		Class:       className,
		Description: "A collection of articles",
		VectorConfig: map[string]models.VectorConfig{
			"content_vector": {
				VectorIndexType: "hnsw",
				Vectorizer: map[string]interface{}{
					"text2vec-openai": map[string]interface{}{
						"sourceProperties": []string{"content"},
						"model":            config.GetConfig().Embeddings.Model,
						"base_url":         config.GetConfig().Embeddings.Host,
						"baseUrl":          config.GetConfig().Embeddings.Host,
						"dimensions":       3584,
					},
				},
			},
		},
		Properties: []*models.Property{
			{
				Name:          "article_id",
				Description:   "ID of the article",
				DataType:      []string{"string"},
				IndexInverted: &indexInverted,
			},
			{
				Name:          "title",
				Description:   "Title of the article",
				DataType:      []string{"string"},
				IndexInverted: &indexInverted,
			},
			{
				Name:          "content",
				Description:   "Content of the article",
				DataType:      []string{"text"},
				IndexInverted: &indexInverted,
			},
			{
				Name:          "source",
				Description:   "URL to the article",
				DataType:      []string{"string"},
				IndexInverted: &indexInverted,
			},
			{
				Name:        "locales",
				Description: "locales of the article",
				DataType:    []string{"string[]"},
			},
			{
				Name:        "section",
				Description: "Article section",
				DataType:    []string{"string"},
			},
			{
				Name:        "category",
				Description: "Article category",
				DataType:    []string{"string"},
			},
			{
				Name:        "prevchunk_id",
				Description: "Article previous chunk id",
				DataType:    []string{"string"},
			},
			{
				Name:        "postchunk_id",
				Description: "Article next chunk id",
				DataType:    []string{"string"},
			},
			{
				Name:        "chunk_id",
				Description: "Article current chunk id",
				DataType:    []string{"string"},
			},
		},
	}

	// Create the new schema
	err = client.Schema().ClassCreator().WithClass(newClass).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to create class '%s' : %w", className, err)
	}

	// Retrieve schema to confirm creation
	schema, err := client.Schema().Getter().Do(ctx)
	if err != nil {
		log.Fatalf("Error retrieving schema: %v", err)
	}

	log.Printf("Connection successful! Available classes: %+v\n", schema.Classes)

	log.Printf("The class '%s' was created successfully.", className)
	return nil
}

func DropClass(className string, ctx context.Context, client *weaviate.Client) error {
	// Configuration du client Weaviate

	err := client.Schema().ClassDeleter().WithClassName(className).Do(ctx)
	if err != nil {
		log.Println("[Error] while deleting class:", err)
		return fmt.Errorf("error while deleting class '%s': %v", className, err)
	}

	fmt.Printf("Class '%s' successfully deleted.\n", className)
	return nil
}
