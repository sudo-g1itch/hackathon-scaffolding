package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/sudo-g1itch/hackathon-scaffolding/internal/apperr"
	"github.com/sudo-g1itch/hackathon-scaffolding/internal/config"
)

const (
	deepgramListenURL = "https://api.deepgram.com/v1/listen"
	deepgramSpeakURL  = "https://api.deepgram.com/v1/speak"

	// defaultAudioMimeType matches what MediaRecorder produces in Chrome and
	// Firefox. Safari sends mp4, which is why callers may override it.
	defaultAudioMimeType = "audio/webm"

	// TTSMimeType is the format Deepgram returns for synthesis, and therefore
	// the Content-Type the API hands back to the browser.
	TTSMimeType = "audio/mpeg"

	// maxTTSChars bounds a synthesis request; Deepgram rejects very long input.
	maxTTSChars = 1900
)

// VoiceService is the speech layer: Deepgram STT and TTS live behind it.
type VoiceService interface {
	// Available reports whether a Deepgram key is configured.
	Available() bool

	// TranscribeAudio converts recorded audio to text. mimeType may be empty,
	// in which case a browser-friendly default is assumed.
	TranscribeAudio(ctx context.Context, audio []byte, mimeType string) (string, error)

	// SynthesizeSpeech renders text as MP3 audio.
	SynthesizeSpeech(ctx context.Context, text string) ([]byte, error)
}

type voiceService struct {
	cfg    config.AI
	log    *zap.Logger
	client *http.Client
}

// NewVoiceService builds the Deepgram-backed voice service. As with the AI
// service, a missing key disables the feature rather than blocking startup.
func NewVoiceService(cfg config.AI, log *zap.Logger) VoiceService {
	return &voiceService{
		cfg:    cfg,
		log:    log,
		client: &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (s *voiceService) Available() bool { return s.cfg.VoiceEnabled() }

func (s *voiceService) TranscribeAudio(ctx context.Context, audio []byte, mimeType string) (string, error) {
	if !s.Available() {
		return "", apperr.Unavailable(
			"Voice features are not configured on this server. Set AI_DEEPGRAM_API_KEY to enable them.")
	}
	if len(audio) == 0 {
		return "", apperr.Validation(apperr.Fields{"audio": {"is empty"}})
	}

	query := url.Values{
		"model":        {s.cfg.DeepgramSTTModel},
		"smart_format": {"true"},
		"punctuate":    {"true"},
	}
	endpoint := deepgramListenURL + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(audio))
	if err != nil {
		return "", apperr.Internal(err)
	}
	req.Header.Set("Authorization", "Token "+s.cfg.DeepgramAPIKey)
	req.Header.Set("Content-Type", normalizeAudioMimeType(mimeType))

	resp, err := s.client.Do(req)
	if err != nil {
		s.log.Error("deepgram transcription request failed", zap.Error(err))
		return "", apperr.Unavailable(
			"The speech service is temporarily unavailable. Please try again.").Wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperr.Unavailable("Could not read the speech service response.").Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		s.log.Error("deepgram transcription rejected",
			zap.Int("status", resp.StatusCode), zap.String("body", truncate(string(body), 500)))

		if resp.StatusCode == http.StatusBadRequest {
			return "", apperr.Unprocessable("That audio recording could not be transcribed. Please record again.")
		}
		return "", apperr.Unavailable(
			"The speech service is temporarily unavailable. Please try again.")
	}

	var parsed struct {
		Results struct {
			Channels []struct {
				Alternatives []struct {
					Transcript string  `json:"transcript"`
					Confidence float64 `json:"confidence"`
				} `json:"alternatives"`
			} `json:"channels"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", apperr.Unavailable("The speech service returned an unreadable response.").Wrap(err)
	}

	for _, channel := range parsed.Results.Channels {
		for _, alt := range channel.Alternatives {
			if transcript := strings.TrimSpace(alt.Transcript); transcript != "" {
				return transcript, nil
			}
		}
	}

	// Deepgram answers 200 with an empty transcript for silence — a real user
	// outcome (mic muted, nothing said), not a server fault.
	return "", nil
}

func (s *voiceService) SynthesizeSpeech(ctx context.Context, text string) ([]byte, error) {
	if !s.Available() {
		return nil, apperr.Unavailable(
			"Voice features are not configured on this server. Set AI_DEEPGRAM_API_KEY to enable them.")
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, apperr.Validation(apperr.Fields{"text": {"is required"}})
	}
	if len(text) > maxTTSChars {
		text = text[:maxTTSChars]
	}

	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, apperr.Internal(err)
	}

	endpoint := deepgramSpeakURL + "?" + url.Values{"model": {s.cfg.DeepgramTTSModel}}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, apperr.Internal(err)
	}
	req.Header.Set("Authorization", "Token "+s.cfg.DeepgramAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		s.log.Error("deepgram synthesis request failed", zap.Error(err))
		return nil, apperr.Unavailable(
			"The speech service is temporarily unavailable. Please try again.").Wrap(err)
	}
	defer func() { _ = resp.Body.Close() }()

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperr.Unavailable("Could not read the speech service response.").Wrap(err)
	}

	if resp.StatusCode != http.StatusOK {
		s.log.Error("deepgram synthesis rejected",
			zap.Int("status", resp.StatusCode), zap.String("body", truncate(string(audio), 500)))
		return nil, apperr.Unavailable(
			"The speech service is temporarily unavailable. Please try again.")
	}

	return audio, nil
}

// normalizeAudioMimeType strips codec parameters Deepgram does not expect
// ("audio/webm;codecs=opus" → "audio/webm") and falls back to a default.
func normalizeAudioMimeType(mimeType string) string {
	mimeType = strings.TrimSpace(mimeType)
	if mimeType == "" {
		return defaultAudioMimeType
	}
	if idx := strings.Index(mimeType, ";"); idx >= 0 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	if !strings.HasPrefix(mimeType, "audio/") && !strings.HasPrefix(mimeType, "video/") {
		return defaultAudioMimeType
	}
	return mimeType
}
