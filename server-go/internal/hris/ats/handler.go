package ats

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func companyIDFrom(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString("companyId"))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func userIDFrom(c *gin.Context) uuid.UUID {
	id, _ := uuid.Parse(c.GetString("userId"))
	return id
}

func (h *Handler) fail(c *gin.Context, err error, action string) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	case errors.Is(err, ErrBadInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
	case errors.Is(err, ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	default:
		slog.Error("ats "+action, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Request failed"})
	}
}

// ---------- Pipelines ----------

// ListPipelines godoc
// @Summary List ATS pipelines
// @Tags ATS
// @Produce json
// @Success 200 {array} AtsPipeline
// @Router /hris/ats/pipelines [get]
func (h *Handler) ListPipelines(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	// Guarantee a default pipeline exists so the board is never empty.
	if _, err := h.svc.EnsureDefaultPipeline(c.Request.Context(), companyID); err != nil {
		h.fail(c, err, "ensure default pipeline")
		return
	}
	res, err := h.svc.ListPipelines(c.Request.Context(), companyID)
	if err != nil {
		h.fail(c, err, "list pipelines")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) CreatePipeline(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	var in PipelineInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.CreatePipeline(c.Request.Context(), companyID, in)
	if err != nil {
		h.fail(c, err, "create pipeline")
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) UpdatePipeline(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}
	var in PipelineInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.UpdatePipeline(c.Request.Context(), companyID, pipelineID, in)
	if err != nil {
		h.fail(c, err, "update pipeline")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DeletePipeline(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}
	if err := h.svc.DeletePipeline(c.Request.Context(), companyID, pipelineID); err != nil {
		h.fail(c, err, "delete pipeline")
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- Candidates ----------

func (h *Handler) ListCandidates(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	res, err := h.svc.ListCandidates(c.Request.Context(), companyID, c.Query("pipelineId"))
	if err != nil {
		h.fail(c, err, "list candidates")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) AddCandidate(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	var in AddCandidateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.AddCandidate(c.Request.Context(), companyID, in)
	if err != nil {
		h.fail(c, err, "add candidate")
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) GetCandidate(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	res, err := h.svc.GetCandidate(c.Request.Context(), companyID, candidateID)
	if err != nil {
		h.fail(c, err, "get candidate")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) MoveCandidate(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	var body struct {
		StageID string `json:"stageId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	stageID, err := uuid.Parse(body.StageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage ID"})
		return
	}
	res, err := h.svc.MoveCandidate(c.Request.Context(), companyID, candidateID, stageID)
	if err != nil {
		h.fail(c, err, "move candidate")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) SetCandidateStatus(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.SetCandidateStatus(c.Request.Context(), companyID, candidateID, body.Status)
	if err != nil {
		h.fail(c, err, "set candidate status")
		return
	}
	c.JSON(http.StatusOK, res)
}

// ---------- Scorecards ----------

func (h *Handler) ListScorecards(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	res, err := h.svc.ListScorecards(c.Request.Context(), companyID, candidateID)
	if err != nil {
		h.fail(c, err, "list scorecards")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) AddScorecard(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	var in ScorecardInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.AddScorecard(c.Request.Context(), companyID, candidateID, userIDFrom(c), "", in)
	if err != nil {
		h.fail(c, err, "add scorecard")
		return
	}
	c.JSON(http.StatusCreated, res)
}

// ---------- Interviews ----------

func (h *Handler) ListInterviews(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	res, err := h.svc.ListInterviews(c.Request.Context(), companyID, candidateID)
	if err != nil {
		h.fail(c, err, "list interviews")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) ScheduleInterview(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	var in InterviewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.ScheduleInterview(c.Request.Context(), companyID, candidateID, userIDFrom(c), in)
	if err != nil {
		h.fail(c, err, "schedule interview")
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) UpdateInterviewStatus(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	interviewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.UpdateInterviewStatus(c.Request.Context(), companyID, interviewID, body.Status)
	if err != nil {
		h.fail(c, err, "update interview status")
		return
	}
	c.JSON(http.StatusOK, res)
}

// ---------- Offer templates ----------

func (h *Handler) ListOfferTemplates(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	res, err := h.svc.ListOfferTemplates(c.Request.Context(), companyID)
	if err != nil {
		h.fail(c, err, "list offer templates")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateOfferTemplate(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	var in OfferTemplateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.CreateOfferTemplate(c.Request.Context(), companyID, in)
	if err != nil {
		h.fail(c, err, "create offer template")
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) UpdateOfferTemplate(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	var in OfferTemplateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.UpdateOfferTemplate(c.Request.Context(), companyID, templateID, in)
	if err != nil {
		h.fail(c, err, "update offer template")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DeleteOfferTemplate(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}
	if err := h.svc.DeleteOfferTemplate(c.Request.Context(), companyID, templateID); err != nil {
		h.fail(c, err, "delete offer template")
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- Offers ----------

func (h *Handler) ListOffers(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	res, err := h.svc.ListOffers(c.Request.Context(), companyID, candidateID)
	if err != nil {
		h.fail(c, err, "list offers")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateOffer(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	candidateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}
	var in OfferInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.CreateOffer(c.Request.Context(), companyID, candidateID, userIDFrom(c), in)
	if err != nil {
		h.fail(c, err, "create offer")
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) SendOffer(c *gin.Context) {
	companyID, ok := companyIDFrom(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Company access required"})
		return
	}
	offerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}
	res, err := h.svc.SendOffer(c.Request.Context(), companyID, offerID)
	if err != nil {
		h.fail(c, err, "send offer")
		return
	}
	c.JSON(http.StatusOK, res)
}

// ---------- Public (candidate token) ----------

// GetPublicOffer godoc
// @Summary Public offer view (candidate token)
// @Tags ATS
// @Produce json
// @Param token path string true "Accept token"
// @Success 200 {object} PublicOffer
// @Router /ats/offers/{token} [get]
func (h *Handler) GetPublicOffer(c *gin.Context) {
	res, err := h.svc.GetPublicOffer(c.Request.Context(), c.Param("token"))
	if err != nil {
		h.fail(c, err, "get public offer")
		return
	}
	c.JSON(http.StatusOK, res)
}

// AcceptOffer godoc
// @Summary Accept an offer (public, candidate token)
// @Tags ATS
// @Accept json
// @Produce json
// @Param token path string true "Accept token"
// @Success 200 {object} AcceptOfferResult
// @Router /ats/offers/{token}/accept [post]
func (h *Handler) AcceptOffer(c *gin.Context) {
	var body struct {
		SignatureName string `json:"signatureName"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	res, err := h.svc.AcceptOffer(c.Request.Context(), c.Param("token"), body.SignatureName)
	if err != nil {
		h.fail(c, err, "accept offer")
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) DeclineOffer(c *gin.Context) {
	if err := h.svc.DeclineOffer(c.Request.Context(), c.Param("token")); err != nil {
		h.fail(c, err, "decline offer")
		return
	}
	c.Status(http.StatusNoContent)
}
