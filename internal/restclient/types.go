package restclient

type Document struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Content     string  `json:"content"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	IsPublished bool    `json:"isPublished"`
	AuthorID    int64   `json:"authorId"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type QuizForm struct {
	FormID     string `json:"formId"`
	Definition any    `json:"definition"`
}

type DocumentWithForms struct {
	Document
	Forms []QuizForm `json:"forms"`
}

type DocumentAnalytics struct {
	TotalSubmissions int     `json:"totalSubmissions"`
	AverageScore     float64 `json:"averageScore"`
	HighestScore     float64 `json:"highestScore"`
	LowestScore      float64 `json:"lowestScore"`
}

type QuizSubmission struct {
	ID          int64          `json:"id"`
	UserID      int64          `json:"userId"`
	DocumentID  int64          `json:"documentId"`
	FormID      string         `json:"formId"`
	Responses   map[string]any `json:"responses"`
	Score       *int           `json:"score"`
	MaxScore    *int           `json:"maxScore"`
	SubmittedAt string         `json:"submittedAt"`
}

type SubmissionWithDocument struct {
	Submission    QuizSubmission `json:"submission"`
	DocumentTitle string         `json:"documentTitle"`
	DocumentSlug  string         `json:"documentSlug"`
}

type SubmissionWithUser struct {
	Submission QuizSubmission `json:"submission"`
	UserName   *string        `json:"userName"`
}

type SubmissionDetail struct {
	Submission     QuizSubmission `json:"submission"`
	DocumentTitle  string         `json:"documentTitle"`
	DocumentSlug   string         `json:"documentSlug"`
	FormDefinition any            `json:"formDefinition"`
}

type CreateDocumentRequest struct {
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	IsPublished bool    `json:"isPublished"`
}

type CreateDocumentResponse struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

type UpdateDocumentRequest struct {
	Title       *string `json:"title,omitempty"`
	Content     *string `json:"content,omitempty"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	IsPublished *bool   `json:"isPublished,omitempty"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

type SubmitQuizRequest struct {
	DocumentID int64          `json:"documentId"`
	FormID     string         `json:"formId"`
	Responses  map[string]any `json:"responses"`
}

type SubmitQuizResponse struct {
	ID       int64 `json:"id"`
	Score    *int  `json:"score"`
	MaxScore *int  `json:"maxScore"`
}

type SubmitQuizBatchItem struct {
	FormID    string         `json:"formId"`
	Responses map[string]any `json:"responses"`
}

type SubmitQuizBatchRequest struct {
	DocumentID  int64                 `json:"documentId"`
	Submissions []SubmitQuizBatchItem `json:"submissions"`
}

type SubmitQuizBatchResult struct {
	FormID   string `json:"formId"`
	Score    int    `json:"score"`
	MaxScore int    `json:"maxScore"`
}

type SubmitQuizBatchResponse struct {
	Results []SubmitQuizBatchResult `json:"results"`
}
