package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chriscow/minds"
	"github.com/google/generative-ai-go/genai"

	pb "cloud.google.com/go/ai/generativelanguage/apiv1beta/generativelanguagepb"
	"github.com/matryer/is"
)

// BUGBUG: The genai package expects the server to respond with some kind of Protobuf message.
// If you trace it back you will see inside of github.com/googleapis/gax-go/v2/proto_json_stream.go
// that it is expecting a JSON array of objects with opening and closing square braces,
// but this server is responding with a JSON object. This is causing the test to fail.
func newMockResponse(role string, parts ...genai.Part) *pb.GenerateContentResponse {
	v := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Role:  role,
					Parts: parts,
				},
			},
		},
	}
	return &pb.GenerateContentResponse{
		Candidates: pvTransformSlice(v.Candidates, candidateToProto),
	}
}

func candidateToProto(v *genai.Candidate) *pb.Candidate {
	if v == nil {
		return nil
	}
	return &pb.Candidate{
		Index:        pvAddrOrNil(v.Index),
		Content:      contentToProto(v.Content),
		FinishReason: pb.Candidate_FinishReason(v.FinishReason),
		// SafetyRatings:    pvTransformSlice(v.SafetyRatings, (*SafetyRating).toProto),
		// CitationMetadata: v.CitationMetadata.toProto(),
		TokenCount: v.TokenCount,
	}
}

func textToPart(p genai.Part) *pb.Part {
	return &pb.Part{
		Data: &pb.Part_Text{Text: string(p.(genai.Text))},
	}
}

func contentToProto(v *genai.Content) *pb.Content {
	if v == nil {
		return nil
	}
	return &pb.Content{
		Parts: pvTransformSlice(v.Parts, textToPart),
		Role:  v.Role,
	}
}

func pvAddrOrNil[T comparable](x T) *T {
	var z T
	if x == z {
		return nil
	}
	return &x
}

func pvTransformSlice[From, To any](from []From, f func(From) To) []To {
	if from == nil {
		return nil
	}
	to := make([]To, len(from))
	for i, e := range from {
		to[i] = f(e)
	}
	return to
}

func TestHandleMessage(t *testing.T) {
	t.Skip("Skipping test: The genai package expects the server to respond with some kind of Protobuf message.")
	t.Run("returns updated thread", func(t *testing.T) {
		is := is.New(t)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(newMockResponse("ai", genai.Text("Hello, world!")))
		}))
		defer server.Close()

		var provider minds.ContentGenerator
		ctx := context.Background()
		provider, err := NewProvider(ctx, WithBaseURL(server.URL), WithAPIKey("test"))
		is.NoErr(err) // Provider initialization should not fail

		thread := minds.NewThreadContext(context.Background()).
			WithMessages(minds.Message{
				Role: minds.RoleUser, Content: "Hi",
			})

		handler, ok := provider.(minds.ThreadHandler)
		is.True(ok) // provider should implement the ThreadHandler interface

		result, err := handler.HandleThread(thread, nil)
		is.NoErr(err) // HandleMessage should not return an error
		messages := result.Messages()
		is.Equal(len(messages), 2)
		is.Equal(messages[1].Role, minds.RoleAssistant)
		is.Equal(messages[1].Content, "Hello, world!")
	})

	t.Run("returns error on failure", func(t *testing.T) {
		is := is.New(t)
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-1*time.Second))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(newMockResponse("ai", genai.Text("Hello, world!")))
		}))
		defer server.Close()

		handler, err := NewProvider(ctx, WithBaseURL(server.URL))
		is.NoErr(err) // Provider initialization should not fail

		thread := minds.NewThreadContext(ctx).
			WithMessages(minds.Message{
				Role: minds.RoleUser, Content: "Hi",
			})

		_, err = handler.HandleThread(thread, nil)
		is.True(err != nil) // HandleMessage should return an error
		is.Equal(err.Error(), context.DeadlineExceeded.Error())
		cancel()
	})
}
