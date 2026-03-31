package httpadapter

import "net/http"

func currentUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok || userID <= 0 {
		return 0, false
	}
	return userID, true
}
