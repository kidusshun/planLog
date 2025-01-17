package whisper


type WhisperResponseBody struct {
	Text string `json:"text"`
	XGroq XGroq `json:"x_groq"`	
}

type XGroq struct {
	ID string `json:"id"`
}