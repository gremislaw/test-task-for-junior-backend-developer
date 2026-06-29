package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    taskusecase "example.com/taskservice/internal/usecase/task"
    "example.com/taskservice/internal/errs"
)

type TagHandler struct {
    svc *taskusecase.TagService
}

func NewTagHandler(svc *taskusecase.TagService) *TagHandler {
    return &TagHandler{svc: svc}
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct { Name string `json:"name"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, errs.ErrInvalidJSON)
        return
    }

    tag, err := h.svc.CreateTag(r.Context(), req.Name)
    if err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, tag)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, _ := strconv.ParseInt(vars["id"], 10, 64)

    if err := h.svc.DeleteTag(r.Context(), id); err != nil {
        writeError(w, http.StatusBadRequest, err.Error())
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) AssignToTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

    var req struct { TagIDs []int64 `json:"tag_ids"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, errs.ErrInvalidJSON)
        return
    }

    if err := h.svc.AssignTagsToTask(r.Context(), taskID, req.TagIDs); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (h *TagHandler) RemoveFromTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    taskID, _ := strconv.ParseInt(vars["id"], 10, 64)

    var req struct { TagIDs []int64 `json:"tag_ids"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, errs.ErrInvalidJSON)
        return
    }

    if err := h.svc.RemoveTagsFromTask(r.Context(), taskID, req.TagIDs); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }
    w.WriteHeader(http.StatusOK)
}