package httpserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/unipe/linkedin/backend/server/internal/apperrors"
	"github.com/unipe/linkedin/backend/server/internal/platform/llm"
	"github.com/unipe/linkedin/backend/server/internal/platform/realm"
	seedsvc "github.com/unipe/linkedin/backend/server/internal/seed/service"
)

type internalHandler struct {
	secret string
	seed   *seedsvc.Service
	llm    *llm.Runner
}

func newInternalHandler(secret string, seed *seedsvc.Service, llmRunner *llm.Runner) *internalHandler {
	return &internalHandler{secret: secret, seed: seed, llm: llmRunner}
}

func (h *internalHandler) seedDemo(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" {
		token := strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		if token != h.secret {
			apperrors.WriteError(w, apperrors.Forbidden(apperrors.CodeInternalUnauthorized, apperrors.MsgInternalUnauthorized))
			return
		}
	}
	out, err := h.seed.SeedDemo(r.Context())
	if err != nil {
		apperrors.WriteError(w, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err))
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *internalHandler) liveLLMRun(w http.ResponseWriter, r *http.Request) {
	if h.secret != "" {
		token := strings.TrimSpace(r.Header.Get("X-Internal-Token"))
		if token != h.secret {
			apperrors.WriteError(w, apperrors.Forbidden(apperrors.CodeInternalUnauthorized, apperrors.MsgInternalUnauthorized))
			return
		}
	}
	if realm.FromHeader(r.Header.Get(realm.Header)) != realm.Live {
		apperrors.WriteError(w, apperrors.Invalid(apperrors.CodeInternal, "live realm required (X-App-Realm: live)"))
		return
	}
	if h.llm == nil {
		apperrors.WriteError(w, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, fmt.Errorf("llm runner not configured")))
		return
	}
	out, err := h.llm.RunE2EScenario(r.Context())
	if err != nil {
		apperrors.WriteError(w, apperrors.InternalCause(apperrors.CodeInternal, apperrors.MsgInternal, err))
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
