package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
	app_models "zendesk_weaviate/app/models"
	"zendesk_weaviate/app/shared"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
	"go.uber.org/zap"
)

const (
	MAX_TOKENS = 4000 // Maximum tokens allowed for embedding generation
	BATCH_SIZE = 30   // Number of tickets to process in each inner batch
)

func ImportTicketsFromJSON(jsonFilePath string, client *weaviate.Client, className string, ctx context.Context) {
	// Step 1 - Load tickets from JSON file
	file, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	var allTickets []app_models.Ticket
	if err := json.Unmarshal(file, &allTickets); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Step 2 - Import data
	fmt.Println("Importing Tickets")
	tickets := allTickets

	for i := 0; i < len(tickets); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(tickets) {
			end = len(tickets)
		}
		batchTickets := tickets[i:end]
		fmt.Printf("Processing [inner batch] %d of %d\n", i/BATCH_SIZE+1, (len(tickets)/BATCH_SIZE)+1)

		// Create a batcher object
		batcher := client.Batch().ObjectsBatcher()

		for _, ticket := range batchTickets {
			ticketID, err := strconv.Atoi(ticket.ID)
			if err != nil {
				fmt.Println("Error:", err)
			}
			// ticketID := ticket.ID
			ticketText := ticket.Metadata.Text
			ticketLang := ticket.Metadata.LangCode
			ticketTags := ticket.Metadata.Tags
			ticketType := ticket.Metadata.Type

			// Check token count and split text if necessary
			// ticketTokens := NumTokensFromText(ticketText)
			// var chunks []string
			// if ticketTokens > MAX_TOKENS {
			// 	fmt.Println("Splitting ticket into chunks")
			// 	fmt.Println("Found ticket with more than 4096 tokens")
			// 	fmt.Println("Ticket ID: ", ticketID)
			// 	fmt.Println("Ticket tokens: ", ticketTokens)
			// 	chunks = SplitText(ticketText, MAX_TOKENS)
			// } else {
			// chunks = []string{ticketText}
			// }

			chunks := []string{ticketText}

			for chunkCounter, chunk := range chunks {
				chunkID := fmt.Sprintf("%d#%d", ticketID, chunkCounter)
				prevChunkID := ""
				if chunkCounter > 0 {
					prevChunkID = fmt.Sprintf("%d#%d", ticketID, chunkCounter-1)
				}
				postChunkID := ""
				if chunkCounter < len(chunks)-1 {
					postChunkID = fmt.Sprintf("%d#%d", ticketID, chunkCounter+1)
				}

				// Map ticket fields to Weaviate properties
				properties := map[string]interface{}{
					"ticket_id":    ticketID,
					"summary":      chunk,
					"lang_code":    ticketLang,
					"tags":         ticketTags,
					"type":         ticketType,
					"prevchunk_id": prevChunkID,
					"postchunk_id": postChunkID,
					"chunk_id":     chunkID,
				}

				// Generate embedding for the chunk
				// vectorEmbedding, err := GetEmbedding(chunk)
				// if err != nil {
				// 	fmt.Printf("Failed to generate embedding for chunk %d of ticket %s: %v\n", chunkCounter, ticketID, err)
				// 	continue
				// }
				// fmt.Printf("Vector embedding (first 5 dims): %v\n", vectorEmbedding[:5])

				// Add the chunk to the batcher
				batcher.WithObjects(&models.Object{
					Class:      className,
					Properties: properties,
					// Vector:     vectorEmbedding,
				})
			}
		}

		// Execute the batch operation
		batchRes, err := batcher.Do(ctx)
		if err != nil {
			log.Fatal("Error while executing batch operation: ", err)
		}

		// Handle errors in the batch response
		for _, res := range batchRes {
			if res.Result.Errors != nil {
				for _, err := range res.Result.Errors.Error {
					if err != nil {
						fmt.Printf("Error details: %v\n", *err)
						panic(err.Message)
					}
				}
			}
		}
		fmt.Println("Importing Tickets complete")
	}
}

func ImportTicketClustersFromJSON(jsonFilePath string, client *weaviate.Client, className string, ctx context.Context) {

	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)

	// Step 1 - Load tickets from JSON file
	file, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	var allTickets app_models.TicketClusterMap
	if err := json.Unmarshal(file, &allTickets); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Step 2 - Import data
	fmt.Println("Importing Tickets")

	// Flatten all tickets from the map into a single slice
	var tickets []app_models.TicketCluster
	for _, ticketList := range allTickets {
		tickets = append(tickets, ticketList...)
	}

	for i := 0; i < len(tickets); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(tickets) {
			end = len(tickets)
		}
		batchTickets := tickets[i:end]
		fmt.Printf("Processing [inner batch] %d of %d\n", i/BATCH_SIZE+1, (len(tickets)/BATCH_SIZE)+1)

		// Create a batcher object
		batcher := client.Batch().ObjectsBatcher()

		for _, ticket := range batchTickets {
			ticketID := ticket.TicketID
			if err != nil {
				fmt.Println("Error:", err)
			}

			description := ticket.Description
			composant := ticket.Composant
			directReason := ticket.DirectReason
			directReasonOption := ticket.DirectReasonOption
			status := ticket.Status

			// Check token count and split text if necessary
			descriptionTokens := shared.NumTokensFromText(description)
			var chunks []string
			if descriptionTokens > MAX_TOKENS {
				fmt.Println("Splitting ticket description into chunks")
				fmt.Println("Found ticket description with more than 4096 tokens")
				fmt.Println("Ticket ID: ", ticketID)
				fmt.Println("Ticket tokens: ", descriptionTokens)
				chunks = shared.SplitText(description, MAX_TOKENS)
			} else {
				chunks = []string{description}
			}

			// Map ticket fields to Weaviate properties
			properties := map[string]interface{}{
				"ticket_id":            ticketID,
				"description":          chunks[0],
				"composant":            composant,
				"direct_reason":        directReason,
				"direct_reason_option": directReasonOption,
				"status":               status,
			}

			fmt.Println("properties: ", zap.Any("properties", properties))

			// Add the chunk to the batcher
			batcher.WithObjects(&models.Object{
				Class:      className,
				Properties: properties,
			})

		}

		// Execute the batch operation
		batchRes, err := batcher.Do(ctx)
		if err != nil {
			log.Fatal("Error while executing batch operation: ", err)
		}

		// Handle errors in the batch response
		for _, res := range batchRes {
			if res.Result.Errors != nil {
				for _, err := range res.Result.Errors.Error {
					if err != nil {
						fmt.Printf("Error details: %v\n", *err)
						panic(err.Message)
					}
				}
			}
		}
		fmt.Println("Importing Tickets complete")
	}
}

func await(duration time.Duration) {
	fmt.Println("Waiting for ", duration, " seconds...")
	time.Sleep(duration * time.Second)
	fmt.Println("Done waiting!")
}

func CreateTicketClassAndImportTickets(client *weaviate.Client, ctx context.Context) error {
	className := "Zendesk_ticket_test1"
	app_models.DropClass(className, ctx, client)

	if err := app_models.EnsureTicketClassExists(ctx, client, className); err != nil {
		log.Println("[Error] creating class:", err)
		fmt.Printf("error when creating class %v", err)
		return err
	}

	// Wait for 20 seconds before importing tickets
	await(60)

	jsonFilePath := "../zendesk-weaviate/etc/tickets.json"
	ImportTicketsFromJSON(jsonFilePath, client, className, ctx)

	// * Running a sample query to check if the data is inserted or not
	query := "APEF Informatique raises an issue where outgoing calls"
	collectionName := className
	if err := app_models.HybridTicketQueryWeaviate(client, query, collectionName, ctx); err != nil {
		log.Println("[Error] querying Weaviate:", err)
		fmt.Printf("error when querying Weaviate %v", err)
		return err
	}

	return nil
}

func CreateTicketClusterClassAndImportTickets(client *weaviate.Client, ctx context.Context) error {
	// Original class name
	// className := "Zendesk_ticket_cluster_final"

	// Fake class name for testing
	className := "Zendesk_ticket_cluster_temp"
	// app_models.DropClass(className, ctx, client)

	// if err := app_models.EnsureTicketClusterClassExists(ctx, client, className); err != nil {
	// 	log.Println("[Error] creating class:", err)
	// 	fmt.Printf("error when creating class %v", err)
	// 	return err
	// }

	// // Wait for 20 seconds before importing tickets
	// await(60)

	jsonFilePath := "ticket_clusters.json"
	ImportTicketClustersFromJSON(jsonFilePath, client, className, ctx)

	// * Running a sample query to check if the data is inserted or not
	query := "i need to downgrade our license count from 17 to 12"
	collectionName := className
	if err := app_models.HybridTicketClusterQueryWeaviate(client, query, collectionName, ctx); err != nil {
		log.Println("[Error] querying Weaviate:", err)
		fmt.Printf("error when querying Weaviate %v", err)
		return err
	}

	return nil
}
