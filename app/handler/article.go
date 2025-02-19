package handler

import (
	"context"
	"fmt"
	"log"
	app_models "zendesk_weaviate/app/models"
	"zendesk_weaviate/app/shared"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

func ImportArticles(client *weaviate.Client, className string, ctx context.Context, zendeskAccesstoken, zendeskSubdomain string) {
	sections, err := shared.FetchSections(zendeskAccesstoken, zendeskSubdomain)
	if err != nil {
		log.Fatalf("Error fetching sections: %v", err)
	}

	allSectionNames := []string{}
	for _, section := range sections {
		allSectionNames = append(allSectionNames, section["section_name"])
	}

	log.Println("All sections: ", allSectionNames)

	for _, section := range sections {
		fmt.Println("Found section :>>", section["section_name"])

		// if section["section_name"] != "Messaging" {
		// 	continue
		// }

		articles, err := shared.FetchArticlesFromSection(section, zendeskAccesstoken, zendeskSubdomain)
		if err != nil {
			log.Fatalf("Error fetching articles: %v", err)
		}

		if len(articles) == 0 {
			fmt.Println("No articles found in this section")
			continue
		}

		batcher := client.Batch().ObjectsBatcher()

		for i := 0; i < len(articles); i += BATCH_SIZE {
			batchArticles := articles[i:min(i+BATCH_SIZE, len(articles))]
			log.Printf("Processing articles count %d\n", len(batchArticles))

			for _, article := range batchArticles {
				// if article["id"] != "16715310545937" {
				// 	continue
				// }

				articleContent := article["content"].(string)
				articleOriginalContent := article["original_content"].(string)
				articleTokens := shared.NumTokensFromText(articleOriginalContent)

				log.Println("Article tokens: ", articleTokens)
				var chunks []string

				if articleTokens > MAX_TOKENS {
					log.Println("Splitting article into chunks")
					chunks = shared.SplitText(articleOriginalContent, MAX_TOKENS)
					log.Println("Received chunks: ", len(chunks))
				} else {
					chunks = []string{articleContent}
				}

				for chunkCounter, chunk := range chunks {
					chunkTokens := shared.NumTokensFromText(chunk)
					log.Println("Chunk tokens: ", chunkTokens)

					chunkID := fmt.Sprintf("%s#%d", article["id"].(string), chunkCounter)
					prevChunkID := ""
					postChunkID := ""
					if chunkCounter > 0 {
						prevChunkID = fmt.Sprintf("%s#%d", article["id"].(string), chunkCounter-1)
					}
					if chunkCounter < len(chunks)-1 {
						postChunkID = fmt.Sprintf("%s#%d", article["id"].(string), chunkCounter+1)
					}

					properties := map[string]interface{}{
						"title":        article["title"],
						"content":      chunk,
						"source":       article["source"],
						"locales":      article["locales"],
						"article_id":   article["id"],
						"section":      article["section"],
						"category":     article["category"],
						"chunk_id":     chunkID,
						"prevchunk_id": prevChunkID,
						"postchunk_id": postChunkID,
					}

					// log.Println("title :>> ", article["title"])
					// log.Println("source :>> ", article["source"])
					// log.Println("locales :>> ", article["locales"])
					// log.Println("article_id :>> ", article["id"])
					// log.Println("section :>> ", article["section"])
					// log.Println("category :>> ", article["category"])
					// log.Println("category in seciont :>> ", section["category_name"])
					// log.Println("chunk_id :>> ", chunkID)
					// log.Println("prevchunk_id :>> ", prevChunkID)
					// log.Println("postchunk_id :>> ", postChunkID)

					batcher.WithObjects(&models.Object{
						Class:      className,
						Properties: properties,
					})
				}
			}

			batchRes, err := batcher.Do(ctx)
			if err != nil {
				log.Fatalf("Batch request failed: %v", err)
			}

			for _, res := range batchRes {
				if res.Result.Errors != nil {
					for _, err := range res.Result.Errors.Error {
						if err != nil {
							fmt.Printf("Error details: %v\n", *err)
							log.Fatal(err.Message)
						}
					}
				}
			}
			fmt.Println("Importing Articles complete")
			// Waiting for 10 seconds before processing next section
			await(10)
		}

	}
}

func CreateArticleClassAndImportArticles(client *weaviate.Client, zendeskAccessToken, zendeskSubdomain string, ctx context.Context) error {
	className := "Zendesk_article_test5"
	app_models.DropClass(className, ctx, client)

	if err := app_models.EnsureArticleClassExists(ctx, client, className); err != nil {
		log.Println("[Error] creating class:", err)
		fmt.Printf("error when creating class %v", err)
		return err
	}

	// Wait for 10 seconds before importing tickets
	await(20)

	ImportArticles(client, className, ctx, zendeskAccessToken, zendeskSubdomain)

	return nil
}
