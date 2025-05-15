package models

import (
	"context"
	"fmt"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

// Section represents a section in the Zendesk knowledge base
type Section struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
}

// Category represents a category in the Zendesk knowledge base
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Article represents an article in the Zendesk knowledge base
type Article struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	HTMLURL string   `json:"html_url"`
	Locale  string   `json:"locale"`
	Locales []string `json:"locales"`
}

func QueryArticlesByVector(ctx context.Context, client *weaviate.Client, className string, query string, limit int) error {
	startConfigClient := time.Now()
	// Build GraphQL query
	nearTextQuery := client.GraphQL().NearTextArgBuilder().WithConcepts([]string{query})

	delayConfigClient := time.Since(startConfigClient)
	startSearch := time.Now()
	// Execute GraphQL query
	result, err := client.GraphQL().Get().
		WithClassName(className).
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
		WithNearText(nearTextQuery).
		WithLimit(limit).
		Do(ctx)
	delaySearch := time.Since(startSearch)

	if err != nil {
		return fmt.Errorf("failed to query articles: %w", err)
	}

	startParsing := time.Now()
	// Safe type assertions
	data, ok := result.Data["Get"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected response format: %+v", result.Data)
	}

	articles, ok := data[className].([]interface{})
	if !ok {
		return fmt.Errorf("no articles found for class %s", className)
	}

	fmt.Printf("query : %s\n", query)

	for _, article := range articles {
		articleData := article.(map[string]interface{})
		id := articleData["article_id"].(string)
		fmt.Printf("Article : ID: %v, Title : %s, Source : %s, Score %v\n",
			id,
			articleData["title"],
			articleData["source"],
			articleData["_additional"].(map[string]interface{})["score"])
		fmt.Println("")
		fmt.Printf("additional : %v\n", articleData["_additional"])
		fmt.Println("")

	}

	delayParsing := time.Since(startParsing)

	CountObjectsInClass(className, client)
	fmt.Printf("Number of articles found : %v \n", len(articles))
	fmt.Printf("Config client time : %v\n", delayConfigClient)
	fmt.Printf("Search time : %v\n", delaySearch)
	fmt.Printf("Parsing time : %v\n", delayParsing)

	return nil
}
