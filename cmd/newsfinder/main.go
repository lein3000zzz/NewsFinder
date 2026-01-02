package main

import (
	"NewsFinder/internal/app"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/clems4ever/all-minilm-l6-v2-go/all_minilm_l6_v2"
	ort "github.com/yalue/onnxruntime_go"
)

func main() {
	app.InitEnv()

	ort.SetSharedLibraryPath(os.Getenv("ONNX_PATH"))

	err := ort.InitializeEnvironment()
	if err != nil {
		log.Fatalf("Error initializing ort environment %s, %s", "error", err)
	}

	model, err := all_minilm_l6_v2.NewModel(
		all_minilm_l6_v2.WithRuntimePath(os.Getenv("ONNX_PATH")),
	)
	if err != nil {
		panic(err)
	}
	defer model.Close()

	// Base sentence to compare against
	baseSentence := "The dog is running in the park"

	// Three candidate sentences with varying degrees of similarity
	candidates := []string{
		"A dog runs through the park",      // Very similar
		"The cat is sleeping on the couch", // Somewhat similar
		"I love eating pizza for dinner",   // Not similar
	}

	// Compute embeddings
	baseEmbedding, _ := model.Compute(baseSentence, true)
	candidateEmbeddings, _ := model.ComputeBatch(candidates, true)

	fmt.Println(baseEmbedding)
	fmt.Println(candidateEmbeddings)

	return

	nf := app.InitApp()

	nf.StartApp()

	select {
	case <-time.After(15 * time.Second):
		fmt.Println("timeout")
	}
}
