package api

import (
	"log"
	"net/http"
)

func (s *Server) logAudit(userID int64, actor, action, entity string, entityID int64, r *http.Request) {
	ip := ""
	ua := ""

	if r != nil {
		ip = r.RemoteAddr
		ua = r.UserAgent()
	}

	log.Printf(
		"audit user_id=%d actor=%q action=%q entity=%q entity_id=%d ip=%q ua=%q",
		userID,
		actor,
		action,
		entity,
		entityID,
		ip,
		ua,
	)
}
