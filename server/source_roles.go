package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bouk/httprouter"
	"github.com/influxdata/chronograf"
)

// NewSourceRole adds role to source
func (h *Service) NewSourceRole(w http.ResponseWriter, r *http.Request) {
	var req sourceRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		invalidJSON(w, h.Logger)
		return
	}

	if err := req.ValidCreate(); err != nil {
		invalidData(w, err, h.Logger)
		return
	}

	ctx := r.Context()
	srcID, ts, err := h.sourcesSeries(ctx, w, r)
	if err != nil {
		return
	}

	roles, ok := h.hasRoles(ctx, ts)
	if !ok {
		Error(w, http.StatusNotFound, fmt.Sprintf("Source %d does not have role capability", srcID), h.Logger)
		return
	}

	if _, err := roles.Get(ctx, req.Name); err == nil {
		Error(w, http.StatusBadRequest, fmt.Sprintf("Source %d already has role %s", srcID, req.Name), h.Logger)
		return
	}

	res, err := roles.Add(ctx, &req.SourceRole)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error(), h.Logger)
		return
	}

	rr := newRoleResponse(srcID, res)
	w.Header().Add("Location", rr.Links.Self)
	encodeJSON(w, http.StatusCreated, rr, h.Logger)
}

// UpdateSourceRole changes the permissions or users of a role
func (h *Service) UpdateSourceRole(w http.ResponseWriter, r *http.Request) {
	var req sourceRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		invalidJSON(w, h.Logger)
		return
	}
	if err := req.ValidUpdate(); err != nil {
		invalidData(w, err, h.Logger)
		return
	}

	ctx := r.Context()
	srcID, ts, err := h.sourcesSeries(ctx, w, r)
	if err != nil {
		return
	}

	roles, ok := h.hasRoles(ctx, ts)
	if !ok {
		Error(w, http.StatusNotFound, fmt.Sprintf("Source %d does not have role capability", srcID), h.Logger)
		return
	}

	rid := httprouter.GetParamFromContext(ctx, "rid")
	req.Name = rid

	if err := roles.Update(ctx, &req.SourceRole); err != nil {
		Error(w, http.StatusBadRequest, err.Error(), h.Logger)
		return
	}

	role, err := roles.Get(ctx, req.Name)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error(), h.Logger)
		return
	}
	rr := newRoleResponse(srcID, role)
	w.Header().Add("Location", rr.Links.Self)
	encodeJSON(w, http.StatusOK, rr, h.Logger)
}

// SourceRoleID retrieves a role with ID from store.
func (h *Service) SourceRoleID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	srcID, ts, err := h.sourcesSeries(ctx, w, r)
	if err != nil {
		return
	}

	roles, ok := h.hasRoles(ctx, ts)
	if !ok {
		Error(w, http.StatusNotFound, fmt.Sprintf("Source %d does not have role capability", srcID), h.Logger)
		return
	}

	rid := httprouter.GetParamFromContext(ctx, "rid")
	role, err := roles.Get(ctx, rid)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error(), h.Logger)
		return
	}
	rr := newRoleResponse(srcID, role)
	encodeJSON(w, http.StatusOK, rr, h.Logger)
}

// SourceRoles retrieves all roles from the store
func (h *Service) SourceRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	srcID, ts, err := h.sourcesSeries(ctx, w, r)
	if err != nil {
		return
	}

	store, ok := h.hasRoles(ctx, ts)
	if !ok {
		Error(w, http.StatusNotFound, fmt.Sprintf("Source %d does not have role capability", srcID), h.Logger)
		return
	}

	roles, err := store.All(ctx)
	if err != nil {
		Error(w, http.StatusBadRequest, err.Error(), h.Logger)
		return
	}

	rr := make([]roleResponse, len(roles))
	for i, role := range roles {
		rr[i] = newRoleResponse(srcID, &role)
	}

	res := struct {
		Roles []roleResponse `json:"roles"`
	}{rr}
	encodeJSON(w, http.StatusOK, res, h.Logger)
}

// RemoveSourceRole removes role from data source.
func (h *Service) RemoveSourceRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	srcID, ts, err := h.sourcesSeries(ctx, w, r)
	if err != nil {
		return
	}

	roles, ok := h.hasRoles(ctx, ts)
	if !ok {
		Error(w, http.StatusNotFound, fmt.Sprintf("Source %d does not have role capability", srcID), h.Logger)
		return
	}

	rid := httprouter.GetParamFromContext(ctx, "rid")
	if err := roles.Delete(ctx, &chronograf.SourceRole{Name: rid}); err != nil {
		Error(w, http.StatusBadRequest, err.Error(), h.Logger)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// sourceRoleRequest is the format used for both creating and updating roles
type sourceRoleRequest struct {
	chronograf.SourceRole
}

func (r *sourceRoleRequest) ValidCreate() error {
	if r.Name == "" || len(r.Name) > 254 {
		return fmt.Errorf("Name is required for a role")
	}
	for _, user := range r.Users {
		if user.Name == "" {
			return fmt.Errorf("Username required")
		}
	}
	return validPermissions(&r.Permissions)
}

func (r *sourceRoleRequest) ValidUpdate() error {
	if len(r.Name) > 254 {
		return fmt.Errorf("Username too long; must be less than 254 characters")
	}
	for _, user := range r.Users {
		if user.Name == "" {
			return fmt.Errorf("Username required")
		}
	}
	return validPermissions(&r.Permissions)
}

type roleResponse struct {
	Users       []*userResponse              `json:"users"`
	Name        string                       `json:"name"`
	Permissions chronograf.SourcePermissions `json:"permissions"`
	Links       selfLinks                    `json:"links"`
}

func newRoleResponse(srcID int, res *chronograf.SourceRole) roleResponse {
	su := make([]*userResponse, len(res.Users))
	for i := range res.Users {
		name := res.Users[i].Name
		su[i] = newUserResponse(srcID, name)
	}

	if res.Permissions == nil {
		res.Permissions = make(chronograf.SourcePermissions, 0)
	}
	return roleResponse{
		Name:        res.Name,
		Permissions: res.Permissions,
		Users:       su,
		Links:       newSelfLinks(srcID, "roles", res.Name),
	}
}
