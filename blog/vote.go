package blog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type VoteStar struct {
	Id   int `json:"id"`
	Star int `json:"star"`
}

func voteHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := ValidateSession(w, r); err != nil {
		RespondError(w, err)
		return
	}

	v := &VoteStar{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(v); err != nil {
		fmt.Println(err)
		http.Error(w, fmt.Sprintf("failed to decode request %v", err), http.StatusBadRequest)
		return
	}

	err := v.save()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "success!")
}

func (v *VoteStar) save() error {

	fail := func(err error) error {
		return fmt.Errorf("save page failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shortDuration)
	defer cancel()

	s := "star" + strconv.Itoa(v.Star)

	update := `INSERT INTO poststatistics (postid,` + s + `) VALUES (?, 1)` +
		` ON DUPLICATE KEY UPDATE ` + s + `=` + s + `+1`
	if _, err := db.ExecContext(ctx, update, v.Id); err != nil {
		return fail(err)
	}

	return nil
}
