package inference

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	onnxruntime_go "github.com/yalue/onnxruntime_go"
)

type PredictionResult struct {
	PredictedClass string     // "LOW", "MEDIUM", "HIGH"
	Probabilities  [3]float32 // [P(LOW), P(MEDIUM), P(HIGH)]
	HighProb       float32
}

type ModelSessions struct {
	LGB  *onnxruntime_go.AdvancedSession
	XGB  *onnxruntime_go.AdvancedSession
	RF   *onnxruntime_go.AdvancedSession
	Meta OnnxMeta

	// Mutexes to protect the shared tensors associated with each AdvancedSession
	lgbMu sync.Mutex
	xgbMu sync.Mutex
	rfMu  sync.Mutex

	// Pointers to the pre-allocated input and output tensors
	lgbInputTensor *onnxruntime_go.Tensor[float32]
	lgbLabelTensor *onnxruntime_go.Tensor[int64]
	lgbProbTensor  *onnxruntime_go.Tensor[float32]

	xgbInputTensor *onnxruntime_go.Tensor[float32]
	xgbLabelTensor *onnxruntime_go.Tensor[int64]
	xgbProbTensor  *onnxruntime_go.Tensor[float32]

	rfInputTensor  *onnxruntime_go.Tensor[float32]
	rfLabelTensor  *onnxruntime_go.Tensor[int64]
	rfProbTensor   *onnxruntime_go.Tensor[float32]
}

// LoadONNXMeta reads and parses the onnx_meta.json file
func LoadONNXMeta(metaPath string) (OnnxMeta, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return OnnxMeta{}, err
	}

	var meta OnnxMeta
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return OnnxMeta{}, err
	}

	// Populate flat police station class list for easier index mapping
	meta.PoliceStationClasses = meta.PoliceStationEncoder.Classes

	return meta, nil
}

// InitModelSessions initializes the ONNX runtime environment and loads all three models
func InitModelSessions(libPath, modelDir string) (*ModelSessions, error) {
	// 1. Initialize ONNX runtime environment
	log.Printf("Initializing ONNX Runtime using shared library at %s...\n", libPath)
	onnxruntime_go.SetSharedLibraryPath(libPath)
	err := onnxruntime_go.InitializeEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX environment: %w", err)
	}

	// 2. Load metadata
	metaPath := filepath.Join(modelDir, "onnx_meta.json")
	meta, err := LoadONNXMeta(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load onnx_meta.json: %w", err)
	}

	log.Printf("Loaded metadata sidecar. Primary model: %s, Features: %d\n", meta.PrimaryModel, meta.NFeatures)

	sessions := &ModelSessions{
		Meta: meta,
	}

	// Create reusable shape objects
	inputShape := onnxruntime_go.NewShape(1, int64(meta.NFeatures))
	labelShape := onnxruntime_go.NewShape(1)
	probShape := onnxruntime_go.NewShape(1, int64(meta.NFeatures)) // wait! The probability shape is always [1, 3] since there are 3 classes!
	// Yes, classes: LOW, MEDIUM, HIGH (3 classes)
	probShape = onnxruntime_go.NewShape(1, 3)

	options, err := onnxruntime_go.NewSessionOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to create session options: %w", err)
	}
	defer options.Destroy()

	// Load LGB model
	lgbPath := filepath.Join(modelDir, meta.Models["lgb"])
	log.Printf("Loading LightGBM model from %s...\n", lgbPath)
	sessions.lgbInputTensor, err = onnxruntime_go.NewTensor(inputShape, make([]float32, meta.NFeatures))
	if err != nil {
		return nil, err
	}
	sessions.lgbLabelTensor, err = onnxruntime_go.NewEmptyTensor[int64](labelShape)
	if err != nil {
		return nil, err
	}
	sessions.lgbProbTensor, err = onnxruntime_go.NewEmptyTensor[float32](probShape)
	if err != nil {
		return nil, err
	}
	sessions.LGB, err = onnxruntime_go.NewAdvancedSession(
		lgbPath,
		[]string{"float_input"},
		[]string{"label", "probabilities"},
		[]onnxruntime_go.Value{sessions.lgbInputTensor},
		[]onnxruntime_go.Value{sessions.lgbLabelTensor, sessions.lgbProbTensor},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LGB session: %w", err)
	}

	// Load XGB model
	xgbPath := filepath.Join(modelDir, meta.Models["xgb"])
	log.Printf("Loading XGBoost model from %s...\n", xgbPath)
	sessions.xgbInputTensor, err = onnxruntime_go.NewTensor(inputShape, make([]float32, meta.NFeatures))
	if err != nil {
		return nil, err
	}
	sessions.xgbLabelTensor, err = onnxruntime_go.NewEmptyTensor[int64](labelShape)
	if err != nil {
		return nil, err
	}
	sessions.xgbProbTensor, err = onnxruntime_go.NewEmptyTensor[float32](probShape)
	if err != nil {
		return nil, err
	}
	sessions.XGB, err = onnxruntime_go.NewAdvancedSession(
		xgbPath,
		[]string{"float_input"},
		[]string{"label", "probabilities"},
		[]onnxruntime_go.Value{sessions.xgbInputTensor},
		[]onnxruntime_go.Value{sessions.xgbLabelTensor, sessions.xgbProbTensor},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create XGB session: %w", err)
	}

	// Load RF model
	rfPath := filepath.Join(modelDir, meta.Models["rf"])
	log.Printf("Loading RandomForest model from %s...\n", rfPath)
	sessions.rfInputTensor, err = onnxruntime_go.NewTensor(inputShape, make([]float32, meta.NFeatures))
	if err != nil {
		return nil, err
	}
	sessions.rfLabelTensor, err = onnxruntime_go.NewEmptyTensor[int64](labelShape)
	if err != nil {
		return nil, err
	}
	sessions.rfProbTensor, err = onnxruntime_go.NewEmptyTensor[float32](probShape)
	if err != nil {
		return nil, err
	}
	sessions.RF, err = onnxruntime_go.NewAdvancedSession(
		rfPath,
		[]string{"float_input"},
		[]string{"label", "probabilities"},
		[]onnxruntime_go.Value{sessions.rfInputTensor},
		[]onnxruntime_go.Value{sessions.rfLabelTensor, sessions.rfProbTensor},
		options,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create RF session: %w", err)
	}

	log.Println("All three ONNX sessions successfully loaded once at startup")
	return sessions, nil
}

// Destroy cleans up the ONNX sessions and pre-allocated tensors
func (s *ModelSessions) Destroy() {
	if s.LGB != nil {
		_ = s.LGB.Destroy()
	}
	if s.XGB != nil {
		_ = s.XGB.Destroy()
	}
	if s.RF != nil {
		_ = s.RF.Destroy()
	}

	if s.lgbInputTensor != nil {
		s.lgbInputTensor.Destroy()
	}
	if s.lgbLabelTensor != nil {
		s.lgbLabelTensor.Destroy()
	}
	if s.lgbProbTensor != nil {
		s.lgbProbTensor.Destroy()
	}

	if s.xgbInputTensor != nil {
		s.xgbInputTensor.Destroy()
	}
	if s.xgbLabelTensor != nil {
		s.xgbLabelTensor.Destroy()
	}
	if s.xgbProbTensor != nil {
		s.xgbProbTensor.Destroy()
	}

	if s.rfInputTensor != nil {
		s.rfInputTensor.Destroy()
	}
	if s.rfLabelTensor != nil {
		s.rfLabelTensor.Destroy()
	}
	if s.rfProbTensor != nil {
		s.rfProbTensor.Destroy()
	}

	_ = onnxruntime_go.DestroyEnvironment()
}

// RunInference executes inference on the primary ONNX model in a thread-safe manner
func RunInference(sessions *ModelSessions, features FeatureInput) (PredictionResult, error) {
	primary := sessions.Meta.PrimaryModel

	var err error
	var labels []int64
	var probs []float32

	switch primary {
	case "xgb":
		sessions.xgbMu.Lock()
		defer sessions.xgbMu.Unlock()

		// Fill input tensor
		sliceData := features.ToSlice(sessions.Meta.FeatureCols)
		copy(sessions.xgbInputTensor.GetData(), sliceData)

		// Run
		err = sessions.XGB.Run()
		if err != nil {
			return PredictionResult{}, err
		}

		labels = sessions.xgbLabelTensor.GetData()
		probs = sessions.xgbProbTensor.GetData()

	case "rf":
		sessions.rfMu.Lock()
		defer sessions.rfMu.Unlock()

		// Fill input tensor
		sliceData := features.ToSlice(sessions.Meta.FeatureCols)
		copy(sessions.rfInputTensor.GetData(), sliceData)

		// Run
		err = sessions.RF.Run()
		if err != nil {
			return PredictionResult{}, err
		}

		labels = sessions.rfLabelTensor.GetData()
		probs = sessions.rfProbTensor.GetData()

	default: // lgb is the default primary model
		sessions.lgbMu.Lock()
		defer sessions.lgbMu.Unlock()

		// Fill input tensor
		sliceData := features.ToSlice(sessions.Meta.FeatureCols)
		copy(sessions.lgbInputTensor.GetData(), sliceData)

		// Run
		err = sessions.LGB.Run()
		if err != nil {
			return PredictionResult{}, err
		}

		labels = sessions.lgbLabelTensor.GetData()
		probs = sessions.lgbProbTensor.GetData()
	}

	if len(labels) == 0 || len(probs) < 3 {
		return PredictionResult{}, fmt.Errorf("empty or invalid outputs from model run")
	}

	predictedIdx := labels[0]
	var predictedClass string
	if predictedIdx >= 0 && predictedIdx < 3 {
		predictedClass = sessions.Meta.ClassNames[predictedIdx]
	} else {
		predictedClass = "LOW"
	}

	return PredictionResult{
		PredictedClass: predictedClass,
		Probabilities:  [3]float32{probs[0], probs[1], probs[2]},
		HighProb:       probs[2], // Probability of HIGH risk class
	}, nil
}
