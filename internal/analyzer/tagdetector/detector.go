package tagdetector

type DetectorRes struct {
	Tags []string `json:"tags"`
}

type TagDetector interface {
	DetectTags(content string) *DetectorRes
}
