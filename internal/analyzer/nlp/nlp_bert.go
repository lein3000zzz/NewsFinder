package nlp

import (
	"fmt"
	"os"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	ort "github.com/yalue/onnxruntime_go"
	"go.uber.org/zap"
)

type CryptoBertRes struct {
	Bearish float32 `json:"bearish"`
	Bullish float32 `json:"bullish"`
}

type AnalyzerNLPBert struct {
	logger    *zap.SugaredLogger
	tokenizer *tokenizer.Tokenizer
	session   *ort.DynamicAdvancedSession
}

func NewNLPAnalyzerBert(logger *zap.SugaredLogger) *AnalyzerNLPBert {
	//defer ort.DestroyEnvironment()

	tk, err := pretrained.FromFile(os.Getenv("MODEL_TOKENIZER_PATH"))
	if err != nil {
		logger.Fatalw("Error loading tokenizer", "error", err)
	}

	session, err := ort.NewDynamicAdvancedSession(
		os.Getenv("MODEL_ONNX_PATH"),
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"logits"},
		nil,
	)
	if err != nil {
		logger.Fatalw("Error creating session", "error", err)
	}

	return &AnalyzerNLPBert{
		logger:    logger,
		tokenizer: tk,
		session:   session,
	}
}

// Analyze TODO: refactor
func (a *AnalyzerNLPBert) Analyze(content string) (*CryptoBertRes, error) {
	a.logger.Infow("nlp analyzing content", "content", content)

	encoded, err := a.tokenizer.EncodeSingle(content)
	if err != nil {
		a.logger.Errorw("Error encoding content", "content", content, "error", err)
		return nil, err
	}

	inputIDs := convertIntToInt64(encoded.GetIds())
	mask := convertIntToInt64(encoded.GetAttentionMask())
	tokenTypeIDs := make([]int64, len(inputIDs))
	inputShape := ort.NewShape(1, int64(len(inputIDs)))
	outputShape := ort.NewShape(1, 2)

	inputTensor, err := ort.NewTensor(inputShape, inputIDs)
	if err != nil {
		a.logger.Errorw("Error creating input tensor", "error", err)
		return nil, err
	}
	defer inputTensor.Destroy()

	maskTensor, err := ort.NewTensor(inputShape, mask)
	if err != nil {
		a.logger.Errorw("Error creating mask tensor", "error", err)
		return nil, err
	}
	defer maskTensor.Destroy()

	tokenTypeTensor, err := ort.NewTensor(inputShape, tokenTypeIDs)
	if err != nil {
		a.logger.Errorw("Error creating token type tensor", "error", err)
		return nil, err
	}
	defer tokenTypeTensor.Destroy()

	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		a.logger.Errorw("Error creating output tensor", "error", err)
		return nil, err
	}
	defer outputTensor.Destroy()

	inputs := []ort.Value{inputTensor, maskTensor, tokenTypeTensor}
	outputs := []ort.Value{outputTensor}

	err = a.session.Run(inputs, outputs)
	if err != nil {
		a.logger.Errorw("Error running session", "error", err)
		return nil, fmt.Errorf("session run failed: %w", err)
	}

	logits := outputTensor.GetData()
	probs := softmax(logits)

	res := CryptoBertRes{
		Bearish: probs[0],
		Bullish: probs[1],
	}

	return &res, nil
}
