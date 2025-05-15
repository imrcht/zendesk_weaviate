package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"zendesk_weaviate/app/models"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

func HtmlToPlainText(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// Remove specified elements
	doc.Find("header, nav, footer, sidebar, script, iframe, noscript").Remove()

	// Get the text content of the body
	text := doc.Find("body").Text()

	// Trim and return
	return strings.TrimSpace(text), nil
}

// sanitizeText removes extra spaces and cleans up extracted text
func sanitizeText(text string) string {
	// Use Bluemonday UGC policy to sanitize content
	// policy := bluemonday.NewPolicy()

	// // Trim spaces and apply the policy
	// cleanedText := policy.Sanitize(text)
	// cleanedText = strings.TrimSpace(cleanedText)
	// cleanedText = strings.Join(strings.Fields(cleanedText), " ") // Normalize spaces

	// Remove extra spaces and newlines
	// cleanedText = strings.TrimSpace(cleanedText)
	// cleanedText = regexp.MustCompile(`\s+`).ReplaceAllString(cleanedText, " ")

	// Remove unnecessary special characters (except essential ones like .,?!)
	cleanedText := regexp.MustCompile(`[^a-zA-Z0-9\s:\-*.,?!]+`).ReplaceAllString(text, "")

	// Trim spaces and reduce multiple newlines to a single one
	cleanedText = regexp.MustCompile(`\n+`).ReplaceAllString(strings.TrimSpace(cleanedText), "\n")
	// cleanedText = regexp.MustCompile(`\n+`).ReplaceAllString(cleanedText, "\n")

	return cleanedText
}

// func HtmlToPlainText(htmlString string) string {
// 	doc, err := html.Parse(strings.NewReader(htmlString))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var plainText strings.Builder
// 	var f func(*html.Node)
// 	f = func(n *html.Node) {
// 		if n.Type == html.TextNode {
// 			plainText.WriteString(n.Data)
// 		}
// 		for c := n.FirstChild; c != nil; c = c.NextSibling {
// 			f(c)
// 		}
// 	}
// 	f(doc)
// 	return plainText.String()
// }

func GetArticleLocales(articleID, sourceLocale string, zendeskAccessToken, zendeskSubdomain string) []string {
	locales := []string{"en-us", "fr", "es"}
	validLocales := []string{sourceLocale}
	for _, locale := range locales {
		if locale == sourceLocale {
			continue
		}
		url := fmt.Sprintf("https://%s.zendesk.com/api/v2/help_center/%s/articles/%s", zendeskSubdomain, locale, articleID)

		// log.Println("url for article locales:", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", zendeskAccessToken))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error fetching article locales: %v", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			log.Printf("Article %s not found in locale %s", articleID, locale)
			continue
		}
		var article map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&article); err != nil {
			log.Printf("Error decoding response: %v", err)
			continue
		}
		if article["article"] != nil {
			validLocales = append(validLocales, locale)
		}
	}
	return validLocales
}

// fetchSections fetches all sections from the Zendesk knowledge base
func FetchSections(zendeskAccessToken, zendeskSubdomain string) ([]map[string]string, error) {
	allSections := []models.Section{}
	allCategories := []models.Category{}
	page := 1
	locale := "en-us"

	for {
		url := fmt.Sprintf("https://%s.zendesk.com/api/v2/help_center/%s/sections?include=categories&per_page=100&page=%d", zendeskSubdomain, locale, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", zendeskAccessToken))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error fetching sections: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("error fetching sections: status code %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("error decoding response: %v", err)
		}

		sections := result["sections"].([]interface{})
		categories := result["categories"].([]interface{})

		for _, s := range sections {
			section := s.(map[string]interface{})
			categoryID := strconv.Itoa(int(section["category_id"].(float64)))
			allSections = append(allSections, models.Section{
				ID:         int64(section["id"].(float64)),
				Name:       section["name"].(string),
				CategoryID: categoryID,
			})
		}

		for _, c := range categories {
			category := c.(map[string]interface{})
			allCategories = append(allCategories, models.Category{
				ID:   int64(category["id"].(float64)),
				Name: category["name"].(string),
			})
		}

		nextPage := result["next_page"]
		if nextPage == nil {
			break
		}
		page++
	}

	// Create a map of category IDs to names
	categoryMap := make(map[string]string)
	for _, category := range allCategories {
		stringCategoryID := strconv.Itoa(int(category.ID))
		categoryMap[stringCategoryID] = category.Name
	}

	// fmt.Println("category map: ", zap.Any("categoryMap", categoryMap))

	// Prepare the final data
	data := []map[string]string{}
	for _, section := range allSections {
		fmt.Println("section: ", zap.Any("section", section))
		stringSectionID := strconv.Itoa(int(section.ID))
		stringCategoryID := section.CategoryID
		data = append(data, map[string]string{
			"section_id":    stringSectionID,
			"section_name":  section.Name,
			"category_id":   section.CategoryID,
			"category_name": categoryMap[stringCategoryID],
		})
	}

	return data, nil
}

// fetchArticlesFromSection fetches articles from a specific section in the Zendesk knowledge base
func FetchArticlesFromSection(section map[string]string, zendeskAccessToken, zendeskSubdomain string) ([]map[string]interface{}, error) {
	sectionID := section["section_id"]
	sectionName := strings.ToLower(strings.ReplaceAll(section["section_name"], " ", "-"))
	categoryName := strings.ToLower(section["category_name"])

	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/help_center/en-us/sections/%s/articles?per_page=100", zendeskSubdomain, sectionID)

	// log.Println("url for articles:", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", zendeskAccessToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching articles: %v", err)
	}
	defer resp.Body.Close()

	log.Println("Article response status code:", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching articles with status code %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	articles := result["articles"].([]interface{})
	data := []map[string]interface{}{}

	for _, a := range articles {
		article := a.(map[string]interface{})
		articleID := strconv.Itoa(int(article["id"].(float64)))
		articleTitle := article["title"].(string)
		articleBody, ok := article["body"].(string)
		if !ok {
			log.Printf("No body for article id: %s", articleID)
			continue
		}
		articleHTMLURL := article["html_url"].(string)
		articleLocale := article["locale"].(string)

		if articleBody == "" {
			log.Printf("Empty body article id: %s", articleID)
			continue
		}

		locales := GetArticleLocales(articleID, articleLocale, zendeskAccessToken, zendeskSubdomain)
		plainText, err := HtmlToPlainText(articleBody)
		if err != nil {
			log.Printf("Error converting HTML to plain text: %v", err)
			continue
		}
		// plainText = strings.ReplaceAll(plainText, "\n\n", "\n")
		sanitizedBody := sanitizeText(plainText)
		sanitizedBody = fmt.Sprintf("Title: %s\n-------------------\n%s", articleTitle, sanitizedBody)

		data = append(data, map[string]interface{}{
			"content":          sanitizedBody,
			"original_content": plainText,
			"id":               articleID,
			"source":           articleHTMLURL,
			"locales":          locales,
			"title":            articleTitle,
			"section":          sectionName,
			"category":         categoryName,
		})
	}

	return data, nil
}
