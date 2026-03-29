package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	luaeval "github.com/aldehir/research/internal/lua"
)

func handleEvalLua(eval *luaeval.Evaluator, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Code string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body", logger)
			return
		}
		if req.Code == "" {
			writeError(w, http.StatusBadRequest, "code is required", logger)
			return
		}

		logger.Info("evaluating lua code", "code_length", len(req.Code))

		output, err := eval.Eval(req.Code)

		resp := struct {
			Output string `json:"output"`
			Error  string `json:"error"`
		}{Output: output}

		if err != nil {
			resp.Error = err.Error()
			logger.Warn("lua evaluation error", "error", err)
		} else {
			logger.Info("lua evaluation complete", "output_length", len(output))
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
